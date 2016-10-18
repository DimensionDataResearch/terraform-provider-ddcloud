package ddcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyPortListNetworkDomainID = "networkdomain"
	resourceKeyPortListName            = "name"
	resourceKeyPortListDescription     = "description"
	resourceKeyPortListPorts           = "port"
	resourceKeyPortListPortBegin       = "begin"
	resourceKeyPortListPortEnd         = "end"
	resourceKeyPortListChildListIDs    = "child_lists"
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
			resourceKeyPortListPorts: &schema.Schema{
				Type:        schema.TypeMap,
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
							Default:     nil,
							Description: "The ending port (for a port range",
						},
					},
				},
			},
			resourceKeyPortListChildListIDs: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Ids of child port lists included in the port list",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
	ports := propertyHelper.GetPortListPorts()
	childListIDs := data.Get(resourceKeyPortListChildListIDs).([]string)

	log.Printf("Create port list '%s' in network domain '%s'.", name, networkDomainID)

	client := provider.(*providerState).Client()
	portListID, err := client.CreatePortList(name, description, networkDomainID, ports, childListIDs)
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

	editRequest := portList.BuildEditRequest()
	if data.HasChange(resourceKeyPortListDescription) {
		editRequest.Description = data.Get(resourceKeyPortListDescription).(string)
	}
	if data.HasChange(resourceKeyPortListPorts) {
		editRequest.Ports = propertyHelper(data).GetPortListPorts()
	}
	if data.HasChange(resourceKeyPortListChildListIDs) {
		editRequest.ChildListIDs = data.Get(resourceKeyPortListChildListIDs).([]string)
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
