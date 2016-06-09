package main

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyVLANNetworkDomainID = "networkdomain"
	resourceKeyVLANName            = "name"
	resourceKeyVLANDescription     = "description"
	resourceKeyVLANIPv4BaseAddress = "ipv4_base_address"
	resourceKeyVLANIPv4PrefixSize  = "ipv4_prefix_size"
)

func resourceVLAN() *schema.Resource {
	return &schema.Resource{
		Create: resourceVLANCreate,
		Read:   resourceVLANRead,
		Update: resourceVLANUpdate,
		Delete: resourceVLANDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyVLANNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Id of the network domain in which the VLAN is deployed.",
			},
			resourceKeyVLANName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VLAN display name.",
			},
			resourceKeyVLANDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The VLAN description.",
			},
			resourceKeyVLANIPv4BaseAddress: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VLAN's private IPv4 base address.",
			},
			resourceKeyVLANIPv4PrefixSize: &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The VLAN's private IPv4 prefix length.",
			},
		},
	}
}

// Create a VLAN resource.
func resourceVLANCreate(data *schema.ResourceData, provider interface{}) error {
	var (
		networkDomainID, name, description, ipv4BaseAddress string
		ipv4PrefixSize                                      int
	)

	networkDomainID = data.Get(resourceKeyVLANNetworkDomainID).(string)
	name = data.Get(resourceKeyVLANName).(string)
	description = data.Get(resourceKeyVLANDescription).(string)
	ipv4BaseAddress = data.Get(resourceKeyVLANIPv4BaseAddress).(string)
	ipv4PrefixSize = data.Get(resourceKeyVLANIPv4PrefixSize).(int)

	log.Printf("Create VLAN '%s' ('%s') in network domain '%s' (IPv4 network = '%s/%d').", name, description, networkDomainID, ipv4BaseAddress, ipv4PrefixSize)

	providerClient := provider.(*compute.Client)

	vlanID, err := providerClient.DeployVLAN(networkDomainID, name, description, ipv4BaseAddress, ipv4PrefixSize)
	if err != nil {
		return err
	}

	data.SetId(vlanID)

	return nil
}

// Read a VLAN resource.
func resourceVLANRead(data *schema.ResourceData, provider interface{}) error {
	var (
		id, networkDomainID, name, description, ipv4BaseAddress string
		ipv4PrefixSize                                          int
	)

	id = data.Id()
	networkDomainID = data.Get(resourceKeyVLANNetworkDomainID).(string)
	name = data.Get(resourceKeyVLANName).(string)
	description = data.Get(resourceKeyVLANDescription).(string)
	ipv4BaseAddress = data.Get(resourceKeyVLANIPv4BaseAddress).(string)
	ipv4PrefixSize = data.Get(resourceKeyVLANIPv4PrefixSize).(int)

	log.Printf("Read VLAN '%s' (Name = '%s', description = '%s') in network domain '%s' (IPv4 network = '%s/%d').", id, name, description, networkDomainID, ipv4BaseAddress, ipv4PrefixSize)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to retrieve the VLAN.

	return nil
}

// Update a VLAN resource.
func resourceVLANUpdate(data *schema.ResourceData, provider interface{}) error {
	var (
		id, name, description, ipv4BaseAddress string
		ipv4PrefixSize                         int
	)

	id = data.Id()

	if data.HasChange(resourceKeyVLANName) {
		name = data.Get(resourceKeyVLANName).(string)
	}

	if data.HasChange(resourceKeyVLANDescription) {
		description = data.Get(resourceKeyVLANDescription).(string)
	}

	if data.HasChange(resourceKeyVLANIPv4BaseAddress) {
		ipv4BaseAddress = data.Get(resourceKeyVLANIPv4BaseAddress).(string)
	}

	if data.HasChange(resourceKeyVLANIPv4PrefixSize) {
		ipv4PrefixSize = data.Get(resourceKeyVLANIPv4PrefixSize).(int)
	}

	log.Printf("Update VLAN '%s' (Name = '%s', description = '%s', IPv4 network = '%s/%d').", id, name, description, ipv4BaseAddress, ipv4PrefixSize)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with actual API call to edit VLAN.

	return nil
}

// Delete a VLAN resource.
func resourceVLANDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVLANName).(string)
	networkDomainID := data.Get(resourceKeyVLANNetworkDomainID).(string)

	log.Printf("Delete VLAN '%s' ('%s') in network domain '%s'.", id, name, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with actual API call to edit VLAN.

	return nil
}
