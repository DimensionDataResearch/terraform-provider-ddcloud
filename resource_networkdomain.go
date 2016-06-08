package main

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyNetworkDomainName = "name"
	resourceKeyNetworkDomainDescription = "description"
	resourceKeyNetworkDomainPlan = "plan"
	resourceKeyNetworkDomainDataCenter = "datacenter"
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
				Default: "",
			},
			resourceKeyNetworkDomainPlan: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default: "ESSENTIALS",
			},
			resourceKeyNetworkDomainDataCenter: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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

	return nil
}

// Read a network domain resource.
func resourceNetworkDomainRead(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get(resourceKeyNetworkDomainName).(string)
	description = data.Get(resourceKeyNetworkDomainDescription).(string)
	plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	dataCenterID = data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Read network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, data.Id(), dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to retrieve the network domain.

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
	var name, description, plan, dataCenterID string

	name = data.Get(resourceKeyNetworkDomainName).(string)
	description = data.Get(resourceKeyNetworkDomainDescription).(string)
	plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	dataCenterID = data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Delete network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, data.Id(), dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to delete the network domain.

	return nil
}
