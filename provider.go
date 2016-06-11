package main

import (
	"compute-api/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"strings"
)

// Provider creates the Dimension Data Cloud resource provider.
func Provider() terraform.ResourceProvider {
	// TODO: Define schema and resources.

	return &schema.Provider{
		// Provider settings schema
		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},

		// Provider resource definitions
		ResourcesMap: map[string]*schema.Resource{
			// A network domain.
			"ddcloud_networkdomain": resourceNetworkDomain(),

			// A VLAN.
			"ddcloud_vlan": resourceVLAN(),

			// A server (virtual machine).
			"ddcloud_server": resourceServer(),
		},

		// Provider configuration
		ConfigureFunc: configure,
	}
}

// Configure the provider.
// Returns the provider's compute API client.
func configure(providerSettings *schema.ResourceData) (interface{}, error) {
	var (
		region   string
		username string
		password string
		client   *compute.Client
	)

	region = providerSettings.Get("region").(string)
	region = strings.ToLower(region)

	username = providerSettings.Get("username").(string)
	password = providerSettings.Get("password").(string)

	client = compute.NewClient(region, username, password)

	return client, nil
}
