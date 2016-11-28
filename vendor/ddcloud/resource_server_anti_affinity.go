package ddcloud

import (
	"fmt"
	"log"
	"time"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerAntiAffinityRuleServer1ID       = "server1"
	resourceKeyServerAntiAffinityRuleServer1Name     = "server1_name"
	resourceKeyServerAntiAffinityRuleServer2ID       = "server2"
	resourceKeyServerAntiAffinityRuleServer2Name     = "server2_name"
	resourceKeyServerAntiAffinityRuleNetworkDomainID = "networkdomain"
	resourceCreateTimeoutServerAntiAffinityRule      = 5 * time.Minute
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

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

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

	networkDomainID := server1.Network.NetworkDomainID

	var (
		ruleID      string
		createError error
	)
	operationDescription := fmt.Sprintf("Create anti-affinity rule between servers '%s' and '%s'", server1ID, server2ID)
	err = providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Create server anti-affinity rule '%s'", networkDomainID)
		defer asyncLock.Release()

		ruleID, createError = apiClient.CreateServerAntiAffinityRule(server1ID, server2ID)
		if compute.IsResourceBusyError(createError) {
			context.Retry()
		} else if createError != nil {
			context.Fail(createError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	data.SetId(ruleID)

	qualifiedRuleID := networkDomainID + "/" + ruleID
	resource, err := apiClient.WaitForChange(compute.ResourceTypeServerAntiAffinityRule, qualifiedRuleID, "Create", resourceCreateTimeoutServerAntiAffinityRule)
	if err != nil {
		return err
	}

	antiAffinityRule := resource.(*compute.ServerAntiAffinityRule)
	if antiAffinityRule == nil {
		return fmt.Errorf("Cannot find newly-created server anti-affinity rule '%s' in network domain '%s'.", ruleID, networkDomainID)
	}

	log.Printf("Created server anti-affinity rule '%s'.", ruleID)

	// CloudControl makes no guarantees about the order in which the target servers are returned
	serversByID := make(map[string]compute.ServerSummary)
	for _, server := range antiAffinityRule.Servers {
		serversByID[server.ID] = server
	}

	targetServer1, ok := serversByID[server1ID]
	if !ok {
		return fmt.Errorf("Anti-affinity rule '%s' targets unexpected server ('%s')", ruleID, server1ID)
	}

	targetServer2, ok := serversByID[server2ID]
	if !ok {
		return fmt.Errorf("Anti-affinity rule '%s' targets unexpected server ('%s')", ruleID, server2ID)
	}

	data.Set(resourceKeyServerAntiAffinityRuleServer1Name, targetServer1.Name)
	data.Set(resourceKeyServerAntiAffinityRuleServer2Name, targetServer2.Name)
	data.Set(resourceKeyServerAntiAffinityRuleNetworkDomainID, server1.Network.NetworkDomainID)

	return nil
}

// Read a server anti-affinity rule resource.
func resourceServerAntiAffinityRuleRead(data *schema.ResourceData, provider interface{}) error {
	ruleID := data.Id()
	server1Name := data.Get(resourceKeyServerAntiAffinityRuleServer1Name).(string)
	server2Name := data.Get(resourceKeyServerAntiAffinityRuleServer2Name).(string)
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

		// CloudControl makes no guarantees about the order in which the target servers are returned
		serversByID := make(map[string]compute.ServerSummary)
		for _, server := range antiAffinityRule.Servers {
			serversByID[server.ID] = server
		}

		server1ID := data.Get(resourceKeyServerAntiAffinityRuleServer1ID).(string)
		server1, ok := serversByID[server1ID]
		if !ok {
			return fmt.Errorf("Anti-affinity rule '%s' relates to unexpected server ('%s')", ruleID, server1ID)
		}

		server2ID := data.Get(resourceKeyServerAntiAffinityRuleServer1ID).(string)
		server2, ok := serversByID[server2ID]
		if !ok {
			return fmt.Errorf("Anti-affinity rule '%s' relates to unexpected server ('%s')", ruleID, server2ID)
		}

		data.Set(resourceKeyServerAntiAffinityRuleServer1Name, server1.Name)
		data.Set(resourceKeyServerAntiAffinityRuleServer2Name, server2.Name)
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

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Delete anti-affinity rule '%s'", ruleID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Delete server anti-affinity rule '%s'", networkDomainID)
		defer asyncLock.Release()

		deleteError := apiClient.DeleteServerAntiAffinityRule(ruleID, networkDomainID)
		if compute.IsResourceBusyError(deleteError) {
			context.Retry()
		} else if deleteError != nil {
			context.Fail(deleteError)
		}

		asyncLock.Release()
	})
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
