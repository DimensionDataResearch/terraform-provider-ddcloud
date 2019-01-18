package ddcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyAddress           = "address"
	resourceKeyAddressBegin      = "begin"
	resourceKeyAddressEnd        = "end"
	resourceKeyAddressNetwork    = "network"
	resourceKeyAddressPrefixSize = "prefix_size"
)

func resourceAddress() *schema.Resource {
	return &schema.Resource{
		Exists: resourceAddressExists,
		Create: resourceAddressCreate,
		Read:   resourceAddressRead,
		Update: resourceAddressUpdate,
		Delete: resourceAddressDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyAddressBegin: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The address or starting address for an address range",
			},
			resourceKeyAddressEnd: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The end address for an address range",
			},
			resourceKeyAddressNetwork: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The base address for an IP network",
				ConflictsWith: []string{
					resourceKeyAddress + "." + resourceKeyAddressBegin,
					resourceKeyAddress + "." + resourceKeyAddressEnd,
				},
			},
			resourceKeyAddressPrefixSize: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The prefix size for an IP network",
				ConflictsWith: []string{
					resourceKeyAddress + "." + resourceKeyAddressBegin,
					resourceKeyAddress + "." + resourceKeyAddressEnd,
				},
			},
		},
	}
}

// Check if an address list resource exists.
func resourceAddressExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	log.Printf("XXXX Inside resourceAddressExists")
	//addressListID := data.Id()
	//
	//log.Printf("Check if address list '%s' exists.", addressListID)
	//
	//client := provider.(*providerState).Client()
	//
	//addressList, err := client.GetIPAddressList(addressListID)
	//exists := (addressList != nil)
	//
	//log.Printf("Address list '%s' exists: %t", addressListID, true)
	//
	//return exists, err
	return false, nil
}

// Create an address list resource.
func resourceAddressCreate(data *schema.ResourceData, provider interface{}) error {
	log.Printf("XXXX Inside resourceAddressCreate")


	//propertyHelper := propertyHelper(data)
	//
	//networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)
	//name := data.Get(resourceKeyAddressListName).(string)
	//description := data.Get(resourceKeyAddressListDescription).(string)
	//ipVersion := data.Get(resourceKeyAddressListIPVersion).(string)
	//childListIDs := propertyHelper.GetStringSetItems(resourceKeyAddressListChildIDs)
	//
	//var addressListEntries []compute.IPAddressListEntry
	//if propertyHelper.HasProperty(resourceKeyAddressListAddresses) {
	//	// Address list entries from a simple set of IP addresses.
	//	simpleAddresses := propertyHelper.GetStringSetItems(resourceKeyAddressListAddresses)
	//	for _, simpleAddress := range simpleAddresses {
	//		addressListEntries = append(addressListEntries, compute.IPAddressListEntry{
	//			Begin: simpleAddress,
	//		})
	//	}
	//} else { // Default for backward compatibility
	//	// Raw address list entries.
	//	addressListEntries = propertyHelper.GetAddressListAddresses()
	//}
	//
	//log.Printf("Create address list '%s' in network domain '%s'.", name, networkDomainID)
	//
	//client := provider.(*providerState).Client()
	//addressListID, err := client.CreateIPAddressList(name, description, ipVersion, networkDomainID, addressListEntries, childListIDs)
	//if err != nil {
	//	return err
	//}
	//
	//data.SetId(addressListID)
	//
	//log.Printf("Successfully created address list '%s'.", addressListID)
	//
	return nil
}

// Read an address resource.
func resourceAddressRead(data *schema.ResourceData, provider interface{}) error {
	//log.Printf("XXXX Inside resourceAddressListRead")
	//addressListID := data.Id()
	//log.Printf("XXXX get data domainID")
	//networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)
	//log.Printf("XXXX get data domainID Done")
	//log.Printf("Read address list '%s' in network domain '%s'.", addressListID, networkDomainID)
	//log.Printf("XXXX Inside resourceAddressListRead before provider")
	//client := provider.(*providerState).Client()
	//log.Printf("XXXX Inside resourceAddressListRead after provider")
	//addressList, err := client.GetIPAddressList(addressListID)
	//if err != nil {
	//	log.Printf("XXXX error %s", err)
	//	return err
	//}
	//
	//if addressList == nil {
	//	log.Printf("Address list '%s' not found in network domain '%s' (will treat as deleted).", addressListID, networkDomainID)
	//
	//	data.SetId("") // Mark as deleted.
	//}
	//
	//log.Printf("XXXX Inside resourceAddressListRead - check 2")
	//childListIDs := make([]string, len(addressList.ChildLists))
	//for index, childList := range addressList.ChildLists {
	//	childListIDs[index] = childList.ID
	//	log.Printf("XXXX Inside resourceAddressListRead - check 3")
	//}
	//
	//propertyHelper := propertyHelper(data)
	//data.Set(resourceKeyAddressListDescription, addressList.Description)
	//propertyHelper.SetStringSetItems(resourceKeyAddressListChildIDs, childListIDs)
	//
	//if propertyHelper.HasProperty(resourceKeyAddressListAddresses) {
	//	log.Printf("XXXX Inside HasProperty(resourceKeyAddressListAddresses)")
	//
	//	// Note that if the address list now has complex entries (rather than the simple ones configured), then we won't pick that up here.
	//	// TODO: Modify this logic to switch over to complex addresses if resource state indicates it's necessary
	//	// For example, if addressListEntry.End or addressListEntry.PrefixSize is populated, then we need to switch over to complex ports.
	//
	//	var addresses []string
	//	for _, addressListEntry := range addressList.Addresses {
	//		addresses = append(addresses, addressListEntry.Begin)
	//	}
	//	propertyHelper.SetStringSetItems(resourceKeyAddressListAddresses, addresses)
	//} else { // Default for backward compatibility
	//	log.Printf("XXXX Inside NO HasProperty(resourceKeyAddressListAddresses)")
	//	propertyHelper.SetAddressListAddresses(addressList.Addresses)
	//}

	//TODO: Implement set resourcekey address. At the moment changes in address will not get detected for change
	return nil
}

// Update an address list resource.
func resourceAddressUpdate(data *schema.ResourceData, provider interface{}) error {
	//log.Printf("XXXX Inside resourceAddressListUpdate")
	//addressListID := data.Id()
	//networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)
	//
	//log.Printf("Update address list '%s' in network domain '%s'.", addressListID, networkDomainID)
	//
	//client := provider.(*providerState).Client()
	//addressList, err := client.GetIPAddressList(addressListID)
	//if err != nil {
	//	return err
	//}
	//
	//if addressList == nil {
	//	log.Printf("Address list '%s' not found in network domain '%s' (will treat as deleted).", addressListID, networkDomainID)
	//
	//	data.SetId("") // Mark as deleted.
	//
	//	return nil
	//}
	//
	//propertyHelper := propertyHelper(data)
	//
	//editRequest := addressList.BuildEditRequest()
	//if data.HasChange(resourceKeyAddressListDescription) {
	//	editRequest.Description = data.Get(resourceKeyAddressListDescription).(string)
	//}
	//if data.HasChange(resourceKeyAddressListAddress) || data.HasChange(resourceKeyAddressListAddresses) {
	//	var addressListEntries []compute.IPAddressListEntry
	//	if propertyHelper.HasProperty(resourceKeyAddressListAddresses) {
	//		// Address list entries from a simple set of IP addresses.
	//		simpleAddresses := propertyHelper.GetStringSetItems(resourceKeyAddressListAddresses)
	//		for _, simpleAddress := range simpleAddresses {
	//			addressListEntries = append(addressListEntries, compute.IPAddressListEntry{
	//				Begin: simpleAddress,
	//			})
	//		}
	//	} else { // Default for backward compatibility
	//		// Raw address list entries.
	//		addressListEntries = propertyHelper.GetAddressListAddresses()
	//	}
	//
	//	editRequest.Addresses = addressListEntries
	//}
	//if data.HasChange(resourceKeyAddressListChildIDs) {
	//	editRequest.ChildListIDs = propertyHelper.GetStringSetItems(resourceKeyAddressListChildIDs)
	//}
	//
	//err = client.EditIPAddressList(editRequest)
	//if err != nil {
	//	return err
	//}
	//
	//log.Printf("Updated address list '%s'.", addressListID)

	return nil
}

// Delete an address list resource.
func resourceAddressDelete(data *schema.ResourceData, provider interface{}) error {
	//log.Printf("XXXX Inside resourceAddressListDelete")
	//addressListID := data.Id()
	//networkDomainID := data.Get(resourceKeyAddressListNetworkDomainID).(string)
	//
	//log.Printf("Delete address list '%s' in network domain '%s'.", addressListID, networkDomainID)
	//
	//client := provider.(*providerState).Client()
	//addressList, err := client.GetIPAddressList(addressListID)
	//if err != nil {
	//	return err
	//}
	//
	//if addressList == nil {
	//	log.Printf("Address list '%s' not found in network domain '%s' (will treat as deleted).", addressListID, networkDomainID)
	//
	//	return nil
	//}
	//
	//err = client.DeleteIPAddressList(addressListID)
	//if err != nil {
	//	return err
	//}
	//
	//log.Printf("Successfully deleted address list '%s' in network domain '%s'.", addressListID, networkDomainID)

	return nil
}
