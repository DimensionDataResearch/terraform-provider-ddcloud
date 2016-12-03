package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerNetworkAdapter       = "network_adapter"
	resourceKeyServerNetworkAdapterID     = "id"
	resourceKeyServerNetworkAdapterMAC    = "mac"
	resourceKeyServerNetworkAdapterVLANID = "vlan"
	resourceKeyServerNetworkAdapterIPV4   = "ipv4"
	resourceKeyServerNetworkAdapterIPV6   = "ipv6"
	resourceKeyServerNetworkAdapterType   = "type"
)

func schemaServerNetworkAdapter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
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
					Description: "VLAN ID of the network adapter",
				},
				resourceKeyServerNetworkAdapterIPV4: &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Default:     nil,
					Description: "The IPV4 address associated with the network adapter",
				},
				resourceKeyServerNetworkAdapterIPV6: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
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

	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Add network adapter to server '%s'", serverID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
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

func modifyServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
	log.Printf("Update IP address(es) for network adapter '%s'.", networkAdapter.ID)

	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Update IP address info for network adapter '%s'", networkAdapter.ID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		changeAddressError := apiClient.NotifyServerIPAddressChange(networkAdapter.ID, &networkAdapter.PrivateIPv4Address, nil)
		if compute.IsResourceBusyError(changeAddressError) {
			context.Retry()
		} else if changeAddressError != nil {
			context.Fail(changeAddressError)
		}
	})
	if err != nil {
		return err
	}

	log.Printf("Updating IP address(es) for network adapter '%s'...", networkAdapter.ID)

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, networkAdapter.ID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	log.Printf("Updated IP address(es) for network adapter '%s'.", networkAdapter.ID)

	return err
}

func removeServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
	log.Printf("Remove network adapter '%s'.", networkAdapter.ID)

	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	removingAdapter := true
	operationDescription := fmt.Sprintf("Remove network adapter '%s'", networkAdapter.ID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
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

// Validate that the specified value represents a valid network adapter type.
func validateNetworkAdapterAdapterType(value interface{}, propertyName string) (messages []string, errors []error) {
	if value == nil {
		return
	}

	adapterType, ok := value.(string)
	if !ok {
		errors = append(errors,
			fmt.Errorf("Unexpected value type '%v'", value),
		)

		return
	}

	switch adapterType {
	case compute.NetworkAdapterTypeE1000:
	case compute.NetworkAdapterTypeVMXNET3:
		break
	default:
		errors = append(errors,
			fmt.Errorf("Invalid network adapter type '%s'", value),
		)
	}

	return
}
