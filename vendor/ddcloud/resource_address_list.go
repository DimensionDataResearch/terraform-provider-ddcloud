package ddcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyAddressListNetworkDomainID   = "networkdomain"
	resourceKeyAddressListName              = "name"
	resourceKeyAddressListDescription       = "description"
	resourceKeyAddressListIPVersion         = "ip_version"
	resourceKeyAddressListAddresses         = "address"
	resourceKeyAddressListAddressBegin      = "begin"
	resourceKeyAddressListAddressEnd        = "end"
	resourceKeyAddressListAddressNetwork    = "network"
	resourceKeyAddressListAddressPrefixSize = "prefix_size"
	resourceKeyAddressListChildIDs          = "child_lists"
)

func resourceAddressList() *schema.Resource {
	return &schema.Resource{
		Exists: resourceAddressListExists,
		Create: resourceAddressListCreate,
		Read:   resourceAddressListRead,
		Update: resourceAddressListUpdate,
		Delete: resourceAddressListDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyAddressListNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the address list rule applies",
			},
			resourceKeyAddressListName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "A name for the address list",
			},
			resourceKeyAddressListDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the firewall rule",
			},
			resourceKeyAddressListIPVersion: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The IP version (IPv4 or IPv6) used by the address list",
			},
			resourceKeyAddressListAddresses: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Addresses included in the address list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						resourceKeyAddressListAddressBegin: &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The address or starting address for an address range",
						},
						resourceKeyAddressListAddressEnd: &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The end address for an address range",
						},
						resourceKeyAddressListAddressNetwork: &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The base address for an IP network",
							ConflictsWith: []string{
								resourceKeyAddressListAddresses + "." + resourceKeyAddressListAddressBegin,
								resourceKeyAddressListAddresses + "." + resourceKeyAddressListAddressEnd,
							},
						},
						resourceKeyAddressListAddressPrefixSize: &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The prefix size for an IP network",
							ConflictsWith: []string{
								resourceKeyAddressListAddresses + "." + resourceKeyAddressListAddressBegin,
								resourceKeyAddressListAddresses + "." + resourceKeyAddressListAddressEnd,
							},
						},
					},
				},
			},
			resourceKeyAddressListChildIDs: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Ids of child address lists included in the address list",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// Check if an address list resource exists.
func resourceAddressListExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	addressListID := data.Id()

	log.Printf("Check if address list '%s' exists.", addressListID)

	client := provider.(*providerState).Client()

	addressList, err := client.GetIPAddressList(addressListID)
	exists := (addressList != nil)

	log.Printf("Address list '%s' exists: %t", addressListID, true)

	return exists, err
}

// Create an address list resource.
func resourceAddressListCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)
	name := data.Get(resourceKeyAddressListName).(string)
	description := data.Get(resourceKeyAddressListDescription).(string)
	ipVersion := data.Get(resourceKeyAddressListIPVersion).(string)
	addresses := propertyHelper.GetAddressListAddresses()
	childListIDs := propertyHelper.GetStringListItems(resourceKeyAddressListChildIDs)

	log.Printf("Create address list '%s' in network domain '%s'.", name, networkDomainID)

	client := provider.(*providerState).Client()
	addressListID, err := client.CreateIPAddressList(name, description, ipVersion, networkDomainID, addresses, childListIDs)
	if err != nil {
		return err
	}

	data.SetId(addressListID)

	log.Printf("Successfully created address list '%s'.", addressListID)

	return nil
}

// Read an address list resource.
func resourceAddressListRead(data *schema.ResourceData, provider interface{}) error {
	addressListID := data.Id()
	networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)

	log.Printf("Read address list '%s' in network domain '%s'.", addressListID, networkDomainID)

	client := provider.(*providerState).Client()
	addressList, err := client.GetIPAddressList(addressListID)
	if err != nil {
		return err
	}

	if addressList == nil {
		log.Printf("Address list '%s' not found in network domain '%s' (will treat as deleted).", addressListID, networkDomainID)

		data.SetId("") // Mark as deleted.
	}

	childListIDs := make([]string, len(addressList.ChildLists))
	for index, childList := range addressList.ChildLists {
		childListIDs[index] = childList.ID
	}

	propertyHelper := propertyHelper(data)
	data.Set(resourceKeyAddressListDescription, addressList.Description)
	propertyHelper.SetAddressListAddresses(addressList.Addresses)
	propertyHelper.SetStringListItems(resourceKeyAddressListChildIDs, childListIDs)

	return nil
}

// Update an address list resource.
func resourceAddressListUpdate(data *schema.ResourceData, provider interface{}) error {
	addressListID := data.Id()
	networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)

	log.Printf("Update address list '%s' in network domain '%s'.", addressListID, networkDomainID)

	client := provider.(*providerState).Client()
	addressList, err := client.GetIPAddressList(addressListID)
	if err != nil {
		return err
	}

	if addressList == nil {
		log.Printf("Address list '%s' not found in network domain '%s' (will treat as deleted).", addressListID, networkDomainID)

		data.SetId("") // Mark as deleted.

		return nil
	}

	propertyHelper := propertyHelper(data)

	editRequest := addressList.BuildEditRequest()
	if data.HasChange(resourceKeyAddressListDescription) {
		editRequest.Description = data.Get(resourceKeyAddressListDescription).(string)
	}
	if data.HasChange(resourceKeyAddressListAddresses) {
		editRequest.Addresses = propertyHelper.GetAddressListAddresses()
	}
	if data.HasChange(resourceKeyAddressListChildIDs) {
		editRequest.ChildListIDs = propertyHelper.GetStringListItems(resourceKeyAddressListChildIDs)
	}

	err = client.EditIPAddressList(addressListID, editRequest)
	if err != nil {
		return err
	}

	log.Printf("Updated address list '%s'.", addressListID)

	return nil
}

// Delete an address list resource.
func resourceAddressListDelete(data *schema.ResourceData, provider interface{}) error {
	addressListID := data.Id()
	networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)

	log.Printf("Delete address list '%s' in network domain '%s'.", addressListID, networkDomainID)

	client := provider.(*providerState).Client()
	addressList, err := client.GetIPAddressList(addressListID)
	if err != nil {
		return err
	}

	if addressList == nil {
		log.Printf("Address list '%s' not found in network domain '%s' (will treat as deleted).", addressListID, networkDomainID)

		return nil
	}

	err = client.DeleteIPAddressList(addressListID)
	if err != nil {
		return err
	}

	log.Printf("Successfully deleted address list '%s' in network domain '%s'.", addressListID, networkDomainID)

	return nil
}
