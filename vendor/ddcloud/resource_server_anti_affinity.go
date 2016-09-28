package ddcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyServerAntiAffinityRuleServer1ID      = "server1"
	resourceKeyServerAntiAffinityRuleServer2ID      = "server2"
)

func resourceServerAntiAffinityRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerAntiAffinityRuleCreate,
		Read:   resourceServerAntiAffinityRuleRead,
		Update: resourceServerAntiAffinityRuleUpdate,
		Delete: resourceServerAntiAffinityRuleDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyServerAntiAffinityRuleServer1ID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the first server that the anti-affinity rule relates to.",
			},
			resourceKeyServerAntiAffinityRuleServer2ID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the second server that the anti-affinity rule relates to.",
			},
		},
	}
}

// Create a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleCreate(data *schema.ResourceData, provider interface{}) error {
	server1ID := data.Get(resourceKeyServerAntiAffinityRuleServer1ID).(string)
	server2ID := data.Get(resourceKeyServerAntiAffinityRuleServer2ID).(string)

	log.Printf("Create server anti-affinity rule for servers '%s' and '%s'.", server1ID, server2ID)

	// TODO: Handle RESOURCE_BUSY response (retry?)
	apiClient := provider.(*providerState).Client()
	ruleID, err := apiClient.CreateServerAntiAffinityRule(server1ID, server2ID)
	if err != nil {
		return err
	}

	data.SetId(ruleID)

	log.Printf("Created server anti-affinity rule '%s'.", ruleID)

	return nil
}

// Read a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleRead(data *schema.ResourceData, provider interface{}) error {
	ruleID := data.Id()

	log.Printf("Read server anti-affinity rule '%s'.", ruleID)

	apiClient := provider.(*providerState).Client()

	antiAffinityRule, err := apiClient.GetServerAntiAffinityRule(ruleID)
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
		data.Set(resourceKeyServerAntiAffinityRuleServer2ID, antiAffinityRule.Servers[1].ID)
	} else {
		data.SetId("") // Mark resource as deleted.
	}

	return nil
}

// Update a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleUpdate(data *schema.ResourceData, provider interface{}) error {
	return nil // Nothing to do.
}

// Delete a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleDelete(data *schema.ResourceData, provider interface{}) error {
	var err error

	ruleID := data.Id()

	log.Printf("Delete server anti-affinity rule '%s'.", ruleID)

	apiClient := provider.(*providerState).Client()
	err = apiClient.DeleteServerAntiAffinityRule(ruleID)
	if err != nil {
		return err
	}

	log.Printf("Deleted server anti-affinity rule '%s'.", ruleID)

	return nil
}
