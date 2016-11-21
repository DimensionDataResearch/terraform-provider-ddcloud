package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

// updateServerConfiguration reconfigures a server, changing the allocated RAM and / or CPU count.
func updateServerConfiguration(apiClient *compute.Client, server *compute.Server, memoryGB *int, cpuCount *int, cpuCoreCount *int, cpuSpeed *string) error {
	const noChange = "no change"

	memoryDescription := noChange
	if memoryGB != nil {
		memoryDescription = fmt.Sprintf("will change to %dGB", *memoryGB)
	}

	cpuCountDescription := noChange
	if cpuCount != nil {
		cpuCountDescription = fmt.Sprintf("will change to %d", *cpuCount)
	}

	cpuCoreCountDescription := noChange
	if cpuCoreCount != nil {
		cpuCoreCountDescription = fmt.Sprintf("will change to %d", *cpuCoreCount)
	}

	cpuSpeedDescription := noChange
	if cpuSpeed != nil {
		cpuSpeedDescription = fmt.Sprintf("will change to '%s'", *cpuSpeed)
	}

	log.Printf("Update configuration for server '%s' (memory: %s, CPU: %s, CPU cores per socket: %s, CPU speed: %s)...", server.ID, memoryDescription, cpuCountDescription, cpuCoreCountDescription, cpuSpeedDescription)

	err := apiClient.ReconfigureServer(server.ID, memoryGB, cpuCount, cpuCoreCount, cpuSpeed)
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForChange(compute.ResourceTypeServer, server.ID, "Reconfigure server", resourceUpdateTimeoutServer)

	return err
}

func captureServerNetworkConfiguration(server *compute.Server, data *schema.ResourceData, isPartial bool) {
	data.Set(resourceKeyServerPrimaryAdapterVLAN, *server.Network.PrimaryAdapter.VLANID)
	if isPartial {
		data.SetPartial(resourceKeyServerPrimaryAdapterVLAN)
	}

	data.Set(resourceKeyServerPrimaryAdapterIPv4, *server.Network.PrimaryAdapter.PrivateIPv4Address)
	if isPartial {
		data.SetPartial(resourceKeyServerPrimaryAdapterIPv4)
	}

	data.Set(resourceKeyServerPrimaryAdapterIPv6, *server.Network.PrimaryAdapter.PrivateIPv6Address)
	if isPartial {
		data.SetPartial(resourceKeyServerPrimaryAdapterIPv6)
	}

	data.Set(resourceKeyServerNetworkDomainID, server.Network.NetworkDomainID)
	if isPartial {
		data.SetPartial(resourceKeyServerNetworkDomainID)
	}
}

// updateServerIPAddress notifies the compute infrastructure that a server's IP address has changed.
func updateServerIPAddresses(apiClient *compute.Client, server *compute.Server, primaryIPv4 *string, primaryIPv6 *string) error {
	log.Printf("Update primary IP address(es) for server '%s'...", server.ID)

	primaryNetworkAdapterID := *server.Network.PrimaryAdapter.ID
	err := apiClient.NotifyServerIPAddressChange(primaryNetworkAdapterID, primaryIPv4, primaryIPv6)
	if err != nil {
		return err
	}

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", server.ID, primaryNetworkAdapterID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	return err
}
