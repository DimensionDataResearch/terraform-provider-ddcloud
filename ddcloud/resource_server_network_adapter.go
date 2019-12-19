package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	resourceKeyServerPrimaryNetworkAdapter    = "primary_network_adapter"
	resourceKeyServerAdditionalNetworkAdapter = "additional_network_adapter"
	resourceKeyServerNetworkAdapterID         = "id"
	resourceKeyServerNetworkAdapterMAC        = "mac"
	resourceKeyServerNetworkAdapterVLANID     = "vlan"
	resourceKeyServerNetworkAdapterIPV4       = "ipv4"
	resourceKeyServerNetworkAdapterIPV6       = "ipv6"
	resourceKeyServerNetworkAdapterType       = "type"
)

func schemaServerPrimaryNetworkAdapter() *schema.Schema {
	return schemaServerNetworkAdapter(true)
}
func schemaServerAdditionalNetworkAdapter() *schema.Schema {
	return schemaServerNetworkAdapter(false)
}

func schemaServerNetworkAdapter(isPrimary bool) *schema.Schema {
	var maxItems int

	if isPrimary {
		maxItems = 1

	} else {
		maxItems = 0
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Required: isPrimary,
		Optional: !isPrimary,
		ForceNew: false,
		MinItems: 1,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerNetworkAdapterID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The network adapter's identifier in CloudControl",
				},
				resourceKeyServerNetworkAdapterMAC: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The network adapter's MAC address",
				},
				resourceKeyServerNetworkAdapterVLANID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Default:     nil,
					ForceNew:    false,
					Description: "VLAN ID of the network adapter",
				},
				resourceKeyServerNetworkAdapterIPV4: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Default:     nil,
					Description: "The IPV4 address associated with the network adapter",
				},
				resourceKeyServerNetworkAdapterIPV6: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Description: "The IPV6 Address associated the network adapter",
				},
				resourceKeyServerNetworkAdapterType: &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					Default:  nil,
					Description: fmt.Sprintf("The type of network adapter (%s or %s)",
						compute.NetworkAdapterTypeE1000,
						compute.NetworkAdapterTypeVMXNET3,
					),
					ValidateFunc: validateNetworkAdapterAdapterType,
				},
			},
		},
	}
}

func addServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
	log.Printf("Add network adapter to server '%s'", serverID)

	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Add network adapter to server '%s'", serverID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		var addAdapterError error
		if networkAdapter.HasExplicitType() {
			networkAdapter.ID, addAdapterError = apiClient.AddNicWithTypeToServer(
				serverID,
				networkAdapter.PrivateIPv4Address,
				networkAdapter.VLANID,
				networkAdapter.AdapterType,
			)
		} else {
			networkAdapter.ID, addAdapterError = apiClient.AddNicToServer(
				serverID,
				networkAdapter.PrivateIPv4Address,
				networkAdapter.VLANID,
			)
		}
		if compute.IsResourceBusyError(addAdapterError) {
			context.Retry()
		} else if addAdapterError != nil {
			context.Fail(addAdapterError)
		}
	})
	if err != nil {
		return err
	}

	log.Printf("Adding network adapter '%s' to server '%s'...", networkAdapter.ID, serverID)

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, networkAdapter.ID)
	_, err = apiClient.WaitForChange(
		compute.ResourceTypeNetworkAdapter,
		compositeNetworkAdapterID,
		"Add network adapter",
		resourceUpdateTimeoutServer,
	)
	if err != nil {
		return err
	}

	log.Printf("Added network adapter '%s' to server '%s'.", networkAdapter.ID, serverID)

	return nil
}

func modifyServerNetworkAdapterIP(providerState *providerState, serverID string, networkAdapter models.NetworkAdapter) error {
	log.Printf("Update IP address(es) for network adapter '%s'.", networkAdapter.ID)

	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Update IP address info for network adapter '%s'", networkAdapter.ID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()
		log.Printf("[DD] resource_server_network > modifyServerNetworkAdapterIP - Updating NIC id:%s ipv4:%s ipv6:%s",
			networkAdapter.ID, networkAdapter.PrivateIPv4Address, networkAdapter.PrivateIPv6Address)

		var changeAddressError error
		if len(networkAdapter.PrivateIPv6Address) > 0 {
			changeAddressError = apiClient.NotifyServerIPAddressChange(networkAdapter.ID, nil, &networkAdapter.PrivateIPv6Address)
		}
		if len(networkAdapter.PrivateIPv4Address) > 0 {
			changeAddressError = apiClient.NotifyServerIPAddressChange(networkAdapter.ID, &networkAdapter.PrivateIPv4Address, nil)
		}

		if compute.IsResourceBusyError(changeAddressError) {
			context.Retry()
		} else if changeAddressError != nil {
			context.Fail(changeAddressError)
		}
	})
	if err != nil {
		return err
	}

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, networkAdapter.ID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	log.Printf("[DD] Updated IP address(es) for network adapter:'%s' ipv4:'%s' ipv6:'%s'.",
		networkAdapter.ID, networkAdapter.PrivateIPv4Address, networkAdapter.PrivateIPv6Address)

	return err
}

func modifyServerNetworkAdapterType(providerState *providerState, serverID string, networkAdapter models.NetworkAdapter) error {
	log.Printf("[DD] Change type of network adapter '%s' to '%s'.",
		networkAdapter.ID,
		networkAdapter.AdapterType,
	)

	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Change type of network adapter '%s'", networkAdapter.ID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		changeTypeError := apiClient.ChangeNetworkAdapterType(networkAdapter.ID, networkAdapter.AdapterType)
		if compute.IsResourceBusyError(changeTypeError) {
			context.Retry()
		} else if changeTypeError != nil {
			context.Fail(changeTypeError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	log.Printf("Changing type of network adapter '%s'...", networkAdapter.ID)

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, networkAdapter.ID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Change type", resourceUpdateTimeoutServer)

	log.Printf("Changed type of network adapter '%s'.", networkAdapter.ID)

	return err
}

func removeServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
	log.Printf("Remove network adapter '%s'.", networkAdapter.ID)

	apiClient := providerState.Client()

	removingAdapter := true
	operationDescription := fmt.Sprintf("Remove network adapter '%s'", networkAdapter.ID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		removeError := apiClient.RemoveNicFromServer(networkAdapter.ID)
		if compute.IsResourceBusyError(removeError) {
			context.Retry()
		} else if compute.IsResourceNotFoundError(removeError) {
			log.Printf("Network adapter '%s' not found (will treat as deleted).",
				networkAdapter.ID,
			)
			removingAdapter = false
		} else if removeError != nil {
			context.Fail(removeError)
		}
	})
	if err != nil {
		return err
	}

	if removingAdapter {
		log.Printf("Removing network adapter '%s'...", networkAdapter.ID)

		compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, networkAdapter.ID)
		_, err = apiClient.WaitForNestedDeleteChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Remove network adapter", resourceUpdateTimeoutServer)
		if err != nil {
			return err
		}

		log.Printf("Removed network adapter '%s'.", networkAdapter.ID)
	}

	return nil
}
