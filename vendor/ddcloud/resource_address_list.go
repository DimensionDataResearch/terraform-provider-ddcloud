package ddcloud

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyAddressListNetworkDomainID   = "networkdomain"
	resourceKeyAddressListName              = "name"
	resourceKeyAddressListDescription       = "description"
	resourceKeyAddressListIPVersion         = "ip_version"
	resourceKeyAddressListAddresses         = "addresses"
	resourceKeyAddressListAddress           = "address"
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
			resourceKeyAddressListAddress: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Complex IP addresses or networks included in the address list",
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
								resourceKeyAddressListAddress + "." + resourceKeyAddressListAddressBegin,
								resourceKeyAddressListAddress + "." + resourceKeyAddressListAddressEnd,
							},
						},
						resourceKeyAddressListAddressPrefixSize: &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The prefix size for an IP network",
							ConflictsWith: []string{
								resourceKeyAddressListAddress + "." + resourceKeyAddressListAddressBegin,
								resourceKeyAddressListAddress + "." + resourceKeyAddressListAddressEnd,
							},
						},
					},
				},
				ConflictsWith: []string{resourceKeyAddressListAddresses},
			},
			resourceKeyAddressListAddresses: &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Simple IP addresses included in the address list",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:           schema.HashString,
				ConflictsWith: []string{resourceKeyAddressListAddress},
			},
			resourceKeyAddressListChildIDs: &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The Ids of child address lists included in the address list",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
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
	childListIDs := propertyHelper.GetStringSetItems(resourceKeyAddressListChildIDs)

	var addressListEntries []compute.IPAddressListEntry
	if propertyHelper.HasProperty(resourceKeyAddressListAddresses) {
		// Address list entries from a simple set of IP addresses.
		simpleAddresses := propertyHelper.GetStringSetItems(resourceKeyAddressListAddresses)
		for _, simpleAddress := range simpleAddresses {
			addressListEntries = append(addressListEntries, compute.IPAddressListEntry{
				Begin: simpleAddress,
			})
		}
	} else { // Default for backward compatibility
		// Raw address list entries.
		addressListEntries = propertyHelper.GetAddressListAddresses()
	}

	log.Printf("Create address list '%s' in network domain '%s'.", name, networkDomainID)

	client := provider.(*providerState).Client()
	addressListID, err := client.CreateIPAddressList(name, description, ipVersion, networkDomainID, addressListEntries, childListIDs)
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
	propertyHelper.SetStringSetItems(resourceKeyAddressListChildIDs, childListIDs)

	if propertyHelper.HasProperty(resourceKeyAddressListAddresses) {
		// Note that if the address list now has complex entries (rather than the simple ones configured), then we won't pick that up here.
		// TODO: Modify this logic to switch over to complex addresses if resource state indicates it's necessary
		// For example, if addressListEntry.End or addressListEntry.PrefixSize is populated, then we need to switch over to complex ports.

		var addresses []string
		for _, addressListEntry := range addressList.Addresses {
			addresses = append(addresses, addressListEntry.Begin)
		}
		propertyHelper.SetStringSetItems(resourceKeyAddressListAddresses, addresses)
	} else { // Default for backward compatibility
		propertyHelper.SetAddressListAddresses(addressList.Addresses)
	}

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
	if data.HasChange(resourceKeyAddressListAddress) || data.HasChange(resourceKeyAddressListAddresses) {
		var addressListEntries []compute.IPAddressListEntry
		if propertyHelper.HasProperty(resourceKeyAddressListAddresses) {
			// Address list entries from a simple set of IP addresses.
			simpleAddresses := propertyHelper.GetStringSetItems(resourceKeyAddressListAddresses)
			for _, simpleAddress := range simpleAddresses {
				addressListEntries = append(addressListEntries, compute.IPAddressListEntry{
					Begin: simpleAddress,
				})
			}
		} else { // Default for backward compatibility
			// Raw address list entries.
			addressListEntries = propertyHelper.GetAddressListAddresses()
		}

		editRequest.Addresses = addressListEntries
	}
	if data.HasChange(resourceKeyAddressListChildIDs) {
		editRequest.ChildListIDs = propertyHelper.GetStringSetItems(resourceKeyAddressListChildIDs)
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
