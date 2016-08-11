package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyVirtualListenerName            = "name"
	resourceKeyVirtualListenerDescription     = "description"
	resourceKeyVirtualListenerType            = "type"
	resourceKeyVirtualListenerProtocol        = "protocol"
	resourceKeyVirtualListenerIPv4Address     = "ipv4_address"
	resourceKeyVirtualListenerPort            = "port"
	resourceKeyVirtualListenerEnabled         = "enabled"
	resourceKeyVirtualListenerNetworkDomainID = "networkdomain"
)

func resourceVirtualListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualListenerCreate,
		Read:   resourceVirtualListenerRead,
		Exists: resourceVirtualListenerExists,
		Update: resourceVirtualListenerUpdate,
		Delete: resourceVirtualListenerDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyVirtualListenerName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			resourceKeyVirtualListenerDescription: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVirtualListenerType: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  compute.VirtualListenerTypeStandard,
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					listenerType := data.(string)
					switch listenerType {
					case compute.VirtualListenerTypeStandard:
					case compute.VirtualListenerTypePerformanceLayer4:
						return
					default:
						errors = append(errors, fmt.Errorf("Invalid virtual listener type '%s'.", listenerType))
					}

					return
				},
			},
			resourceKeyVirtualListenerProtocol: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			resourceKeyVirtualListenerIPv4Address: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVirtualListenerPort: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ForceNew: true,
			},
			resourceKeyVirtualListenerEnabled: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// TODO: Add remaining properties.
		},
	}
}

func resourceVirtualListenerCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID).(string)
	name := data.Get(resourceKeyVirtualListenerName).(string)
	description := data.Get(resourceKeyVirtualListenerDescription).(string)

	log.Printf("Create virtual listener '%s' ('%s') in network domain '%s'.", name, description, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	virtualListenerID, err := apiClient.CreateVirtualListener(compute.NewVirtualListenerConfiguration{
		Name:              name,
		Description:       description,
		Type:              data.Get(resourceKeyVirtualListenerType).(string),
		Protocol:          data.Get(resourceKeyVirtualListenerProtocol).(string),
		Port:              data.Get(resourceKeyVirtualListenerPort).(int),
		ListenerIPAddress: data.Get(resourceKeyVirtualListenerIPv4Address).(string),
		Enabled:           data.Get(resourceKeyVirtualListenerEnabled).(bool),
		NetworkDomainID:   networkDomainID,
	})
	if err != nil {
		return err
	}

	data.SetId(virtualListenerID)

	log.Printf("Successfully created virtual listener '%s'.", virtualListenerID)

	return nil
}

func resourceVirtualListenerExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()

	log.Printf("Check if virtual listener '%s' exists...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	vipPool, err := apiClient.GetVirtualListener(id)
	if err != nil {
		return false, err
	}

	exists := vipPool != nil

	log.Printf("virtual listener '%s' exists: %t.", id, exists)

	return exists, nil
}

func resourceVirtualListenerRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()

	log.Printf("Read virtual listener '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	virtualListener, err := apiClient.GetVirtualListener(id)
	if err != nil {
		return err
	}
	if virtualListener == nil {
		data.SetId("") // Virtual listener has been deleted

		return nil
	}

	data.Set(resourceKeyVirtualListenerDescription, virtualListener.Description)
	data.Set(resourceKeyVirtualListenerEnabled, virtualListener.Enabled)

	// TODO: Capture other properties.

	return nil
}

func resourceVirtualListenerUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	log.Printf("Update virtual listener '%s'...", id)

	configuration := &compute.EditVirtualListenerConfiguration{}

	propertyHelper := propertyHelper(data)
	if data.HasChange(resourceKeyVirtualListenerDescription) {
		configuration.Description = propertyHelper.GetOptionalString(resourceKeyVirtualListenerDescription, true)
	}

	if data.HasChange(resourceKeyVirtualListenerEnabled) {
		configuration.Enabled = propertyHelper.GetOptionalBool(resourceKeyVirtualListenerEnabled)
	}

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.EditVirtualListener(id, *configuration)
}

func resourceVirtualListenerDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVirtualListenerName).(string)
	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID)

	log.Printf("Delete virtual listener '%s' ('%s') from network domain '%s'...", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.DeleteVirtualListener(id)
}
