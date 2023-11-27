package openstack

import (
	"fmt"
	"github.com/cuongpiger/k8s-ccm/pkg/metrics"
	"github.com/cuongpiger/k8s-ccm/pkg/util"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsecurity"
	neutronports "github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/api/core/v1"

	"github.com/cuongpiger/k8s-ccm/pkg/client"
	"github.com/cuongpiger/k8s-ccm/pkg/util/metadata"
	"github.com/spf13/pflag"
	"gopkg.in/gcfg.v1"
	"io"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

// userAgentData is used to add extra information to the gophercloud user-agent
var userAgentData []string

// supportedLBProvider map is used to define LoadBalancer providers that we support
var supportedLBProvider = []string{"amphora", "octavia", "ovn"}

// supportedContainerStore map is used to define supported tls-container-ref store
var supportedContainerStore = []string{"barbican", "external"}

const (
	// ProviderName is the name of the openstack provider
	ProviderName   = "openstack"
	defaultTimeOut = 60 * time.Second
)

// RouterOpts is used for Neutron routes
type RouterOpts struct {
	RouterID string `gcfg:"router-id"`
}

// OpenStack is an implementation of cloud provider Interface for OpenStack.
type OpenStack struct {
	provider       *gophercloud.ProviderClient
	epOpts         *gophercloud.EndpointOpts
	lbOpts         LoadBalancerOpts
	routeOpts      RouterOpts
	metadataOpts   metadata.Opts
	networkingOpts NetworkingOpts
	// InstanceID of the server where this OpenStack object is instantiated.
	localInstanceID       string
	kclient               kubernetes.Interface
	useV1Instances        bool // TODO: v1 instance apis can be deleted after the v2 is verified enough
	nodeInformer          coreinformers.NodeInformer
	nodeInformerHasSynced func() bool

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder
}

// NetworkingOpts is used for networking settings
type NetworkingOpts struct {
	IPv6SupportDisabled bool     `gcfg:"ipv6-support-disabled"`
	PublicNetworkName   []string `gcfg:"public-network-name"`
	InternalNetworkName []string `gcfg:"internal-network-name"`
	AddressSortOrder    string   `gcfg:"address-sort-order"`
}

// Config is used to read and store information from the cloud configuration file
type Config struct {
	Global            client.AuthOpts
	LoadBalancer      LoadBalancerOpts
	LoadBalancerClass map[string]*LBClass
	Route             RouterOpts
	Metadata          metadata.Opts
	Networking        NetworkingOpts
}

// LoadBalancerOpts have the options to talk to Neutron LBaaSV2 or Octavia
type LoadBalancerOpts struct {
	Enabled                        bool                `gcfg:"enabled"`              // if false, disables the controller
	LBVersion                      string              `gcfg:"lb-version"`           // overrides autodetection. Only support v2.
	SubnetID                       string              `gcfg:"subnet-id"`            // overrides autodetection.
	MemberSubnetID                 string              `gcfg:"member-subnet-id"`     // overrides autodetection.
	NetworkID                      string              `gcfg:"network-id"`           // If specified, will create virtual ip from a subnet in network which has available IP addresses
	FloatingNetworkID              string              `gcfg:"floating-network-id"`  // If specified, will create floating ip for loadbalancer, or do not create floating ip.
	FloatingSubnetID               string              `gcfg:"floating-subnet-id"`   // If specified, will create floating ip for loadbalancer in this particular floating pool subnetwork.
	FloatingSubnet                 string              `gcfg:"floating-subnet"`      // If specified, will create floating ip for loadbalancer in one of the matching floating pool subnetworks.
	FloatingSubnetTags             string              `gcfg:"floating-subnet-tags"` // If specified, will create floating ip for loadbalancer in one of the matching floating pool subnetworks.
	LBClasses                      map[string]*LBClass // Predefined named Floating networks and subnets
	LBMethod                       string              `gcfg:"lb-method"` // default to ROUND_ROBIN.
	LBProvider                     string              `gcfg:"lb-provider"`
	CreateMonitor                  bool                `gcfg:"create-monitor"`
	MonitorDelay                   util.MyDuration     `gcfg:"monitor-delay"`
	MonitorTimeout                 util.MyDuration     `gcfg:"monitor-timeout"`
	MonitorMaxRetries              uint                `gcfg:"monitor-max-retries"`
	MonitorMaxRetriesDown          uint                `gcfg:"monitor-max-retries-down"`
	ManageSecurityGroups           bool                `gcfg:"manage-security-groups"`
	InternalLB                     bool                `gcfg:"internal-lb"` // default false
	CascadeDelete                  bool                `gcfg:"cascade-delete"`
	FlavorID                       string              `gcfg:"flavor-id"`
	AvailabilityZone               string              `gcfg:"availability-zone"`
	EnableIngressHostname          bool                `gcfg:"enable-ingress-hostname"`            // Used with proxy protocol by adding a dns suffix to the load balancer IP address. Default false.
	IngressHostnameSuffix          string              `gcfg:"ingress-hostname-suffix"`            // Used with proxy protocol by adding a dns suffix to the load balancer IP address. Default nip.io.
	MaxSharedLB                    int                 `gcfg:"max-shared-lb"`                      //  Number of Services in maximum can share a single load balancer. Default 2
	ContainerStore                 string              `gcfg:"container-store"`                    // Used to specify the store of the tls-container-ref
	ProviderRequiresSerialAPICalls bool                `gcfg:"provider-requires-serial-api-calls"` // default false, the provider supportes the "bulk update" API call
	// revive:disable:var-naming
	TlsContainerRef string `gcfg:"default-tls-container-ref"` //  reference to a tls container
	// revive:enable:var-naming
}

// LoadBalancer is used for creating and maintaining load balancers
type LoadBalancer struct {
	secret        *gophercloud.ServiceClient
	network       *gophercloud.ServiceClient
	lb            *gophercloud.ServiceClient
	opts          LoadBalancerOpts
	kclient       kubernetes.Interface
	eventRecorder record.EventRecorder
}

// LBClass defines the corresponding floating network, floating subnet or internal subnet ID
type LBClass struct {
	FloatingNetworkID  string `gcfg:"floating-network-id,omitempty"`
	FloatingSubnetID   string `gcfg:"floating-subnet-id,omitempty"`
	FloatingSubnet     string `gcfg:"floating-subnet,omitempty"`
	FloatingSubnetTags string `gcfg:"floating-subnet-tags,omitempty"`
	NetworkID          string `gcfg:"network-id,omitempty"`
	SubnetID           string `gcfg:"subnet-id,omitempty"`
	MemberSubnetID     string `gcfg:"member-subnet-id,omitempty"`
}

type PortWithPortSecurity struct {
	neutronports.Port
	portsecurity.PortSecurityExt
}

// AddExtraFlags is called by the main package to add component specific command line flags
func AddExtraFlags(fs *pflag.FlagSet) {
	fs.StringArrayVar(&userAgentData, "user-agent", nil, "Extra data to add to gophercloud user-agent. Use multiple times to add more than one component.")
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider
func (os *OpenStack) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	clientset := clientBuilder.ClientOrDie("cloud-controller-manager")
	os.kclient = clientset
	os.eventBroadcaster = record.NewBroadcaster()
	os.eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: os.kclient.CoreV1().Events("")})
	os.eventRecorder = os.eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-provider-openstack"})
}

// LoadBalancer initializes a LbaasV2 object
func (os *OpenStack) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	klog.V(4).Info("openstack.LoadBalancer() called")
	if !os.lbOpts.Enabled {
		klog.V(4).Info("openstack.LoadBalancer() support for LoadBalancer controller is disabled")
		return nil, false
	}

	network, err := client.NewNetworkV2(os.provider, os.epOpts)
	if err != nil {
		klog.Errorf("Failed to create an OpenStack Network client: %v", err)
		return nil, false
	}

	lb, err := client.NewLoadBalancerV2(os.provider, os.epOpts)
	if err != nil {
		klog.Errorf("Failed to create an OpenStack LoadBalancer client: %v", err)
		return nil, false
	}

	// keymanager client is optional
	secret, err := client.NewKeyManagerV1(os.provider, os.epOpts)
	if err != nil {
		klog.Warningf("Failed to create an OpenStack Secret client: %v", err)
	}

	// LBaaS v1 is deprecated in the OpenStack Liberty release.
	// Currently kubernetes OpenStack cloud provider just support LBaaS v2.
	lbVersion := os.lbOpts.LBVersion
	if lbVersion != "" && lbVersion != "v2" {
		klog.Warningf("Config error: currently only support LBaaS v2, unrecognised lb-version \"%v\"", lbVersion)
		return nil, false
	}

	klog.V(1).Info("Claiming to support LoadBalancer")

	return &LbaasV2{LoadBalancer{secret, network, lb, os.lbOpts, os.kclient, os.eventRecorder}}, true
}

// ReadConfig reads values from the cloud.conf
func ReadConfig(config io.Reader) (Config, error) {
	if config == nil {
		return Config{}, fmt.Errorf("no OpenStack cloud provider config file given")
	}
	var cfg Config

	// Set default values explicitly
	cfg.LoadBalancer.Enabled = true
	cfg.LoadBalancer.InternalLB = false
	cfg.LoadBalancer.LBProvider = "amphora"
	cfg.LoadBalancer.LBMethod = "ROUND_ROBIN"
	cfg.LoadBalancer.CreateMonitor = false
	cfg.LoadBalancer.ManageSecurityGroups = false
	cfg.LoadBalancer.MonitorDelay = util.MyDuration{Duration: 5 * time.Second}
	cfg.LoadBalancer.MonitorTimeout = util.MyDuration{Duration: 3 * time.Second}
	cfg.LoadBalancer.MonitorMaxRetries = 1
	cfg.LoadBalancer.MonitorMaxRetriesDown = 3
	cfg.LoadBalancer.CascadeDelete = true
	cfg.LoadBalancer.EnableIngressHostname = false
	cfg.LoadBalancer.IngressHostnameSuffix = defaultProxyHostnameSuffix
	cfg.LoadBalancer.TlsContainerRef = ""
	cfg.LoadBalancer.ContainerStore = "barbican"
	cfg.LoadBalancer.MaxSharedLB = 2
	cfg.LoadBalancer.ProviderRequiresSerialAPICalls = false

	err := gcfg.FatalOnly(gcfg.ReadInto(&cfg, config))
	if err != nil {
		return Config{}, err
	}

	klog.V(5).Infof("Config, loaded from the config file:")
	client.LogCfg(cfg.Global)

	if cfg.Global.UseClouds {
		if cfg.Global.CloudsFile != "" {
			os.Setenv("OS_CLIENT_CONFIG_FILE", cfg.Global.CloudsFile)
		}
		err = client.ReadClouds(&cfg.Global)
		if err != nil {
			return Config{}, err
		}
		klog.V(5).Infof("Config, loaded from the %s:", cfg.Global.CloudsFile)
		client.LogCfg(cfg.Global)
	}
	// Set the default values for search order if not set
	if cfg.Metadata.SearchOrder == "" {
		cfg.Metadata.SearchOrder = fmt.Sprintf("%s,%s", metadata.ConfigDriveID, metadata.MetadataID)
	}

	if !util.Contains(supportedLBProvider, cfg.LoadBalancer.LBProvider) {
		klog.Warningf("Unsupported LoadBalancer Provider: %s", cfg.LoadBalancer.LBProvider)
	}

	if !util.Contains(supportedContainerStore, cfg.LoadBalancer.ContainerStore) {
		klog.Warningf("Unsupported Container Store: %s", cfg.LoadBalancer.ContainerStore)
	}

	return cfg, err
}

// NewOpenStack creates a new new instance of the openstack struct from a config struct
func NewOpenStack(cfg Config) (*OpenStack, error) {
	provider, err := client.NewOpenStackClient(&cfg.Global, "openstack-cloud-controller-manager", userAgentData...)
	if err != nil {
		return nil, err
	}

	if cfg.Metadata.RequestTimeout == (util.MyDuration{}) {
		cfg.Metadata.RequestTimeout.Duration = time.Duration(defaultTimeOut)
	}
	provider.HTTPClient.Timeout = cfg.Metadata.RequestTimeout.Duration

	useV1Instances := false
	v1instances := os.Getenv("OS_V1_INSTANCES")
	if strings.ToLower(v1instances) == "true" {
		useV1Instances = true
	}

	os := OpenStack{
		provider: provider,
		epOpts: &gophercloud.EndpointOpts{
			Region:       cfg.Global.Region,
			Availability: cfg.Global.EndpointType,
		},
		lbOpts:         cfg.LoadBalancer,
		routeOpts:      cfg.Route,
		metadataOpts:   cfg.Metadata,
		networkingOpts: cfg.Networking,
		useV1Instances: useV1Instances,
	}

	// ini file doesn't support maps so we are reusing top level sub sections
	// and copy the resulting map to corresponding loadbalancer section
	os.lbOpts.LBClasses = cfg.LoadBalancerClass

	err = checkOpenStackOpts(&os)
	if err != nil {
		return nil, err
	}

	return &os, nil
}

func init() {
	metrics.RegisterMetrics("occm")

	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := ReadConfig(config)
		if err != nil {
			klog.Warningf("failed to read config: %v", err)
			return nil, err
		}
		cloud, err := NewOpenStack(cfg)
		if err != nil {
			klog.Warningf("New openstack client created failed with config: %v", err)
		}
		return cloud, err
	})
}

// check opts for OpenStack
func checkOpenStackOpts(openstackOpts *OpenStack) error {
	return metadata.CheckMetadataSearchOrder(openstackOpts.metadataOpts.SearchOrder)
}
