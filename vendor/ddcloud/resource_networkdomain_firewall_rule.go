package ddcloud

import (
	"fmt"
	"log"
	"strings"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyNetworkDomainFirewallRuleType    = "type"
	resourceKeyNetworkDomainFirewallRuleEnabled = "enabled"
)

const defaultFirewallRulePrefix = "CCDEFAULT."

func defaultFirewallRuleNameToType(ruleName string) string {
	return strings.TrimPrefix(ruleName, defaultFirewallRulePrefix)
}
func defaultFirewallRuleTypeToName(ruleType string) string {
	return defaultFirewallRulePrefix + ruleType
}

// Well-known default rule types.
var defaultRuleTypes = []string{
	"BlockOutboundMailIPv4",
	"BlockOutboundMailIPv4Secure",
	"BlockOutboundMailIPv6",
	"BlockOutboundMailIPv6Secure",
	"DenyExternalInboundIPv6",
}

// Schema for ddcloud_networkdomain.default_firewall_rule
func schemaNetworkDomainFirewallRule() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "One or more default firewall rules (name starts with 'CCDefault.') for the network domain",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyNetworkDomainFirewallRuleType: &schema.Schema{
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateNetworkDomainFirewallRuleType,
					Description:  "The type of default firewall rule (without the 'CCDefault.' prefix)",
				},
				resourceKeyNetworkDomainFirewallRuleEnabled: &schema.Schema{
					Type:        schema.TypeBool,
					Required:    true,
					Description: "Is the firewall rule enabled",
				},
			},
		},
		Set: hashNetworkDomainFirewallRule,
	}
}

// Read state for a network domain's default firewall rules
func readNetworkDomainDefaultFirewallRules(data *schema.ResourceData, apiClient *compute.Client) error {
	networkDomainID := data.Id()
	value, ok := data.GetOk(resourceKeyNetworkDomainFirewallRule)
	if !ok {
		return nil
	}
	defaultRules := value.(*schema.Set).List()

	log.Printf("Read default firewall rules for network domain '%s'...",
		networkDomainID,
	)

	existingRules, err := getNetworkDomainDefaultFirewallRulesByType(data, apiClient)
	if err != nil {
		return err
	}

	// We only update state for rules already defined in the network domain's configuration
	for _, defaultRule := range defaultRules {
		ruleProperties := defaultRule.(map[string]interface{})
		ruleType := ruleProperties[resourceKeyNetworkDomainFirewallRuleType].(string)
		existingRule, ok := existingRules[ruleType]
		if !ok {
			return fmt.Errorf("No firewall rule named '%s' found in network domain '%s'",
				defaultFirewallRuleTypeToName(ruleType),
				networkDomainID,
			)
		}

		ruleProperties[resourceKeyNetworkDomainFirewallRuleEnabled] = existingRule.Enabled
	}

	// Persist changes.
	defaultRuleSet := schema.NewSet(hashNetworkDomainFirewallRule, defaultRules)
	data.Set(resourceKeyNetworkDomainFirewallRule, defaultRuleSet)
	data.SetPartial(resourceKeyNetworkDomainFirewallRule)

	return nil
}

func applyNetworkDomainDefaultFirewallRules(data *schema.ResourceData, apiClient *compute.Client) error {
	networkDomainID := data.Id()
	defaultRuleSet, ok := data.GetOk(resourceKeyNetworkDomainFirewallRule)
	if !ok {
		return nil
	}

	log.Printf("Configure default firewall rules for network domain '%s'...",
		networkDomainID,
	)

	existingRules, err := getNetworkDomainDefaultFirewallRulesByType(data, apiClient)
	if err != nil {
		return err
	}

	for _, defaultRule := range defaultRuleSet.(*schema.Set).List() {
		ruleProperties := defaultRule.(map[string]interface{})

		ruleType := ruleProperties[resourceKeyNetworkDomainFirewallRuleType].(string)
		existingRule := existingRules[ruleType]

		ruleEnabled := ruleProperties[resourceKeyNetworkDomainFirewallRuleEnabled].(bool)
		if ruleEnabled == existingRule.Enabled {
			continue // Nothing to do
		}

		log.Printf("Updating default firewall rule '%s' (%s): Enabled = %t",
			existingRule.Name,
			existingRule.ID,
			ruleEnabled,
		)
		err = apiClient.EditFirewallRule(existingRule.ID, ruleEnabled)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get all default (built-in) firewall rules in a network domain, keyed by type (name without prefix).
func getNetworkDomainDefaultFirewallRulesByType(data *schema.ResourceData, apiClient *compute.Client) (rules map[string]compute.FirewallRule, err error) {
	networkDomainID := data.Id()

	page := compute.DefaultPaging()
	page.PageSize = 50

	rules = make(map[string]compute.FirewallRule)

	for {
		var results *compute.FirewallRules
		results, err = apiClient.ListFirewallRules(networkDomainID, page)
		if err != nil {
			return
		}
		if results.IsEmpty() {
			break
		}

		for _, rule := range results.Rules {
			if rule.RuleType != "DEFAULT_RULE" {
				continue
			}

			ruleType := defaultFirewallRuleNameToType(rule.Name)
			rules[ruleType] = rule
		}

		page.Next()
	}

	return
}

// Produce a hash code to represent the values for a network domain default firewall rule.
func hashNetworkDomainFirewallRule(value interface{}) int {
	firewallRule, ok := value.(map[string]interface{})
	if !ok {
		return -1
	}

	return schema.HashString(fmt.Sprintf("%s=%t",
		firewallRule[resourceKeyNetworkDomainFirewallRuleType],
		firewallRule[resourceKeyNetworkDomainFirewallRuleEnabled],
	))
}

func validateNetworkDomainFirewallRuleType(value interface{}, propertyName string) (messages []string, errors []error) {
	ruleType, ok := value.(string)
	if !ok {
		errors = append(errors,
			fmt.Errorf("Invalid data type for '%s' (expected string)", propertyName),
		)

		return
	}

	for _, defaultRuleType := range defaultRuleTypes {
		if ruleType == defaultRuleType {
			return // Valid
		}
	}

	errors = append(errors,
		fmt.Errorf("Invalid firewall rule type '%s' (expected one of [%s])",
			ruleType,
			strings.Join(defaultRuleTypes, ","),
		),
	)

	return
}
