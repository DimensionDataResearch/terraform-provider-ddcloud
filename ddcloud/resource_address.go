package ddcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"log"
	"strings"
)

const (
	resourceKeyAddressBegin           = "begin"
	resourceKeyAddressEnd             = "end"
	resourceKeyAddressNetwork         = "network"
	resourceKeyAddressPrefixSize      = "prefix_size"
	resourceKeyAddressNetworkDomainID = "networkdomain"
	resourceKeyAddressListID          = "addresslist_id"
)

func resourceAddress() *schema.Resource {
	return &schema.Resource{
		Exists: resourceAddressExists,
		Create: resourceAddressCreate,
		Read:   resourceAddressRead,
		Update: resourceAddressUpdate,
		Delete: resourceAddressDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyAddressNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the address applies",
			},
			resourceKeyAddressListID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the address list",
			},
			resourceKeyAddressBegin: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The address or starting address for an address range",
			},
			resourceKeyAddressEnd: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The end address for an address range",
			},
			resourceKeyAddressNetwork: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The base address for an IP network",
			},
			resourceKeyAddressPrefixSize: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The prefix size for an IP network",
			},
		},
	}
}

// Check if an address list resource exists.
func resourceAddressExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	log.Printf("resourceAddressExists")
	addressListId := data.Get(resourceKeyAddressListID).(string)

	client := provider.(*providerState).Client()

	begin, okBegin := data.GetOk(resourceKeyAddressBegin)
	network, okNetwork := data.GetOk(resourceKeyAddressNetwork)

	// Value been deleted in dd cloud manually
	if okBegin || okNetwork {
		if !client.CheckAddressExists(addressListId, begin.(string), network.(string)) {
			data.SetId("") // Mark as deleted.
			return false, nil
		}
	}

	return true, nil
}

// Create an address list resource.
func resourceAddressCreate(data *schema.ResourceData, provider interface{}) error {

	log.Printf("resourceAddressCreate")

	var begin string
	var end string
	var network string
	var prefixSize int

	addressListId := data.Get(resourceKeyAddressListID).(string)

	valBegin, beginOk := data.GetOk(resourceKeyAddressBegin)
	if beginOk {
		begin = valBegin.(string)
	}

	valEnd, endOk := data.GetOk(resourceKeyAddressEnd)
	if endOk {
		end = valEnd.(string)
	}

	valNetwork, networkOk := data.GetOk(resourceKeyAddressNetwork)
	if networkOk {
		// networkIp := valNetwork.(string)
		network = valNetwork.(string)
	}

	valPrefix, prefixOk := data.GetOkExists(resourceKeyAddressPrefixSize)
	if prefixOk {
		prefixSize = valPrefix.(int)
	}

	// Validation when ip address begin is used, only address type single IP or IP range is defined.
	if beginOk {
		if networkOk || prefixOk {
			err := errors.New("INPUT ERROR: You must define one type of Address ONLY; Either single IP (begin) " +
				"or IP range (begin and end) or Subnet (begin and prefix_size)")
			return err
		}
	}

	client := provider.(*providerState).Client()
	_, err := client.AddAddress(addressListId, begin, end, network, prefixSize)

	if err != nil {
		return err
	}

	var ip string
	if begin != "" {
		ip = strings.Replace(begin, ":", "", -1)
	} else {
		ip = strings.Replace(network, ":", "", -1)
	}

	data.SetId(ip)

	return resourceAddressRead(data, provider)
}

// Read an address resource.
func resourceAddressRead(data *schema.ResourceData, provider interface{}) error {
	log.Printf("resourceAddressRead")

	addressListId := data.Get(resourceKeyAddressListID).(string)
	begin, _ := data.GetOk(resourceKeyAddressBegin)
	network, _ := data.GetOk(resourceKeyAddressNetwork)

	client := provider.(*providerState).Client()

	// Check if address exists in cloud
	addr, addrOk := client.GetAddressOk(addressListId, begin.(string), network.(string))
	if !addrOk {
		// Address has been deleted in cloud
		// data.SetId("")
		return nil
	}

	// Type subnet
	if addr.PrefixSize != nil {
		data.Set(resourceKeyAddressNetwork, addr.Begin)
		data.Set(resourceKeyAddressPrefixSize, addr.PrefixSize)

	} else {
		// Type IP
		data.Set(resourceKeyAddressBegin, addr.Begin)

		// Type IP range
		if addr.End != nil {
			data.Set(resourceKeyAddressEnd, addr.End)
		}
	}

	return nil
}

// Update an address list resource.
func resourceAddressUpdate(data *schema.ResourceData, provider interface{}) error {

	log.Printf("resourceAddressUpdate")

	var begin string
	var end string
	var network string
	var prefixSize int

	// Enable partial state mode
	// data.Partial(true)

	addressListId := data.Get(resourceKeyAddressListID).(string)

	valBegin, beginOk := data.GetOk(resourceKeyAddressBegin)
	if beginOk {
		begin = valBegin.(string)
	}

	valEnd, endOk := data.GetOk(resourceKeyAddressEnd)
	if endOk {
		end = valEnd.(string)
	}

	valNetwork, networkOk := data.GetOk(resourceKeyAddressNetwork)
	if networkOk {
		network = valNetwork.(string)
	}

	valPrefix, prefixOk := data.GetOkExists(resourceKeyAddressPrefixSize)
	if prefixOk {
		prefixSize = valPrefix.(int)
	}

	client := provider.(*providerState).Client()

	oldBegin, _ := data.GetChange(resourceKeyAddressBegin)
	oldNetwork, _ := data.GetChange(resourceKeyAddressNetwork)

	var ipAddress string

	if oldBegin.(string) != "" {
		ipAddress = oldBegin.(string)
	} else {
		ipAddress = oldNetwork.(string)
	}

	// Update step 1: Remove old address
	_, errOld := client.DeleteAddress(addressListId, ipAddress)
	if errOld != nil {
		return errOld
	}

	// Update step 2: Add new address
	newAddress, err := client.AddAddress(addressListId, begin, end, network, prefixSize)

	if err != nil {
		return err
	}

	log.Printf("Updated address: %s", newAddress.Begin)

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	// data.Partial(false)

	return resourceAddressRead(data, provider)
}

// Delete an address list resource.
func resourceAddressDelete(data *schema.ResourceData, provider interface{}) error {
	log.Printf("resourceAddressDelete")

	addressListId := data.Get(resourceKeyAddressListID).(string)
	begin := data.Get(resourceKeyAddressBegin).(string)
	network := data.Get(resourceKeyAddressNetwork).(string)

	var ip string
	if begin != "" {
		ip = begin
	} else {
		ip = network
	}

	client := provider.(*providerState).Client()
	_, err := client.DeleteAddress(addressListId, ip)

	if err != nil {
		return err
	}

	data.SetId("")
	return nil
}
