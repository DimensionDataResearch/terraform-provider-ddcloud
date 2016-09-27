package ddcloud

import "github.com/hashicorp/terraform/helper/schema"

const (
	resourceKeyNicServerID     = "serverid"
	resourceKeyNicID           = "id"
	resourceKeyNicVLANID       = "vlan_id"
	resourceKeyNicPriavateIPV4 = "private_ipv4"
	resourceKeyNicPrivateIPV6  = "private_ipv6"
	resourceKeyIsShutdownOk    = "shutdown_ok"
)

func resourceAdditionalNic() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdditionalNicCreate,
		Read:   resourceAdditionalNicRead,
		Update: resourceAdditionalNicUpdate,
		Delete: resourceAdditionalNicDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNicServerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the server to which the additional nics needs to be updated",
			},
			resourceKeyNicID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the nic",
			},
			resourceKeyNicVLANID: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VLAN ID of the nic",
			},
			resourceKeyNicPriavateIPV4: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Private IPV4 address for the nic",
			},
			resourceKeyNicPrivateIPV6: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Private IPV6 Address for the nic",
			},
			resourceKeyIsShutdownOk: &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Server needs to be shutdown to do any modifications for nic",
			},
		},
	}

}

func resourceAdditionalNicCreate(data *schema.ResourceData, provider interface{}) error {
	var err error
	return err
}

func resourceAdditionalNicRead(data *schema.ResourceData, provider interface{}) error {
	var err error
	return err
}

func resourceAdditionalNicUpdate(data *schema.ResourceData, provider interface{}) error {
	var err error
	return err
}

func resourceAdditionalNicDelete(data *schema.ResourceData, provider interface{}) error {
	var err error
	return err
}
