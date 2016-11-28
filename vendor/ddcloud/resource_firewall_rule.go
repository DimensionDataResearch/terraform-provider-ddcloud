package ddcloud

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyFirewallRuleNetworkDomainID             = "networkdomain"
	resourceKeyFirewallRuleName                        = "name"
	resourceKeyFirewallRuleAction                      = "action"
	resourceKeyFirewallRuleEnabled                     = "enabled"
	resourceKeyFirewallRulePlacement                   = "placement"
	resourceKeyFirewallRulePlacementRelativeToRuleName = "placement_relative_to"
	resourceKeyFirewallRuleIPVersion                   = "ip_version"
	resourceKeyFirewallRuleProtocol                    = "protocol"
	resourceKeyFirewallRuleSourceAddress               = "source_address"
	resourceKeyFirewallRuleSourceNetwork               = "source_network"
	resourceKeyFirewallRuleSourceAddressListID         = "source_address_list"
	resourceKeyFirewallRuleSourcePort                  = "source_port"
	resourceKeyFirewallRuleSourcePortListID            = "source_port_list"
	resourceKeyFirewallRuleDestinationAddress          = "destination_address"
	resourceKeyFirewallRuleDestinationNetwork          = "destination_network"
	resourceKeyFirewallRuleDestinationAddressListID    = "destination_address_list"
	resourceKeyFirewallRuleDestinationPort             = "destination_port"
	resourceKeyFirewallRuleDestinationPortListID       = "destination_port_list"
	resourceCreateTimeoutFirewallRule                  = 30 * time.Minute
	resourceUpdateTimeoutFirewallRule                  = 10 * time.Minute
	resourceDeleteTimeoutFirewallRule                  = 15 * time.Minute
)

const matchAny = "any"
const computedPropertyDescription = "<computed>"

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallRuleCreate,
		Read:   resourceFirewallRuleRead,
		Update: resourceFirewallRuleUpdate,
		Delete: resourceFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyFirewallRuleNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain to which the firewall rule applies",
			},
			resourceKeyFirewallRuleName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "A name for the firewall rule",
			},
			resourceKeyFirewallRuleAction: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The action performed by the firewall rule",
				StateFunc: func(value interface{}) string {
					action := value.(string)

					return normalizeFirewallRuleAction(action)
				},
			},
			resourceKeyFirewallRuleEnabled: &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Is the firewall rule enabled",
			},
			resourceKeyFirewallRulePlacement: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     "first",
				Description: "Where in the firewall ACL this particular rule will be created",
			},
			resourceKeyFirewallRulePlacementRelativeToRuleName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     nil,
				Description: "When placement is 'before' or 'after', specifies the name of the firewall rule to which the placement instruction refers",
			},
			resourceKeyFirewallRuleIPVersion: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The IP version to which the firewall rule applies",
			},
			resourceKeyFirewallRuleProtocol: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The protocol to which the rule applies",
			},
			resourceKeyFirewallRuleSourceAddress: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The source IP address to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleSourceNetwork,
				},
			},
			resourceKeyFirewallRuleSourceNetwork: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The source IP network to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleSourceAddress,
					resourceKeyFirewallRuleSourceAddressListID,
				},
			},
			resourceKeyFirewallRuleSourceAddressListID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The Id of the source IP address list to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleSourceAddress,
					resourceKeyFirewallRuleSourceNetwork,
				},
			},
			resourceKeyFirewallRuleSourcePort: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The source port to be matched by the rule",
			},
			resourceKeyFirewallRuleSourcePortListID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The Id of the source port list to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleSourcePort,
				},
			},
			resourceKeyFirewallRuleDestinationAddress: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The destination IP address to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleDestinationNetwork,
				},
			},
			resourceKeyFirewallRuleDestinationNetwork: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The destination IP network to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleDestinationAddress,
					resourceKeyFirewallRuleDestinationAddressListID,
				},
			},
			resourceKeyFirewallRuleDestinationAddressListID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The Id of the destination IP address list to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleDestinationAddress,
					resourceKeyFirewallRuleDestinationNetwork,
				},
			},
			resourceKeyFirewallRuleDestinationPort: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The destination port to be matched by the rule",
			},
			resourceKeyFirewallRuleDestinationPortListID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The Id of the destination port list to be matched by the rule",
				ConflictsWith: []string{
					resourceKeyFirewallRuleDestinationPort,
				},
			},
		},
	}
}

// Create a firewall rule resource.
func resourceFirewallRuleCreate(data *schema.ResourceData, provider interface{}) error {
	var err error

	propertyHelper := propertyHelper(data)

	configuration := &compute.FirewallRuleConfiguration{
		Name: data.Get(resourceKeyFirewallRuleName).(string),
		Action: normalizeFirewallRuleAction(
			data.Get(resourceKeyFirewallRuleAction).(string),
		),
		Placement: compute.FirewallRulePlacement{
			Position: strings.ToUpper(
				data.Get(resourceKeyFirewallRulePlacement).(string),
			),
			RelativeToRuleName: propertyHelper.GetOptionalString(
				resourceKeyFirewallRulePlacementRelativeToRuleName, false,
			),
		},
		Enabled:         data.Get(resourceKeyFirewallRuleEnabled).(bool),
		NetworkDomainID: data.Get(resourceKeyFirewallRuleNetworkDomainID).(string),
		IPVersion: strings.ToUpper(
			data.Get(resourceKeyFirewallRuleIPVersion).(string),
		),
		Protocol: strings.ToUpper(
			data.Get(resourceKeyFirewallRuleProtocol).(string),
		),
	}

	err = configureSourceScope(propertyHelper, configuration)
	if err != nil {
		return err
	}
	err = configureDestinationScope(propertyHelper, configuration)
	if err != nil {
		return err
	}

	log.Printf("Create firewall rule '%s' in network domain '%s'.", configuration.Name, configuration.NetworkDomainID)
	log.Printf("Firewall rule configuration: '%#v'", configuration)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	var (
		ruleID      string
		createError error
	)
	operationDescription := fmt.Sprintf("Create firewall rule '%s'", configuration.Name)
	err = providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		ruleID, createError = apiClient.CreateFirewallRule(*configuration)
		if createError != nil {
			if compute.IsResourceBusyError(createError) {
				context.Retry()
			} else {
				context.Fail(createError)
			}
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	data.SetId(ruleID)

	_, err = apiClient.WaitForDeploy(compute.ResourceTypeFirewallRule, ruleID, resourceCreateTimeoutFirewallRule)

	return err
}

// Read a firewall rule resource.
func resourceFirewallRuleRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Read firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	apiClient := provider.(*providerState).Client()

	rule, err := apiClient.GetFirewallRule(id)
	if err != nil {
		return err
	}
	if rule == nil {
		log.Printf("Firewall rule '%s' has been deleted.", id)

		data.SetId("")

		return nil
	}

	data.Set(resourceKeyFirewallRuleEnabled, rule.Enabled)

	return nil
}

// Update a firewall rule resource.
func resourceFirewallRuleUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Update firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	if data.HasChange(resourceKeyFirewallRuleEnabled) {
		enable := data.Get(resourceKeyFirewallRuleEnabled).(bool)

		if enable {
			log.Printf("Enabling firewall rule '%s'...", id)
		} else {
			log.Printf("Disabling firewall rule '%s'...", id)
		}

		err := apiClient.EditFirewallRule(id, enable)
		if err != nil {
			return err
		}

		log.Printf("Updated configuration for firewall rule '%s'.", id)
	}

	return nil
}

// Delete a firewall rule resource.
func resourceFirewallRuleDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeyFirewallRuleNetworkDomainID).(string)

	log.Printf("Delete firewall rule '%s' in network domain '%s'.", id, networkDomainID)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	var deleteError error
	operationDescription := fmt.Sprintf("Delete firewall rule '%s'", id)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		deleteError = apiClient.DeleteFirewallRule(id)
		if deleteError != nil {
			if compute.IsResourceBusyError(deleteError) {
				context.Retry()
			} else {
				context.Fail(deleteError)
			}
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	return apiClient.WaitForDelete(compute.ResourceTypeFirewallRule, id, resourceDeleteTimeoutFirewallRule)
}

func configureSourceScope(propertyHelper resourcePropertyHelper, configuration *compute.FirewallRuleConfiguration) error {
	sourceAddress := propertyHelper.GetOptionalString(resourceKeyFirewallRuleSourceAddress, false)
	sourceNetwork := propertyHelper.GetOptionalString(resourceKeyFirewallRuleSourceNetwork, false)
	sourceAddressListID := propertyHelper.GetOptionalString(resourceKeyFirewallRuleSourceAddressListID, false)

	if sourceAddress != nil {
		log.Printf("Rule will match source address '%s'.", *sourceAddress)
		configuration.MatchSourceAddress(*sourceAddress)
	} else if sourceNetwork != nil {
		log.Printf("Rule will match source network '%s'.", *sourceNetwork)

		baseAddress, prefixSize, ok := parseNetworkAndPrefix(*sourceNetwork)
		if !ok {
			return fmt.Errorf("Source network '%s' for firewall rule '%s' is invalid (must be 'BaseAddress/PrefixSize')",
				*sourceNetwork,
				propertyHelper.data.Get(resourceKeyFirewallRuleName).(string),
			)
		}

		configuration.MatchSourceNetwork(baseAddress, prefixSize)
	} else if sourceAddressListID != nil {
		log.Printf("Rule will match source address list '%s'.", *sourceAddressListID)

		configuration.MatchSourceAddressList(*sourceAddressListID)
	} else {
		log.Printf("Rule will match any source address.")
		configuration.MatchAnySourceAddress()
	}

	sourcePort, err := parseFirewallPort(
		propertyHelper.GetOptionalString(resourceKeyFirewallRuleSourcePort, false),
	)
	if err != nil {
		return err
	}
	sourcePortListID := propertyHelper.GetOptionalString(resourceKeyFirewallRuleSourcePortListID, false)

	if sourcePort != nil {
		if sourcePort.End != nil {
			log.Printf("Rule will match source ports %d-%d.", sourcePort.Begin, *sourcePort.End)
			configuration.MatchSourcePortRange(sourcePort.Begin, *sourcePort.End)
		} else {
			log.Printf("Rule will match source port %d.", sourcePort.Begin)
			configuration.MatchSourcePort(sourcePort.Begin)
		}
	} else if sourcePortListID != nil {
		log.Printf("Rule will match source port list '%s'.", *sourcePortListID)
		configuration.MatchSourcePortList(*sourcePortListID)
	} else {
		log.Printf("Rule will match any destination port.")
		configuration.MatchAnySourcePort()
	}

	return nil
}

func configureDestinationScope(propertyHelper resourcePropertyHelper, configuration *compute.FirewallRuleConfiguration) error {
	destinationNetwork := propertyHelper.GetOptionalString(resourceKeyFirewallRuleDestinationNetwork, false)
	destinationAddress := propertyHelper.GetOptionalString(resourceKeyFirewallRuleDestinationAddress, false)
	destinationAddressListID := propertyHelper.GetOptionalString(resourceKeyFirewallRuleDestinationAddressListID, false)

	if destinationAddress != nil {
		log.Printf("Rule will match destination address '%s'.", *destinationAddress)
		configuration.MatchDestinationAddress(*destinationAddress)
	} else if destinationNetwork != nil {
		log.Printf("Rule will match destination network '%s'.", *destinationNetwork)

		baseAddress, prefixSize, ok := parseNetworkAndPrefix(*destinationNetwork)
		if !ok {
			return fmt.Errorf("Source network '%s' for firewall rule '%s' is invalid (must be 'BaseAddress/PrefixSize')",
				*destinationNetwork,
				propertyHelper.data.Get(resourceKeyFirewallRuleName).(string),
			)
		}

		configuration.MatchDestinationNetwork(baseAddress, prefixSize)
	} else if destinationAddressListID != nil {
		log.Printf("Rule will match destination address list '%s'.", *destinationAddressListID)

		configuration.MatchDestinationAddressList(*destinationAddressListID)
	} else {
		log.Printf("Rule will match any destination address.")
		configuration.MatchAnyDestinationAddress()
	}

	destinationPort, err := parseFirewallPort(
		propertyHelper.GetOptionalString(resourceKeyFirewallRuleDestinationPort, false),
	)
	if err != nil {
		return err
	}
	destinationPortListID := propertyHelper.GetOptionalString(resourceKeyFirewallRuleDestinationPortListID, false)

	if destinationPort != nil {
		if destinationPort.End != nil {
			log.Printf("Rule will match destination ports %d-%d.", destinationPort.Begin, *destinationPort.End)
			configuration.MatchDestinationPortRange(destinationPort.Begin, *destinationPort.End)
		} else {
			log.Printf("Rule will match destination port %d.", *destinationPort)
			configuration.MatchDestinationPort(destinationPort.Begin)
		}
	} else if destinationPortListID != nil {
		log.Printf("Rule will match destination port list '%s'.", *destinationPortListID)
		configuration.MatchDestinationPortList(*destinationPortListID)
	} else {
		log.Printf("Rule will match any destination port.")
		configuration.MatchAnyDestinationPort()
	}

	return nil
}

func normalizeFirewallRuleAction(action string) string {
	switch strings.ToLower(action) {
	case "accept":
		return compute.FirewallRuleActionAccept

	case "accept_decisively":
		return compute.FirewallRuleActionAccept

	case "allow":
		return compute.FirewallRuleActionAccept

	case "drop":
		return compute.FirewallRuleActionDrop

	case "deny":
		return compute.FirewallRuleActionDrop

	default:
		return action
	}
}

func parseFirewallPort(port *string) (*compute.FirewallRulePort, error) {
	if port == nil || *port == "any" {
		return nil, nil
	}

	portRangeComponents := strings.Split(*port, "-")

	parsedPort := &compute.FirewallRulePort{}

	parsedValue, err := strconv.Atoi(portRangeComponents[0])
	if err != nil {
		return nil, err
	}

	parsedPort.Begin = parsedValue
	if len(portRangeComponents) > 1 {
		parsedValue, err = strconv.Atoi(portRangeComponents[1])
		if err != nil {
			return nil, err
		}

		parsedPort.End = &parsedValue
	}

	return parsedPort, nil
}

func parsePortRange(portRange *string) (beginPort string, endPort *string) {
	ports := strings.Split(*portRange, "-")
	beginPort = strings.TrimSpace(ports[0])

	if len(ports) == 1 {
		return
	}

	ports[1] = strings.TrimSpace(ports[1])
	endPort = &ports[1]

	return
}

func parseNetworkAndPrefix(networkAndPrefix string) (baseAddress string, prefixSize int, ok bool) {
	networkComponents := strings.Split(networkAndPrefix, "/")
	if len(networkComponents) != 2 {
		return
	}

	baseAddress = networkComponents[0]
	prefixSize, err := strconv.Atoi(networkComponents[1])
	if err != nil {
		return
	}

	ok = true

	return
}
