package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyVIPNodeName                = "name"
	resourceKeyVIPNodeDescription         = "description"
	resourceKeyVIPNodeIPv4Address         = "ipv4_address"
	resourceKeyVIPNodeIPv6Address         = "ipv6_address"
	resourceKeyVIPNodeStatus              = "status"
	resourceKeyVIPNodeHealthMonitorID     = "health_monitor"
	resourceKeyVIPNodeConnectionLimit     = "connection_limit"
	resourceKeyVIPNodeConnectionRateLimit = "connection_rate_limit"
	resourceKeyVIPNodeNetworkDomainID     = "networkdomain"
	resourceCreateTimeoutVIPNode          = 3 * time.Minute
	resourceDeleteTimeoutVIPNode          = 3 * time.Minute
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
				Type:     schema.TypeString,
				Required: true,
			},
			resourceKeyVIPNodeDescription: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVIPNodeIPv4Address: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ConflictsWith: []string{
					resourceKeyVIPNodeIPv6Address,
				},
			},
			resourceKeyVIPNodeIPv6Address: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ConflictsWith: []string{
					resourceKeyVIPNodeIPv4Address,
				},
			},
			resourceKeyVIPNodeStatus: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  compute.VIPNodeStatusEnabled,
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					status := data.(string)
					switch status {
					case compute.VIPNodeStatusEnabled:
					case compute.VIPNodeStatusDisabled:
					case compute.VIPNodeStatusForcedOffline:
						return
					default:
						errors = append(errors, fmt.Errorf("Invalid VIP node status '%s'.", status))
					}

					return
				},
			},
			resourceKeyVIPNodeHealthMonitorID: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVIPNodeConnectionLimit: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  20000,
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
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2000,
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
		},
	}
}

func resourceVIPNodeCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyVIPNodeNetworkDomainID).(string)
	name := data.Get(resourceKeyVIPNodeName).(string)
	description := data.Get(resourceKeyVIPNodeDescription).(string)

	log.Printf("Create VIP node '%s' ('%s') in network domain '%s'.", name, description, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	vipNodeID, err := apiClient.CreateVIPNode(compute.NewVIPNodeConfiguration{
		Name:                name,
		Description:         description,
		IPv4Address:         data.Get(resourceKeyVIPNodeIPv4Address).(string),
		IPv6Address:         data.Get(resourceKeyVIPNodeIPv6Address).(string),
		HealthMonitorID:     data.Get(resourceKeyVIPNodeHealthMonitorID).(string),
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
	data.Set(resourceKeyVIPNodeHealthMonitorID, vipNode.HealthMonitorID)
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

	configuration := &compute.EditVIPNodeConfiguration{}

	propertyHelper := propertyHelper(data)
	if data.HasChange(resourceKeyVIPNodeDescription) {
		configuration.Description = propertyHelper.GetOptionalString(resourceKeyVIPNodeDescription, true)
	}

	if data.HasChange(resourceKeyVIPNodeStatus) {
		configuration.Status = propertyHelper.GetOptionalString(resourceKeyVIPNodeStatus, false)
	}

	if data.HasChange(resourceKeyVIPNodeHealthMonitorID) {
		configuration.HealthMonitorID = propertyHelper.GetOptionalString(resourceKeyVIPNodeHealthMonitorID, true)
	}

	if data.HasChange(resourceKeyVIPNodeConnectionLimit) {
		configuration.ConnectionLimit = propertyHelper.GetOptionalInt(resourceKeyVIPNodeConnectionLimit, false)
	}

	if data.HasChange(resourceKeyVIPNodeHealthMonitorID) {
		configuration.ConnectionRateLimit = propertyHelper.GetOptionalInt(resourceKeyVIPNodeConnectionRateLimit, false)
	}

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.EditVIPNode(id, *configuration)
}

func resourceVIPNodeDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVIPNodeName).(string)
	networkDomainID := data.Get(resourceKeyVIPNodeNetworkDomainID)

	log.Printf("Delete VIP node '%s' ('%s') from network domain '%s'...", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.DeleteNATRule(id)
}
