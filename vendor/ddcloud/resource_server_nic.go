package ddcloud

import (
	"fmt"
	"log"

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

func addNetworkAdapterToServer(apiClient *compute.Client, serverID string, ipv4Address *string, vlanID *string, adapterType *string) error {
	// TODO: Implement

	return fmt.Errorf("addNetworkAdapterToServer is not yet implemented")
}

func removeNetworkAdapterFromServer(apiClient *compute.Client, serverID string, networkAdapterID string) error {
	// TODO: Implement

	return fmt.Errorf("addNetworkAdapterToServer is not yet implemented")
}

// updateNetworkAdapterIPAddress notifies the compute infrastructure that a NetworkAdapter's IP address has changed.
func updateNetworkAdapterIPAddress(apiClient *compute.Client, serverID string, networkAdapterID string, primaryIPv4 *string) error {
	log.Printf("Update IP address(es) for NetworkAdapter '%s'...", networkAdapterID)

	err := apiClient.NotifyServerIPAddressChange(networkAdapterID, primaryIPv4, nil)
	if err != nil {
		return err
	}

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, networkAdapterID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	return err
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
