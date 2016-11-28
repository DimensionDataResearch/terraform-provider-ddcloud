package ddcloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyNetworkDomainName           = "name"
	resourceKeyNetworkDomainDescription    = "description"
	resourceKeyNetworkDomainPlan           = "plan"
	resourceKeyNetworkDomainDataCenter     = "datacenter"
	resourceKeyNetworkDomainNatIPv4Address = "nat_ipv4_address"
	resourceKeyNetworkDomainFirewallRule   = "default_firewall_rule"
	resourceCreateTimeoutNetworkDomain     = 5 * time.Minute
	resourceDeleteTimeoutNetworkDomain     = 5 * time.Minute
)

func resourceNetworkDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkDomainCreate,
		Read:   resourceNetworkDomainRead,
		Update: resourceNetworkDomainUpdate,
		Delete: resourceNetworkDomainDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNetworkDomainName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the network domain",
			},
			resourceKeyNetworkDomainDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the network domain",
			},
			resourceKeyNetworkDomainPlan: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "ESSENTIALS",
				Description: "The plan (service level) for the network domain (ESSENTIALS or ADVANCED)",
				StateFunc: func(value interface{}) string {
					plan := value.(string)

					return strings.ToUpper(plan)
				},
			},
			resourceKeyNetworkDomainDataCenter: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the MCP 2.0 datacenter in which the network domain is created",
			},
			resourceKeyNetworkDomainNatIPv4Address: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv4 address for the network domain's IPv6->IPv4 Source Network Address Translation (SNAT). This is the IPv4 address of the network domain's IPv4 egress",
			},
			resourceKeyNetworkDomainFirewallRule: schemaNetworkDomainFirewallRule(),
		},
	}
}

// Create a network domain resource.
func resourceNetworkDomainCreate(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	name = data.Get(resourceKeyNetworkDomainName).(string)
	description = data.Get(resourceKeyNetworkDomainDescription).(string)
	plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	dataCenterID = data.Get(resourceKeyNetworkDomainDataCenter).(string)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	log.Printf("Create network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	var networkDomainID string
	operationDescription := fmt.Sprintf("Create network domain '%s'", name)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Create network domain '%s'", name)
		defer asyncLock.Release()

		var deployError error
		networkDomainID, deployError = apiClient.DeployNetworkDomain(name, description, plan, dataCenterID)
		if compute.IsResourceBusyError(deployError) {
			context.Retry()
		} else if deployError != nil {
			context.Fail(deployError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	data.SetId(networkDomainID)

	log.Printf("Network domain '%s' is being provisioned...", networkDomainID)

	resource, err := apiClient.WaitForDeploy(compute.ResourceTypeNetworkDomain, networkDomainID, resourceCreateTimeoutVLAN)
	if err != nil {
		return err
	}

	data.Partial(true)

	// Capture additional properties that are only available after deployment.
	networkDomain := resource.(*compute.NetworkDomain)
	data.Set(resourceKeyNetworkDomainNatIPv4Address, networkDomain.NatIPv4Address)
	data.SetPartial(resourceKeyNetworkDomainNatIPv4Address)

	err = applyNetworkDomainDefaultFirewallRules(data, apiClient)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

// Read a network domain resource.
func resourceNetworkDomainRead(data *schema.ResourceData, provider interface{}) error {
	var name, description, plan, dataCenterID string

	id := data.Id()
	name = data.Get(resourceKeyNetworkDomainName).(string)
	description = data.Get(resourceKeyNetworkDomainDescription).(string)
	plan = data.Get(resourceKeyNetworkDomainPlan).(string)
	dataCenterID = data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Read network domain '%s' (Id = '%s') in data center '%s' (plan = '%s', description = '%s').", name, id, dataCenterID, plan, description)

	apiClient := provider.(*providerState).Client()

	networkDomain, err := apiClient.GetNetworkDomain(id)
	if err != nil {
		return err
	}

	data.Partial(true)

	if networkDomain != nil {
		data.Set(resourceKeyNetworkDomainName, networkDomain.Name)
		data.SetPartial(resourceKeyNetworkDomainName)
		data.Set(resourceKeyNetworkDomainDescription, networkDomain.Description)
		data.SetPartial(resourceKeyNetworkDomainDescription)
		data.Set(resourceKeyNetworkDomainPlan, networkDomain.Type)
		data.SetPartial(resourceKeyNetworkDomainPlan)
		data.Set(resourceKeyNetworkDomainDataCenter, networkDomain.DatacenterID)
		data.SetPartial(resourceKeyNetworkDomainDataCenter)
		data.Set(resourceKeyNetworkDomainNatIPv4Address, networkDomain.NatIPv4Address)
		data.SetPartial(resourceKeyNetworkDomainNatIPv4Address)
	} else {
		data.SetId("") // Mark resource as deleted.
	}

	err = readNetworkDomainDefaultFirewallRules(data, apiClient)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

// Update a network domain resource.
func resourceNetworkDomainUpdate(data *schema.ResourceData, provider interface{}) error {
	var (
		id, name, description, plan      string
		newName, newDescription, newPlan *string
	)

	id = data.Id()

	if data.HasChange(resourceKeyNetworkDomainName) {
		name = data.Get(resourceKeyNetworkDomainName).(string)
		newName = &name
	}

	if data.HasChange(resourceKeyNetworkDomainDescription) {
		description = data.Get(resourceKeyNetworkDomainDescription).(string)
		newDescription = &description
	}

	if data.HasChange(resourceKeyNetworkDomainPlan) {
		plan = data.Get(resourceKeyNetworkDomainPlan).(string)
		newPlan = &plan
	}

	log.Printf("Update network domain '%s' (Name = '%s', Description = '%s', Plan = '%s').", data.Id(), name, description, plan)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	var err error
	if newName != nil || newPlan != nil {
		err = apiClient.EditNetworkDomain(id, newName, newDescription, newPlan)
		if err != nil {
			return err
		}
	}

	err = applyNetworkDomainDefaultFirewallRules(data, apiClient)
	if err != nil {
		return err
	}

	return nil
}

// Delete a network domain resource.
func resourceNetworkDomainDelete(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Id()
	name := data.Get(resourceKeyNetworkDomainName).(string)
	dataCenterID := data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Delete network domain '%s' ('%s') in data center '%s'.", networkDomainID, name, dataCenterID)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	err := deleteAllPublicIPBlocks(networkDomainID, providerState)
	if err != nil {
		return err
	}

	operationDescription := fmt.Sprintf("Create network domain '%s'", name)
	err = providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Delete network domain '%s'", networkDomainID)
		defer asyncLock.Release()

		deleteError := apiClient.DeleteNetworkDomain(networkDomainID)
		if compute.IsResourceBusyError(deleteError) {
			context.Retry()
		} else if err != nil {
			context.Fail(deleteError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	log.Printf("Network domain '%s' is being deleted...", networkDomainID)

	return apiClient.WaitForDelete(compute.ResourceTypeNetworkDomain, networkDomainID, resourceDeleteTimeoutServer)
}

// Delete all public IP blocks (if any) in a network domain.
func deleteAllPublicIPBlocks(networkDomainID string, providerState *providerState) error {
	apiClient := providerState.Client()

	page := compute.DefaultPaging()
	for {
		var publicIPBlocks *compute.PublicIPBlocks
		publicIPBlocks, err := apiClient.ListPublicIPBlocks(networkDomainID, page)
		if err != nil {
			return err
		}
		if publicIPBlocks.IsEmpty() {
			break // We're done
		}

		for _, block := range publicIPBlocks.Blocks {
			log.Printf("Removing public IP block '%s' (%s+%d) from network domain '%s'...", block.ID, block.BaseIP, block.Size, networkDomainID)

			err := apiClient.RemovePublicIPBlock(block.ID)
			if err != nil {
				return err
			}

			log.Printf("Successfully deleted public IP block '%s' from network domain '%s'.", block.ID, networkDomainID)
		}

		page.Next()
	}

	return nil
}
