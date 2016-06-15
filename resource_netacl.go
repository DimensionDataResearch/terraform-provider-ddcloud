package main

import (
	"compute-api/compute"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

/*
name 					= "test-vm HTTP Inbound" # Do we want to try matching on name before matching on index? Ask around to determine people's preference.
index					= 20 # If this is omitted, then the first available index is selected.
action					= "accept" # Valid values are "accept" or "drop."
enabled					= true


ip_version				= "ipv4"
protocol				= "tcp"

# source_address is computed at deploy time (="any").
# source_port is computed at deploy time (="any).
# You can also specify source_network (e.g. 10.2.198.0/24) instead of source_address.
# For a ddcloud_vlan, you can obtain these values using the ipv4_baseaddress and ipv4_prefixsize properties.

destination_address		= "${ddcloud_nat.test-vm-nat.public_ipv4}" # You can also specify destination_network instead of source_address.
destination_port 		= "80,443"
*/

const (
	resourceKeyNetACLNetworkDomainID    = "networkdomain"
	resourceKeyNetACLName               = "name"
	resourceKeyNetACLIndex              = "index"
	resourceKeyNetACLAction             = "action"
	resourceKeyNetACLEnabled            = "enabled"
	resourceKeyNetACLIPVersion          = "ip_version"
	resourceKeyNetACLProtocol           = "protocol"
	resourceKeyNetACLSourceAddress      = "source_address"
	resourceKeyNetACLSourceNetwork      = "source_network"
	resourceKeyNetACLSourcePort         = "source_port"
	resourceKeyNetACLDestinationAddress = "destination_address"
	resourceKeyNetACLDestinationNetwork = "destination_network"
	resourceKeyNetACLDestinationPort    = "destination_port"
	resourceCreateTimeoutNetACL         = 30 * time.Minute
	resourceUpdateTimeoutNetACL         = 10 * time.Minute
	resourceDeleteTimeoutNetACL         = 15 * time.Minute
)

const computedPropertyDescription = "<computed>"

func resourceNetACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetACLCreate,
		Read:   resourceNetACLRead,
		Update: resourceNetACLUpdate,
		Delete: resourceNetACLDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNetACLNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetACLName: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetACLIndex: &schema.Schema{
				Type:     schema.TypeInt,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetACLAction: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetACLEnabled: &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			resourceKeyNetACLIPVersion: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetACLProtocol: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyNetACLSourceAddress: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "any",
				ConflictsWith: []string{
					resourceKeyNetACLSourceNetwork,
				},
			},
			resourceKeyNetACLSourceNetwork: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyNetACLSourceAddress,
				},
			},
			resourceKeyNetACLSourcePort: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "any",
			},
			resourceKeyNetACLDestinationAddress: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyNetACLDestinationNetwork,
				},
			},
			resourceKeyNetACLDestinationNetwork: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyNetACLDestinationAddress,
				},
			},
			resourceKeyNetACLDestinationPort: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

// Create a NetACL resource.
func resourceNetACLCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyNetACLNetworkDomainID).(string)

	log.Printf("Create firewall rule in network domain '%s'.", networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	// Dummy Id.
	data.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	return nil
}

// Read a NetACL resource.
func resourceNetACLRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyNetACLNetworkDomainID).(string)

	log.Printf("Read firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}

// Update a NetACL resource.
func resourceNetACLUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyNetACLNetworkDomainID).(string)

	log.Printf("Update firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}

// Delete a NetACL resource.
func resourceNetACLDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyNetACLNetworkDomainID).(string)

	log.Printf("Delete firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}
