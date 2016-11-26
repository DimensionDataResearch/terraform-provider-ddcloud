package ddcloud

import (
	"log"
	"strings"
	"time"

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

	log.Printf("Create network domain '%s' in data center '%s' (plan = '%s', description = '%s').", name, dataCenterID, plan, description)

	// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
	asyncLock := providerState.AcquireAsyncOperationLock("Create network domain '%s'", name)
	defer asyncLock.Release()

	// TODO: Handle RESOURCE_BUSY response (retry?)
	apiClient := providerState.Client()
	networkDomainID, err := apiClient.DeployNetworkDomain(name, description, plan, dataCenterID)
	if err != nil {
		return err
	}

	// Operation initiated; we don't need the lock anymore.
	asyncLock.Release()

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
		// TODO: Handle RESOURCE_BUSY response (retry?)
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
	var err error

	networkDomainID := data.Id()
	name := data.Get(resourceKeyNetworkDomainName).(string)
	dataCenterID := data.Get(resourceKeyNetworkDomainDataCenter).(string)

	log.Printf("Delete network domain '%s' ('%s') in data center '%s'.", networkDomainID, name, dataCenterID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	// First, check if the network domain has any allocated public IP blocks.
	page := compute.DefaultPaging()
	for {
		var publicIPBlocks *compute.PublicIPBlocks
		publicIPBlocks, err = apiClient.ListPublicIPBlocks(networkDomainID, page)
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

	// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
	asyncLock := providerState.AcquireAsyncOperationLock("Delete network domain '%s'", networkDomainID)
	defer asyncLock.Release()

	// TODO: Handle RESOURCE_BUSY response (retry?)
	err = apiClient.DeleteNetworkDomain(networkDomainID)
	if err != nil {
		return err
	}

	// Operation initiated; we don't need the lock anymore.
	asyncLock.Release()

	log.Printf("Network domain '%s' is being deleted...", networkDomainID)

	return apiClient.WaitForDelete(compute.ResourceTypeNetworkDomain, networkDomainID, resourceDeleteTimeoutServer)
}
