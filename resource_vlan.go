package main

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
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

	log.Printf("VLAN '%s' is being provisioned...", networkDomainID)

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			return fmt.Errorf("Timed out after waiting %d seconds for provisioning of VLAN '%s' to complete.", 60, vlanID)

		case <-ticker.C:
			log.Printf("Polling status for VLAN '%s'...", networkDomainID)
			vlan, err := providerClient.GetVLAN(vlanID)
			if err != nil {
				return err
			}

			if vlan == nil {
				return fmt.Errorf("Newly-created network domain was not found with Id '%s'.", vlanID)
			}

			switch vlan.State {
			case "PENDING_ADD":
				log.Printf("VLAN '%s' is still being provisioned...", vlanID)

				continue
			case "NORMAL":
				log.Printf("VLAN '%s' has been successfully provisioned.", networkDomainID)

				return nil
			default:
				log.Printf("Unexpected status for VLAN '%s' ('%s').", vlanID, vlan.State)

				return fmt.Errorf("Failed to provision VLAN '%s' ('%s'): encountered unexpected state '%s'.", vlanID, name, vlan.State)
			}
		}
	}
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

	vlan, err := providerClient.GetVLAN(id)
	if err != nil {
		return err
	}

	if vlan != nil {
		data.Set(resourceKeyVLANName, vlan.Name)
		data.Set(resourceKeyNetworkDomainDescription, vlan.Description)
	} else {
		data.SetId("") // Mark resource as deleted.
	}

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
