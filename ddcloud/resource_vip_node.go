package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyVIPNodeName                = "name"
	resourceKeyVIPNodeDescription         = "description"
	resourceKeyVIPNodeIPv4Address         = "ipv4_address"
	resourceKeyVIPNodeIPv6Address         = "ipv6_address"
	resourceKeyVIPNodeStatus              = "status"
	resourceKeyVIPNodeHealthMonitorName   = "health_monitor"
	resourceKeyVIPNodeHealthMonitorID     = "health_monitor_id"
	resourceKeyVIPNodeConnectionLimit     = "connection_limit"
	resourceKeyVIPNodeConnectionRateLimit = "connection_rate_limit"
	resourceKeyVIPNodeNetworkDomainID     = "networkdomain"
)

func resourceVIPNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceVIPNodeCreate,
		Read:   resourceVIPNodeRead,
		Exists: resourceVIPNodeExists,
		Update: resourceVIPNodeUpdate,
		Delete: resourceVIPNodeDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyVIPNodeName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A name for the VIP node",
			},
			resourceKeyVIPNodeDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the VIP node",
			},
			resourceKeyVIPNodeIPv4Address: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The VIP node's IPv4 address",
				ConflictsWith: []string{
					resourceKeyVIPNodeIPv6Address,
				},
			},
			resourceKeyVIPNodeIPv6Address: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The VIP node's IPv6 address",
				ConflictsWith: []string{
					resourceKeyVIPNodeIPv4Address,
				},
			},
			resourceKeyVIPNodeStatus: &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      compute.VIPNodeStatusEnabled,
				Description:  "The VIP node status",
				ValidateFunc: vipStatusValidator("VIP node"),
			},
			resourceKeyVIPNodeHealthMonitorName: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "CCDEFAULT.Icmp",
				Description: "The Name of the VIP node's associated health monitor (if any)",
			},
			resourceKeyVIPNodeHealthMonitorID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The Id of the VIP node's associated health monitor (if any)",
			},
			resourceKeyVIPNodeConnectionLimit: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20000,
				Description: "The number of active connections that the node supports",
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					connectionRate := data.(int)
					if connectionRate > 0 {
						return
					}

					errors = append(errors,
						fmt.Errorf("Connection rate ('%s') must be greater than 0.", fieldName),
					)

					return
				},
			},
			resourceKeyVIPNodeConnectionRateLimit: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2000,
				Description: "The number of connections per second that the node supports",
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					connectionRate := data.(int)
					if connectionRate > 0 {
						return
					}

					errors = append(errors,
						fmt.Errorf("Connection rate limit ('%s') must be greater than 0.", fieldName),
					)

					return
				},
			},
			resourceKeyVIPNodeNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the network domain in which the VIP node is created",
			},
		},
	}
}

func resourceVIPNodeCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyVIPNodeNetworkDomainID).(string)
	name := data.Get(resourceKeyVIPNodeName).(string)
	description := data.Get(resourceKeyVIPNodeDescription).(string)
	healthMonitorName := data.Get(resourceKeyVIPNodeHealthMonitorName).(string)
	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	log.Printf("Create VIP node '%s' ('%s') in network domain '%s'.", name, description, networkDomainID)
	healthMonitorID := ""
	if len(healthMonitorName) > 0 {
		log.Printf("Find Healt Monitor ID by Name '%s' in network domain '%s'.", healthMonitorName, networkDomainID)
		page := compute.DefaultPaging()
		page.PageSize = 50
		healthMonitors, err := apiClient.ListDefaultHealthMonitors(networkDomainID, page)
		if err != nil {
			return err
		}
		for _, healthMonitor := range healthMonitors.Items {
			if healthMonitor.Name == healthMonitorName {
				healthMonitorID = healthMonitor.ID
			}
		}
	}
	vipNodeID, err := apiClient.CreateVIPNode(compute.NewVIPNodeConfiguration{
		Name:                name,
		Description:         description,
		Status:              data.Get(resourceKeyVIPNodeStatus).(string),
		IPv4Address:         data.Get(resourceKeyVIPNodeIPv4Address).(string),
		IPv6Address:         data.Get(resourceKeyVIPNodeIPv6Address).(string),
		HealthMonitorID:     healthMonitorID,
		ConnectionLimit:     data.Get(resourceKeyVIPNodeConnectionLimit).(int),
		ConnectionRateLimit: data.Get(resourceKeyVIPNodeConnectionRateLimit).(int),
		NetworkDomainID:     networkDomainID,
	})
	if err != nil {
		return err
	}

	data.SetId(vipNodeID)

	log.Printf("Successfully created VIP node '%s'.", vipNodeID)

	vipNode, err := apiClient.GetVIPNode(vipNodeID)
	if err != nil {
		return err
	}

	if vipNode == nil {
		return fmt.Errorf("Cannot find newly-added VIP node '%s'.", vipNodeID)
	}

	data.Set(resourceKeyVIPNodeDescription, vipNode.Description)
	data.Set(resourceKeyVIPNodeStatus, vipNode.Status)
	data.Set(resourceKeyVIPNodeHealthMonitorID, healthMonitorID)
	data.Set(resourceKeyVIPNodeConnectionLimit, vipNode.ConnectionLimit)
	data.Set(resourceKeyVIPNodeConnectionRateLimit, vipNode.ConnectionRateLimit)

	return nil
}

func resourceVIPNodeExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()

	log.Printf("Check if VIP node '%s' exists...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	vipNode, err := apiClient.GetVIPNode(id)
	if err != nil {
		return false, err
	}

	exists := vipNode != nil

	log.Printf("VIP node '%s' exists: %t.", id, exists)

	return exists, nil
}

func resourceVIPNodeRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()

	log.Printf("Read VIP node '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	vipNode, err := apiClient.GetVIPNode(id)
	if err != nil {
		return err
	}
	if vipNode == nil {
		data.SetId("") // VIP node has been deleted

		return nil
	}

	data.Set(resourceKeyVIPNodeStatus, vipNode.Status)

	return nil
}

func resourceVIPNodeUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	log.Printf("Update VIP node '%s'...", id)
	networkDomainID := data.Get(resourceKeyVIPNodeNetworkDomainID).(string)
	configuration := &compute.EditVIPNodeConfiguration{}
	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	propertyHelper := propertyHelper(data)
	if data.HasChange(resourceKeyVIPNodeDescription) {
		configuration.Description = propertyHelper.GetOptionalString(resourceKeyVIPNodeDescription, true)
	}

	if data.HasChange(resourceKeyVIPNodeStatus) {
		configuration.Status = propertyHelper.GetOptionalString(resourceKeyVIPNodeStatus, false)
	}

	if data.HasChange(resourceKeyVIPNodeHealthMonitorName) {
		healthMonitorName := propertyHelper.GetOptionalString(resourceKeyVIPNodeHealthMonitorName, false)
		healthMonitorID := ""
		if len(*healthMonitorName) > 0 {
			page := compute.DefaultPaging()
			page.PageSize = 50
			healthMonitors, err := apiClient.ListDefaultHealthMonitors(networkDomainID, page)
			if err != nil {
				return err
			}
			for _, healthMonitor := range healthMonitors.Items {
				if healthMonitor.Name == *healthMonitorName {
					healthMonitorID = healthMonitor.ID
				}
			}
		}
		configuration.HealthMonitorID = propertyHelper.GetOptionalString(healthMonitorID, true)
	}

	if data.HasChange(resourceKeyVIPNodeConnectionLimit) {
		configuration.ConnectionLimit = propertyHelper.GetOptionalInt(resourceKeyVIPNodeConnectionLimit, false)
	}

	if data.HasChange(resourceKeyVIPNodeConnectionRateLimit) {
		configuration.ConnectionRateLimit = propertyHelper.GetOptionalInt(resourceKeyVIPNodeConnectionRateLimit, false)
	}

	return apiClient.EditVIPNode(id, *configuration)
}

func resourceVIPNodeDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVIPNodeName).(string)
	networkDomainID := data.Get(resourceKeyVIPNodeNetworkDomainID).(string)

	log.Printf("Delete VIP node '%s' ('%s') from network domain '%s'...", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	poolMemberships, err := getVIPNodePoolMemberships(apiClient, id, networkDomainID)
	if err != nil {
		return err
	}
	for _, membership := range poolMemberships {
		log.Printf("Removing membership '%s' (pool '%s') for node '%s'...", membership.ID, membership.Pool.ID, id)
		err = apiClient.RemoveVIPPoolMember(membership.ID)
		if err != nil {
			return err
		}
		log.Printf("Removed membership '%s' (pool '%s') for node '%s'.", membership.ID, membership.Pool.ID, id)
	}

	return apiClient.DeleteVIPNode(id)
}

func getVIPNodePoolMemberships(apiClient *compute.Client, nodeID string, networkDomainID string) (memberships []compute.VIPPoolMember, err error) {
	page := compute.DefaultPaging()
	page.PageSize = 50
	for {
		var members *compute.VIPPoolMembers
		members, err = apiClient.ListVIPPoolMembershipsInNetworkDomain(networkDomainID, page)
		if err != nil {
			return
		}
		if members.IsEmpty() {
			break // We're done.
		}

		for _, member := range members.Items {
			if member.Node.ID == nodeID {
				memberships = append(memberships, member)
			}
		}

		page.Next()
	}

	return
}

func vipStatusValidator(targetDescription string) schema.SchemaValidateFunc {
	return func(data interface{}, fieldName string) (messages []string, errors []error) {
		status := data.(string)
		switch status {
		case compute.VIPNodeStatusEnabled:
		case compute.VIPNodeStatusDisabled:
		case compute.VIPNodeStatusForcedOffline:
			return
		default:
			errors = append(errors, fmt.Errorf("Invalid %s status value '%s' for field '%s'.", targetDescription, status, fieldName))
		}

		return
	}
}
