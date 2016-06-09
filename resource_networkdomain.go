package main

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyNetworkDomainName           = "name"
	resourceKeyNetworkDomainDescription    = "description"
	resourceKeyNetworkDomainPlan           = "plan"
	resourceKeyNetworkDomainDataCenter     = "datacenter"
	resourceKeyNetworkDomainNatIPv4Address = "nat_ipv4_address"
)

func resourceNetworkDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkDomainCreate,
		Read:   resourceNetworkDomainRead,
		Update: resourceNetworkDomainUpdate,
		Delete: resourceNetworkDomainDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNetworkDomainName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			resourceKeyNetworkDomainDescription: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyNetworkDomainPlan: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ESSENTIALS",
			},
			resourceKeyNetworkDomainDataCenter: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetworkDomainNatIPv4Address: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// Create a network domain resource.
func resourceNetworkDomainCreate(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get(resourceKeyNetworkDomainName).(string)
	description = data.Get(resourceKeyNetworkDomainDataCenter).(string)
	plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	dataCenterID = data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Create network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)

	networkDomainID, err := providerClient.DeployNetworkDomain(name, description, plan, dataCenterID)
	if err != nil {
		return err
	}

	data.SetId(networkDomainID)

	log.Printf("Network domain '%s' is being provisioned...", networkDomainID)

	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			return fmt.Errorf("Timed out after waiting %d seconds for provisioning of network domain '%s' to complete.", 60, networkDomainID)

		case <-ticker.C:
			log.Printf("Polling status for network domain '%s'...", networkDomainID)
			networkDomain, err := providerClient.GetNetworkDomain(networkDomainID)
			if err != nil {
				return err
			}

			if networkDomain == nil {
				return fmt.Errorf("Newly-created network domain was not found with Id '%s'.", networkDomainID)
			}

			switch networkDomain.State {
			case "PENDING_ADD":
				log.Printf("Network domain '%s' is still being provisioned...", networkDomainID)

				continue
			case "NORMAL":
				log.Printf("Network domain '%s' has been successfully provisioned.", networkDomainID)

				// Capture IPv4 NAT address.
				data.Set(resourceKeyNetworkDomainNatIPv4Address, networkDomain.NatIPv4Address)

				return nil
			default:
				log.Printf("Unexpected status for network domain '%s' ('%s').", networkDomainID, networkDomain.State)

				return fmt.Errorf("Failed to provision network domain '%s' ('%s'): encountered unexpected state '%s'.", networkDomainID, name, networkDomain.State)
			}
		}
	}
}

// Read a network domain resource.
func resourceNetworkDomainRead(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	id := data.Id()
	name = data.Get(resourceKeyNetworkDomainName).(string)
	description = data.Get(resourceKeyNetworkDomainDescription).(string)
	plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	dataCenterID = data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Read network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, id, dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)

	networkDomain, err := providerClient.GetNetworkDomain(id)
	if err != nil {
		return err
	}

	if networkDomain != nil {
		data.Set(resourceKeyNetworkDomainName, networkDomain.Name)
		data.Set(resourceKeyNetworkDomainDescription, networkDomain.Description)
		data.Set(resourceKeyNetworkDomainPlan, networkDomain.Type)
		data.Set(resourceKeyNetworkDomainDataCenter, networkDomain.DatacenterID)
		data.Set(resourceKeyNetworkDomainNatIPv4Address, networkDomain.NatIPv4Address)
	} else {
		data.SetId("") // Mark resource as deleted.
	}

	return nil
}

// Update a network domain resource.
func resourceNetworkDomainUpdate(data *schema.ResourceData, provider interface{}) error {
	var id, name, description, plan string

	id = data.Id()

	if data.HasChange(resourceKeyNetworkDomainName) {
		name = data.Get(resourceKeyNetworkDomainName).(string)
	}

	if data.HasChange(resourceKeyNetworkDomainDescription) {
		description = data.Get(resourceKeyNetworkDomainDescription).(string)
	}

	if data.HasChange(resourceKeyNetworkDomainPlan) {
		plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	}

	log.Printf("Update network domain '%s' (Name = '%s', Description = '%s', Plan = '%s').", data.Id(), name, description, plan)

	providerClient := provider.(*compute.Client)

	return providerClient.EditNetworkDomain(id, name, description, plan)
}

// Delete a network domain resource.
func resourceNetworkDomainDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyNetworkDomainName).(string)
	dataCenterID := data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Delete network domain '%s' ('%s') in data center '%s'.", id, name, dataCenterID)

	providerClient := provider.(*compute.Client)
	err := providerClient.DeleteNetworkDomain(id)

	return err
}
