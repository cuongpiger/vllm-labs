package client

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

// NewNetworkV2 creates a ServiceClient that may be used with the neutron v2 API
func NewNetworkV2(provider *gophercloud.ProviderClient, eo *gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	network, err := openstack.NewNetworkV2(provider, *eo)
	if err != nil {
		return nil, fmt.Errorf("failed to find network v2 %s endpoint for region %s: %v", eo.Availability, eo.Region, err)
	}
	return network, nil
}

// NewComputeV2 creates a ServiceClient that may be used with the nova v2 API
func NewComputeV2(provider *gophercloud.ProviderClient, eo *gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	compute, err := openstack.NewComputeV2(provider, *eo)
	if err != nil {
		return nil, fmt.Errorf("failed to find compute v2 %s endpoint for region %s: %v", eo.Availability, eo.Region, err)
	}
	return compute, nil
}

// NewBlockStorageV3 creates a ServiceClient that may be used with the Cinder v3 API
func NewBlockStorageV3(provider *gophercloud.ProviderClient, eo *gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	storage, err := openstack.NewBlockStorageV3(provider, *eo)
	if err != nil {
		return nil, fmt.Errorf("unable to find cinder v3 %s endpoint for region %s: %v", eo.Availability, eo.Region, err)
	}
	return storage, nil
}

// NewLoadBalancerV2 creates a ServiceClient that may be used with the Neutron LBaaS v2 API
func NewLoadBalancerV2(provider *gophercloud.ProviderClient, eo *gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	var lb *gophercloud.ServiceClient
	var err error
	lb, err = openstack.NewLoadBalancerV2(provider, *eo)
	if err != nil {
		return nil, fmt.Errorf("failed to find load-balancer v2 %s endpoint for region %s: %v", eo.Availability, eo.Region, err)
	}
	return lb, nil
}

// NewKeyManagerV1 creates a ServiceClient that can be used with KeyManager v1 API
func NewKeyManagerV1(provider *gophercloud.ProviderClient, eo *gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error) {
	secret, err := openstack.NewKeyManagerV1(provider, *eo)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize keymanager client for region %s: %v", eo.Region, err)
	}
	return secret, nil
}
