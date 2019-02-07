package ddcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func dataSourceAddressList() *schema.Resource {
	return &schema.Resource{

		Read: dataSourceAddressListRead,

		Schema: map[string]*schema.Schema{
			resourceKeyAddressListNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Id of the network domain in which the address list rule applies",
			},
			resourceKeyAddressListName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the address list",
			},
			resourceKeyAddressListDescription: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description for the firewall rule",
			},
			resourceKeyAddressListIPVersion: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP version (IPv4 or IPv6) used by the address list",
				// ConflictsWith: []string{resourceKeyAddressListIPVersion},
			},
		},
	}
}

// Read a address list data source.
func dataSourceAddressListRead(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyAddressListName).(string)
	domainId := data.Get(resourceKeyAddressListNetworkDomainID).(string)
	log.Printf("Read address list '%s' from network domain '%s'.", name, domainId)

	apiClient := provider.(*providerState).Client()

	addressList, err := apiClient.GetIPAddressListByName(name, domainId)

	if err != nil {
		return err
	}

	if addressList != nil {
		log.Printf("Found addresslist '%s'.", name)

		data.SetId(addressList.ID)
		data.Set(resourceKeyAddressListName, addressList.Name)

		data.Set(resourceKeyAddressListDescription, addressList.Description)
		data.Set(resourceKeyAddressListIPVersion, addressList.IPVersion)

	} else {
		return fmt.Errorf("failed to find Addresslist with name '%s'", name)
	}

	return nil
}
