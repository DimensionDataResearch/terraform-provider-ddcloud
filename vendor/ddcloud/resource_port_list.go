package ddcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyPortListNetworkDomainID = "networkdomain"
	resourceKeyPortListName            = "name"
	resourceKeyPortListDescription     = "description"
	resourceKeyPortListPorts           = "ports"
	resourceKeyPortListChildPortLists  = "child_port_lists"
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
				Description: "A description for the firewall rule",
			},
			resourceKeyPortListPorts: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The ports included in the port list",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			resourceKeyPortListChildPortLists: &schema.Schema{
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
func resourcePortListExists(data *schema.ResourceData, provider interface{}) (exists bool, err error) {
	return false, nil
}

// Create a port list resource.
func resourcePortListCreate(data *schema.ResourceData, provider interface{}) error {
	return nil
}

// Read a port list resource.
func resourcePortListRead(data *schema.ResourceData, provider interface{}) error {
	return nil
}

// Update a port list resource.
func resourcePortListUpdate(data *schema.ResourceData, provider interface{}) error {
	return nil
}

// Delete a port list resource.
func resourcePortListDelete(data *schema.ResourceData, provider interface{}) error {
	return nil
}
