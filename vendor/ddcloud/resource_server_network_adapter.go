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
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
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
					Optional:    true,
					Default:     0,
					Description: "A unique identifier for the network adapter (0 is the primary adapter)",
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
		Set: hashServerNetworkAdapter,
		ConflictsWith: []string{
			resourceKeyServerPrimaryAdapterVLAN,
			resourceKeyServerPrimaryAdapterIPv4,
			resourceKeyServerPrimaryAdapterType,
		},
	}
}

// TODO: Define MapStructure-compatible structures to represent configured network adapters.
// TODO: Give these structures methods for reading / writing VirtualMachineNetworkAdapter.

func addServerNetworkAdapter(providerState *providerState, serverID string, networkAdapter *models.NetworkAdapter) error {
	// TODO: Implement (remember to use providerState.AcquireAsyncOperationLock)

	return fmt.Errorf("addServerNetworkAdapter is not yet implemented")
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
