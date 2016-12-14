package ddcloud

import (
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyPortListNetworkDomainID = "networkdomain"
	resourceKeyPortListName            = "name"
	resourceKeyPortListDescription     = "description"
	resourceKeyPortListPort            = "port"
	resourceKeyPortListPorts           = "ports"
	resourceKeyPortListPortBegin       = "begin"
	resourceKeyPortListPortEnd         = "end"
	resourceKeyPortListChildIDs        = "child_lists"
)

func resourcePortList() *schema.Resource {
	return &schema.Resource{
		Exists: resourcePortListExists,
		Create: resourcePortListCreate,
		Read:   resourcePortListRead,
		Update: resourcePortListUpdate,
		Delete: resourcePortListDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyPortListNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the port list rule applies",
			},
			resourceKeyPortListName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "A name for the port list",
			},
			resourceKeyPortListDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the firewall rule",
			},
			resourceKeyPortListPort: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Ports included in the port list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						resourceKeyPortListPortBegin: &schema.Schema{
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The port (or starting port, for a port range)",
						},
						resourceKeyPortListPortEnd: &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "The ending port (for a port range",
						},
					},
				},
				ConflictsWith: []string{resourceKeyPortListPorts},
			},
			resourceKeyPortListPorts: &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Simple ports included in the port list",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				ConflictsWith: []string{resourceKeyPortListPort},
			},
			resourceKeyPortListChildIDs: &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The Ids of child port lists included in the port list",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
		},
	}
}

// Check if a port list resource exists.
func resourcePortListExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	portListID := data.Id()

	log.Printf("Check if port list '%s' exists.", portListID)

	client := provider.(*providerState).Client()

	portList, err := client.GetPortList(portListID)
	exists := (portList != nil)

	log.Printf("Port list '%s' exists: %t", portListID, true)

	return exists, err
}

// Create a port list resource.
func resourcePortListCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	networkDomainID := data.Get(resourceKeyPortListNetworkDomainID).(string)
	name := data.Get(resourceKeyPortListName).(string)
	description := data.Get(resourceKeyPortListDescription).(string)
	childListIDs := propertyHelper.GetStringSetItems(resourceKeyPortListChildIDs)

	var portListEntries []compute.PortListEntry
	if propertyHelper.HasProperty(resourceKeyPortListPorts) {
		// Port list entries from a simple set of ports.
		simplePorts := propertyHelper.GetIntSetItems(resourceKeyPortListPorts)
		for _, simplePort := range simplePorts {
			portListEntries = append(portListEntries, compute.PortListEntry{
				Begin: simplePort,
			})
		}
	} else { // Default for backward compatibility
		// Raw port list entries.
		portListEntries = propertyHelper.GetPortListPorts()
	}

	log.Printf("Create port list '%s' in network domain '%s'.", name, networkDomainID)

	client := provider.(*providerState).Client()
	portListID, err := client.CreatePortList(name, description, networkDomainID, portListEntries, childListIDs)
	if err != nil {
		return err
	}

	data.SetId(portListID)

	log.Printf("Successfully created port list '%s'.", portListID)

	return nil
}

// Read a port list resource.
func resourcePortListRead(data *schema.ResourceData, provider interface{}) error {
	portListID := data.Id()
	networkDomainID := data.Get(resourceKeyPortListNetworkDomainID).(string)

	log.Printf("Read port list '%s' in network domain '%s'.", portListID, networkDomainID)

	client := provider.(*providerState).Client()
	portList, err := client.GetPortList(portListID)
	if err != nil {
		return err
	}

	if portList == nil {
		log.Printf("Port list '%s' not found in network domain '%s' (will treat as deleted).", portListID, networkDomainID)

		data.SetId("") // Mark as deleted.
	}

	childListIDs := make([]string, len(portList.ChildLists))
	for index, childList := range portList.ChildLists {
		childListIDs[index] = childList.ID
	}

	propertyHelper := propertyHelper(data)
	data.Set(resourceKeyPortListDescription, portList.Description)
	propertyHelper.SetStringSetItems(resourceKeyPortListChildIDs, childListIDs)

	if propertyHelper.HasProperty(resourceKeyPortListPorts) {
		// Note that if the port list now has complex entries (rather than the simple ones configured), then we won't pick that up here.
		// TODO: Modify this logic to switch over to complex ports if resource state indicates it's necessary
		// For example, if portListEntry.End is populated, then we need to switch over to complex ports.

		var ports []int
		for _, portListEntry := range portList.Ports {
			ports = append(ports, portListEntry.Begin)
		}
		propertyHelper.SetIntSetItems(resourceKeyPortListPorts, ports)
	} else { // Default for backward compatibility
		propertyHelper.SetPortListPorts(portList.Ports)
	}

	return nil
}

// Update a port list resource.
func resourcePortListUpdate(data *schema.ResourceData, provider interface{}) error {
	portListID := data.Id()
	networkDomainID := data.Get(resourceKeyPortListNetworkDomainID).(string)

	log.Printf("Update port list '%s' in network domain '%s'.", portListID, networkDomainID)

	client := provider.(*providerState).Client()
	portList, err := client.GetPortList(portListID)
	if err != nil {
		return err
	}

	if portList == nil {
		log.Printf("Port list '%s' not found in network domain '%s' (will treat as deleted).", portListID, networkDomainID)

		data.SetId("") // Mark as deleted.

		return nil
	}

	propertyHelper := propertyHelper(data)

	editRequest := portList.BuildEditRequest()
	if data.HasChange(resourceKeyPortListDescription) {
		editRequest.Description = data.Get(resourceKeyPortListDescription).(string)
	}
	if data.HasChange(resourceKeyPortListPort) || data.HasChange(resourceKeyPortListPorts) {
		var portListEntries []compute.PortListEntry
		if propertyHelper.HasProperty(resourceKeyPortListPorts) {
			// Port list entries from a simple set of IP ports.
			simplePorts := propertyHelper.GetIntSetItems(resourceKeyPortListPorts)
			for _, simplePort := range simplePorts {
				portListEntries = append(portListEntries, compute.PortListEntry{
					Begin: simplePort,
				})
			}
		} else { // Default for backward compatibility
			// Raw port list entries.
			portListEntries = propertyHelper.GetPortListPorts()
		}

		editRequest.Ports = portListEntries
	}
	if data.HasChange(resourceKeyPortListChildIDs) {
		editRequest.ChildListIDs = propertyHelper.GetStringSetItems(resourceKeyPortListChildIDs)
	}

	err = client.EditPortList(portListID, editRequest)
	if err != nil {
		return err
	}

	log.Printf("Updated port list '%s'.", portListID)

	return nil
}

// Delete a port list resource.
func resourcePortListDelete(data *schema.ResourceData, provider interface{}) error {
	portListID := data.Id()
	networkDomainID := data.Get(resourceKeyPortListNetworkDomainID).(string)

	log.Printf("Delete port list '%s' in network domain '%s'.", portListID, networkDomainID)

	client := provider.(*providerState).Client()
	portList, err := client.GetPortList(portListID)
	if err != nil {
		return err
	}

	if portList == nil {
		log.Printf("Port list '%s' not found in network domain '%s' (will treat as deleted).", portListID, networkDomainID)

		return nil
	}

	err = client.DeletePortList(portListID)
	if err != nil {
		return err
	}

	log.Printf("Successfully deleted port list '%s' in network domain '%s'.", portListID, networkDomainID)

	return nil
}
