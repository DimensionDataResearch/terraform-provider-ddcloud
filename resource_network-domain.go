package main

import (
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

	log.Printf("Read network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	return nil
}

// Update a network domain resource.
func resourceNetworkDomainUpdate(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get("name").(string)
	description = data.Get("description").(string)
	plan = data.Get("plan").(string)
	dataCenterID = data.Get("data-center-id").(string)

	log.Printf("Update network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	return nil
}

// Delete a network domain resource.
func resourceNetworkDomainDelete(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get("name").(string)
	description = data.Get("description").(string)
	plan = data.Get("plan").(string)
	dataCenterID = data.Get("data-center-id").(string)

	log.Printf("Delete network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	return nil
}
