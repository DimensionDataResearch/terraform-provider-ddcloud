package main

import (
	"compute-api/compute"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyFirewallRuleNetworkDomainID    = "networkdomain"
	resourceKeyFirewallRuleName               = "name"
	resourceKeyFirewallRuleIndex              = "index"
	resourceKeyFirewallRuleAction             = "action"
	resourceKeyFirewallRuleEnabled            = "enabled"
	resourceKeyFirewallRuleIPVersion          = "ip_version"
	resourceKeyFirewallRuleProtocol           = "protocol"
	resourceKeyFirewallRuleSourceAddress      = "source_address"
	resourceKeyFirewallRuleSourceNetwork      = "source_network"
	resourceKeyFirewallRuleSourcePort         = "source_port"
	resourceKeyFirewallRuleDestinationAddress = "destination_address"
	resourceKeyFirewallRuleDestinationNetwork = "destination_network"
	resourceKeyFirewallRuleDestinationPort    = "destination_port"
	resourceCreateTimeoutFirewallRule         = 30 * time.Minute
	resourceUpdateTimeoutFirewallRule         = 10 * time.Minute
	resourceDeleteTimeoutFirewallRule         = 15 * time.Minute
)

const computedPropertyDescription = "<computed>"

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallRuleCreate,
		Read:   resourceFirewallRuleRead,
		Update: resourceFirewallRuleUpdate,
		Delete: resourceFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyFirewallRuleNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyFirewallRuleName: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyFirewallRuleIndex: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "FIRST",
			},
			resourceKeyFirewallRuleAction: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyFirewallRuleEnabled: &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			resourceKeyFirewallRuleIPVersion: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyFirewallRuleProtocol: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyFirewallRuleSourceAddress: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "any",
				ConflictsWith: []string{
					resourceKeyFirewallRuleSourceNetwork,
				},
			},
			resourceKeyFirewallRuleSourceNetwork: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyFirewallRuleSourceAddress,
				},
			},
			resourceKeyFirewallRuleSourcePort: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "any",
			},
			resourceKeyFirewallRuleDestinationAddress: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyFirewallRuleDestinationNetwork,
				},
			},
			resourceKeyFirewallRuleDestinationNetwork: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyFirewallRuleDestinationAddress,
				},
			},
			resourceKeyFirewallRuleDestinationPort: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

// Create a firewall rule resource.
func resourceFirewallRuleCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Create firewall rule in network domain '%s'.", networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	// Dummy Id.
	data.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	return nil
}

// Read a firewall rule resource.
func resourceFirewallRuleRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Read firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}

// Update a firewall rule resource.
func resourceFirewallRuleUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Update firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}

// Delete a firewall rule resource.
func resourceFirewallRuleDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Delete firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}
