package openstack

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/trunk_details"
	neutronports "github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type PortWithTrunkDetails struct {
	neutronports.Port
	trunk_details.TrunkDetailsExt
}

// getAttachedPorts returns a list of ports attached to a server.
func getAttachedPorts(client *gophercloud.ServiceClient, serverID string) ([]PortWithTrunkDetails, error) {
	listOpts := neutronports.ListOpts{
		DeviceID: serverID,
	}

	var ports []PortWithTrunkDetails

	allPages, err := neutronports.List(client, listOpts).AllPages()
	if err != nil {
		return ports, err
	}
	err = neutronports.ExtractPortsInto(allPages, &ports)
	if err != nil {
		return ports, err
	}

	return ports, nil
}
