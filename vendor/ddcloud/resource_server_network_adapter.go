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
	resourceKeyServerNetworkAdapterIndex  = "index"
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
				resourceKeyServerNetworkAdapterIndex: &schema.Schema{
					Type:        schema.TypeInt,
					Computed:    true,
					Default:     nil,
					Description: "The index of the network adapter in CloudControl (0 is the primary adapter)",
				},
				resourceKeyServerNetworkAdapterVLANID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Default:     nil,
					Description: "VLAN ID of the network adapter",
					ForceNew:    true,
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
					ForceNew: true,
					Default:  compute.NetworkAdapterTypeE1000,
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

// AF: This is unnecessary - we do this in the initial deployment configuration.
func createNetworkAdapters(server *compute.Server, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Id()

	configuredNetworkAdapters := propertyHelper.GetNetworkAdapters()
	actualNetworkAdapters := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network)

	addNetworkAdapters, _, _ := configuredNetworkAdapters.SplitByAction(actualNetworkAdapters)
	if addNetworkAdapters.IsEmpty() {
		log.Printf("No post-deploy changes required for network adapters of server '%s'.", serverID)

		return nil
	}

	return nil
}

func addServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
	log.Printf("Add network adapter with index %d to server '%s'",
		networkAdapter.Index,
		serverID,
	)

	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Add network adapter to server '%s'", serverID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		var addAdapterError error
		if networkAdapter.HasExplicitType() {
			networkAdapter.ID, addAdapterError = apiClient.AddNicWithTypeToServer(
				networkAdapter.ID,
				networkAdapter.PrivateIPv4Address,
				networkAdapter.VLANID,
				networkAdapter.AdapterType,
			)
		} else {
			networkAdapter.ID, addAdapterError = apiClient.AddNicToServer(
				networkAdapter.ID,
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
	resource, err := apiClient.WaitForChange(
		compute.ResourceTypeNetworkAdapter,
		compositeNetworkAdapterID,
		"Add network adapter",
		resourceUpdateTimeoutServer,
	)
	if err != nil {
		return err
	}

	server := resource.(*compute.Server)
	for index, serverNetworkAdapter := range server.Network.AdditionalNetworkAdapters {
		if *serverNetworkAdapter.ID == networkAdapter.ID {
			networkAdapter.Index = index

			break
		}
	}
	if networkAdapter.Index == 0 {
		return fmt.Errorf("Unable to find network adapter '%s' in server '%s'",
			networkAdapter.ID,
			serverID,
		)
	}

	log.Printf("Added network adapter '%s' to server '%s' at index %d.", networkAdapter.ID, serverID, networkAdapter.Index)

	return nil
}

func updateServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
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
	return fmt.Errorf("removeServerNetworkAdapter is not yet implemented")
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

// Create a hash code to represent the property values for a server network adapter.
func hashServerNetworkAdapter(value interface{}) int {
	adapterProperties, ok := value.(map[string]interface{})
	if !ok {
		return -1
	}

	var (
		index       int
		vlanID      string
		ipv4Address string
		adapterType string
	)
	adapterProperty, ok := adapterProperties[resourceKeyServerNetworkAdapterIndex]
	if ok {
		index = adapterProperty.(int)
	}

	adapterProperty, ok = adapterProperties[resourceKeyServerNetworkAdapterVLANID]
	if ok {
		vlanID = adapterProperty.(string)
	}

	adapterProperty, ok = adapterProperties[resourceKeyServerNetworkAdapterIPV4]
	if ok {
		ipv4Address = adapterProperty.(string)
	}

	adapterProperty, ok = adapterProperties[resourceKeyServerNetworkAdapterType]
	if ok {
		adapterType = adapterProperty.(string)
	}

	return schema.HashString(fmt.Sprintf(
		"%d|%s|%s|%s", index, vlanID, ipv4Address, adapterType,
	))
}
