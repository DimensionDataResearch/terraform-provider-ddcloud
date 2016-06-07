package main

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceNetworkDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkDomainCreate,
		Read:   resourceNetworkDomainRead,
		Update: resourceNetworkDomainUpdate,
		Delete: resourceNetworkDomainDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
			},
			"plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
			},
			"data-center-id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

// Create a network domain resource.
func resourceNetworkDomainCreate(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get("name").(string)
	description = data.Get("description").(string)
	plan = data.Get("plan").(string)
	dataCenterID = data.Get("data-center-id").(string)

	log.Printf("Create network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to create a network domain.

	data.SetId(name) // TODO: Use CaaS domain Id instead.

	return nil
}

// Read a network domain resource.
func resourceNetworkDomainRead(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get("name").(string)
	description = data.Get("description").(string)
	plan = data.Get("plan").(string)
	dataCenterID = data.Get("data-center-id").(string)

	log.Printf("Read network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, data.Id(), dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to retrieve the network domain.

	return nil
}

// Update a network domain resource.
func resourceNetworkDomainUpdate(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get("name").(string)
	description = data.Get("description").(string)
	plan = data.Get("plan").(string)
	dataCenterID = data.Get("data-center-id").(string)

	log.Printf("Update network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, data.Id(), dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to update the network domain.

	return nil
}

// Delete a network domain resource.
func resourceNetworkDomainDelete(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get("name").(string)
	description = data.Get("description").(string)
	plan = data.Get("plan").(string)
	dataCenterID = data.Get("data-center-id").(string)

	log.Printf("Delete network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, data.Id(), dataCenterID, plan, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with call to delete the network domain.

	return nil
}
