package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyServerAntiAffinityRuleServer1ID       = "server1"
	resourceKeyServerAntiAffinityRuleServer1Name     = "server1_name"
	resourceKeyServerAntiAffinityRuleServer2ID       = "server2"
	resourceKeyServerAntiAffinityRuleServer2Name     = "server2_name"
	resourceKeyServerAntiAffinityRuleNetworkDomainID = "networkdomain"
	resourceDeleteTimeoutServerAntiAffinityRule      = 5 * time.Minute
)

func resourceServerAntiAffinityRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerAntiAffinityRuleCreate,
		Read:   resourceServerAntiAffinityRuleRead,
		Delete: resourceServerAntiAffinityRuleDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyServerAntiAffinityRuleServer1ID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the first server that the anti-affinity rule relates to.",
			},
			resourceKeyServerAntiAffinityRuleServer1Name: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the first server that the anti-affinity rule relates to.",
			},
			resourceKeyServerAntiAffinityRuleServer2ID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the second server that the anti-affinity rule relates to.",
			},
			resourceKeyServerAntiAffinityRuleServer2Name: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the second server that the anti-affinity rule relates to.",
			},
			resourceKeyServerAntiAffinityRuleNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Id of the network domain in which the anti-affinity rule applies.",
			},
		},
	}
}

// Create a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleCreate(data *schema.ResourceData, provider interface{}) error {
	server1ID := data.Get(resourceKeyServerAntiAffinityRuleServer1ID).(string)
	server2ID := data.Get(resourceKeyServerAntiAffinityRuleServer2ID).(string)

	log.Printf("Create server anti-affinity rule for servers '%s' and '%s'.", server1ID, server2ID)

	apiClient := provider.(*providerState).Client()

	// Capture server details

	server1, err := apiClient.GetServer(server1ID)
	if err != nil {
		return err
	}
	if server1 == nil {
		return fmt.Errorf("Cannot create anti-affinity rule (server 1 not found with Id '%s')", server1ID)
	}

	server2, err := apiClient.GetServer(server2ID)
	if err != nil {
		return err
	}
	if server2 == nil {
		return fmt.Errorf("Cannot create anti-affinity rule (server 2 not found with Id '%s')", server2ID)
	}

	// We don't support anti-affinity rules between servers in different network domains.
	if server1.Network.NetworkDomainID != server2.Network.NetworkDomainID {
		return fmt.Errorf("Cannot create server anti-affinity rule (server '%s' is in network domain '%s', but server '%s' is in network domain '%s'", server1ID, server1.Network.NetworkDomainID, server2ID, server2.Network.NetworkDomainID)
	}

	// TODO: Handle RESOURCE_BUSY response (retry?)
	ruleID, err := apiClient.CreateServerAntiAffinityRule(server1ID, server2ID)
	if err != nil {
		return err
	}

	data.SetId(ruleID)

	log.Printf("Created server anti-affinity rule '%s'.", ruleID)

	data.Set(resourceKeyServerAntiAffinityRuleServer1Name, server1.Name)
	data.Set(resourceKeyServerAntiAffinityRuleServer2Name, server2.Name)
	data.Set(resourceKeyServerAntiAffinityRuleNetworkDomainID, server1.Network.NetworkDomainID)

	return nil
}

// Read a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleRead(data *schema.ResourceData, provider interface{}) error {
	ruleID := data.Id()
	server1Name := data.Get(resourceKeyServerAntiAffinityRuleServer1Name).(string)
	server2Name := data.Get(resourceKeyServerAntiAffinityRuleServer1Name).(string)
	networkDomainID := data.Get(resourceKeyServerAntiAffinityRuleNetworkDomainID).(string)

	log.Printf("Read server anti-affinity rule '%s' (servers '%s' and '%s').", ruleID, server1Name, server2Name)

	apiClient := provider.(*providerState).Client()

	antiAffinityRule, err := apiClient.GetServerAntiAffinityRule(ruleID, networkDomainID)
	if err != nil {
		return err
	}

	if antiAffinityRule != nil {
		if len(antiAffinityRule.Servers) != 2 {
			return fmt.Errorf("Anti-affinity rule relates to unexpected number of servers (%d).",
				len(antiAffinityRule.Servers),
			)
		}

		data.Set(resourceKeyServerAntiAffinityRuleServer1ID, antiAffinityRule.Servers[0].ID)
		data.Set(resourceKeyServerAntiAffinityRuleServer1Name, antiAffinityRule.Servers[0].Name)
		data.Set(resourceKeyServerAntiAffinityRuleServer2ID, antiAffinityRule.Servers[1].ID)
		data.Set(resourceKeyServerAntiAffinityRuleServer2Name, antiAffinityRule.Servers[1].Name)
	} else {
		data.SetId("") // Mark resource as deleted.
	}

	return nil
}

// Delete a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleDelete(data *schema.ResourceData, provider interface{}) error {
	ruleID := data.Id()
	networkDomainID := data.Get(resourceKeyServerAntiAffinityRuleNetworkDomainID).(string)

	log.Printf("Delete server anti-affinity rule '%s' in network domain '%s'.", ruleID, networkDomainID)

	apiClient := provider.(*providerState).Client()
	err := apiClient.DeleteServerAntiAffinityRule(ruleID, networkDomainID)
	if err != nil {
		return err
	}

	log.Printf("Deleting server anti-affinity rule '%s' in network domain '%s'...", ruleID, networkDomainID)

	qualifiedRuleID := networkDomainID + "/" + ruleID
	err = apiClient.WaitForDelete(compute.ResourceTypeServerAntiAffinityRule, qualifiedRuleID, resourceDeleteTimeoutServerAntiAffinityRule)
	if err != nil {
		return err
	}

	log.Printf("Deleted server anti-affinity rule '%s' in network domain '%s'.", ruleID, networkDomainID)

	return nil
}
