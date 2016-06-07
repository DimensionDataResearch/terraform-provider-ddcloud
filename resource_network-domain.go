package main

import (
    "github.com/hashicorp/terraform/helper/schema"
)

func resourceNetworkDomain() *schema.Resource {
    return &schema.Resource{
        Create: resourceNetworkDomainCreate,
        Read:   resourceNetworkDomainRead,
        Update: resourceNetworkDomainUpdate,
        Delete: resourceNetworkDomainDelete,

        Schema: map[string]*schema.Schema{
			"data-center-id": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },
			"description": &schema.Schema{
                Type:     schema.TypeString,
                Required: false,
            },
			"type": &schema.Schema{
                Type:     schema.TypeString,
                Required: false,
            },
        },
    }
}

// Create a network domain resource.
func resourceNetworkDomainCreate(data *schema.ResourceData, m interface{}) error {
    return nil
}

// Read a network domain resource.
func resourceNetworkDomainRead(data *schema.ResourceData, m interface{}) error {
    return nil
}

// Update a network domain resource.
func resourceNetworkDomainUpdate(data *schema.ResourceData, m interface{}) error {
    return nil
}

// Delete a network domain resource.
func resourceNetworkDomainDelete(data *schema.ResourceData, m interface{}) error {
    return nil
}
