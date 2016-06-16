package main

import (
	"compute-api/compute"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"strings"
)

// Provider creates the Dimension Data Cloud resource provider.
func Provider() terraform.ResourceProvider {
	// TODO: Define schema and resources.

	return &schema.Provider{
		// Provider settings schema
		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region code that identifies the target end-point for the Dimension Data Cloud Compute API.",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The user name used to authenticate to the Dimension Data Cloud Compute API (if not specified, then the DD_COMPUTE_USER environment variable will be used).",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The password used to authenticate to the Dimension Data Cloud Compute API (if not specified, then the DD_COMPUTE_PASSWORD environment variable will be used).",
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

			// A Network Address Translation (NAT) rule.
			"ddcloud_nat": resourceNAT(),

			// A firewall rule.
			"ddcloud_firewall_rule": resourceFirewallRule(),
		},

		// Provider configuration
		ConfigureFunc: configureProvider,
	}
}

// Configure the provider.
// Returns the provider's compute API client.
func configureProvider(providerSettings *schema.ResourceData) (interface{}, error) {
	var (
		region   string
		username *string
		password *string
		client   *compute.Client
	)

	region = providerSettings.Get("region").(string)
	region = strings.ToLower(region)

	propertyHelper := propertyHelper(providerSettings)
	username = propertyHelper.GetOptionalString("username", false)
	if username == nil {
		computeUser := os.Getenv("DD_COMPUTE_USER")
		if len(computeUser) == 0 {
			return nil, fmt.Errorf("The 'username' property was not specified for the 'ddcloud' provider, and the 'DD_COMPUTE_USER' environment variable is not present. Please supply either one of these to configure the user name used to authenticate to Dimension Data Cloud Compute.")
		}

		username = &computeUser
	}
	password = propertyHelper.GetOptionalString("password", false)
	if password == nil {
		computePassword := os.Getenv("DD_COMPUTE_PASSWORD")
		if len(computePassword) == 0 {
			return nil, fmt.Errorf("The 'password' property was not specified for the 'ddcloud' provider, and the 'DD_COMPUTE_PASSWORD' environment variable is not present. Please supply either one of these to configure the password used to authenticate to Dimension Data Cloud Compute.")
		}

		password = &computePassword
	}

	client = compute.NewClient(region, *username, *password)

	return client, nil
}
