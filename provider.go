package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider creates the Dimension Data Cloud resource provider.
func Provider() terraform.ResourceProvider {
	// TODO: Define schema and resources.

	return &schema.Provider{
		// Provider settings
		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"userName": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ResourcesMap:  map[string]*schema.Resource{
			// A network domain.
			"ddcloud_network-domain": resourceNetworkDomain(),
		},
		ConfigureFunc: configure,
	}
}

// Configure the provider client.
func configure(data *schema.ResourceData) (interface{}, error) {
	// TODO: Create provider client.

	return nil, fmt.Errorf("Not implemented yet.")
}
