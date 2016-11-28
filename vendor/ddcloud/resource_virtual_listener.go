package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyVirtualListenerName                   = "name"
	resourceKeyVirtualListenerDescription            = "description"
	resourceKeyVirtualListenerType                   = "type"
	resourceKeyVirtualListenerProtocol               = "protocol"
	resourceKeyVirtualListenerIPv4Address            = "ipv4"
	resourceKeyVirtualListenerPort                   = "port"
	resourceKeyVirtualListenerEnabled                = "enabled"
	resourceKeyVirtualListenerConnectionLimit        = "connection_limit"
	resourceKeyVirtualListenerConnectionRateLimit    = "connection_rate_limit"
	resourceKeyVirtualListenerSourcePortPreservation = "source_port_preservation"
	resourceKeyVirtualListenerPoolID                 = "pool"
	resourceKeyVirtualListenerPersistenceProfileName = "persistence_profile"
	resourceKeyVirtualListenerIRuleNames             = "irules"
	resourceKeyVirtualListenerOptimizationProfiles   = "optimization_profiles"
	resourceKeyVirtualListenerNetworkDomainID        = "networkdomain"
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
				Computed: true,
				Default:  nil,
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
			resourceKeyVirtualListenerConnectionLimit: &schema.Schema{
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
			resourceKeyVirtualListenerConnectionRateLimit: &schema.Schema{
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
			resourceKeyVirtualListenerSourcePortPreservation: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  compute.SourcePortPreservationEnabled,
			},
			resourceKeyVirtualListenerPoolID: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			resourceKeyVirtualListenerPersistenceProfileName: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyVirtualListenerIRuleNames: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: func(item interface{}) int {
					iRuleID := item.(string)

					return schema.HashString(iRuleID)
				},
			},
			resourceKeyVirtualListenerOptimizationProfiles: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: func(item interface{}) int {
					optimizationProfile := item.(string)

					return schema.HashString(optimizationProfile)
				},
			},
			resourceKeyVirtualListenerNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// TODO: Add remaining properties.
		},
	}
}

func resourceVirtualListenerCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID).(string)
	name := data.Get(resourceKeyVirtualListenerName).(string)
	description := data.Get(resourceKeyVirtualListenerDescription).(string)
	listenerIPAddress := data.Get(resourceKeyVirtualListenerIPv4Address).(string)

	log.Printf("Create virtual listener '%s' ('%s') in network domain '%s'.", name, description, networkDomainID)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	propertyHelper := propertyHelper(data)

	var virtualListenerID string

	operationDescription := fmt.Sprintf("Create virtual listener '%s' ", name)
	operationError := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// Map from names to Ids, as required.
		persistenceProfileID, err := propertyHelper.GetVirtualListenerPersistenceProfileID(apiClient)
		if err != nil {
			context.Fail(err)

			return
		}

		iRuleIDs, err := propertyHelper.GetVirtualListenerIRuleIDs(apiClient)
		if err != nil {
			context.Fail(err)

			return
		}

		if len(listenerIPAddress) == 0 {
			var freeIPs map[string]string
			freeIPs, err = apiClient.GetAvailablePublicIPAddresses(networkDomainID)
			if err != nil {
				context.Fail(err)
			}

			if len(freeIPs) == 0 {
				log.Printf("There are no free public IPv4 addresses in network domain '%s'; requesting allocation of a new address block...", networkDomainID)

				// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
				asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
				defer asyncLock.Release() // Released at the end of the current attempt.

				var blockID string
				blockID, err = apiClient.AddPublicIPBlock(networkDomainID)
				if err != nil {
					if compute.IsResourceBusyError(err) {
						context.Retry()
					} else {
						context.Fail(err)
					}

					return
				}

				asyncLock.Release()

				var block *compute.PublicIPBlock
				block, err = apiClient.GetPublicIPBlock(blockID)
				if err != nil {
					context.Fail(err)

					return
				}

				if block == nil {
					context.Fail(
						fmt.Errorf("Cannot find newly-added public IPv4 address block '%s'.", blockID),
					)

					return
				}

				log.Printf("Allocated a new public IPv4 address block '%s' (%d addresses, starting at '%s').", block.ID, block.Size, block.BaseIP)
			}

		}

		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		virtualListenerID, err = apiClient.CreateVirtualListener(compute.NewVirtualListenerConfiguration{
			Name:                   name,
			Description:            description,
			Type:                   data.Get(resourceKeyVirtualListenerType).(string),
			Protocol:               data.Get(resourceKeyVirtualListenerProtocol).(string),
			Port:                   data.Get(resourceKeyVirtualListenerPort).(int),
			ListenerIPAddress:      propertyHelper.GetOptionalString(resourceKeyVirtualListenerIPv4Address, false),
			Enabled:                data.Get(resourceKeyVirtualListenerEnabled).(bool),
			ConnectionLimit:        data.Get(resourceKeyVirtualListenerConnectionLimit).(int),
			ConnectionRateLimit:    data.Get(resourceKeyVirtualListenerConnectionRateLimit).(int),
			SourcePortPreservation: data.Get(resourceKeyVirtualListenerSourcePortPreservation).(string),
			PoolID:                 propertyHelper.GetOptionalString(resourceKeyVirtualListenerPoolID, false),
			PersistenceProfileID:   persistenceProfileID,
			IRuleIDs:               iRuleIDs,
			OptimizationProfiles:   propertyHelper.GetStringSetItems(resourceKeyVirtualListenerOptimizationProfiles),
			NetworkDomainID:        networkDomainID,
		})
		if err != nil {
			if compute.IsResourceBusyError(err) {
				context.Retry()
			} else {
				context.Fail(err)
			}
		}

		asyncLock.Release()
	})
	if operationError != nil {
		return operationError
	}

	data.SetId(virtualListenerID)

	log.Printf("Successfully created virtual listener '%s'.", virtualListenerID)

	virtualListener, err := apiClient.GetVirtualListener(virtualListenerID)
	if err != nil {
		return err
	}
	if virtualListener == nil {
		return fmt.Errorf("Cannot find newly-created virtual listener with Id '%s'.", virtualListenerID)
	}

	data.Set(resourceKeyVirtualListenerIPv4Address, virtualListener.ListenerIPAddress)

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
	data.Set(resourceKeyVirtualListenerConnectionLimit, virtualListener.ConnectionLimit)
	data.Set(resourceKeyVirtualListenerConnectionRateLimit, virtualListener.ConnectionRateLimit)
	data.Set(resourceKeyVirtualListenerSourcePortPreservation, virtualListener.SourcePortPreservation)
	data.Set(resourceKeyVirtualListenerPersistenceProfileName, virtualListener.PersistenceProfile.Name)
	data.Set(resourceKeyVirtualListenerIPv4Address, virtualListener.ListenerIPAddress)

	propertyHelper := propertyHelper(data)
	propertyHelper.SetVirtualListenerIRules(virtualListener.IRules)

	// TODO: Capture other properties.

	return nil
}

func resourceVirtualListenerUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	log.Printf("Update virtual listener '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	configuration := &compute.EditVirtualListenerConfiguration{}

	propertyHelper := propertyHelper(data)
	if data.HasChange(resourceKeyVirtualListenerDescription) {
		configuration.Description = propertyHelper.GetOptionalString(resourceKeyVirtualListenerDescription, true)
	}

	if data.HasChange(resourceKeyVirtualListenerEnabled) {
		configuration.Enabled = propertyHelper.GetOptionalBool(resourceKeyVirtualListenerEnabled)
	}

	if data.HasChange(resourceKeyVirtualListenerConnectionLimit) {
		configuration.ConnectionLimit = propertyHelper.GetOptionalInt(resourceKeyVirtualListenerConnectionLimit, false)
	}

	if data.HasChange(resourceKeyVirtualListenerConnectionRateLimit) {
		configuration.ConnectionRateLimit = propertyHelper.GetOptionalInt(resourceKeyVirtualListenerConnectionRateLimit, false)
	}

	if data.HasChange(resourceKeyVirtualListenerSourcePortPreservation) {
		configuration.SourcePortPreservation = propertyHelper.GetOptionalString(resourceKeyVirtualListenerSourcePortPreservation, true)
	}

	if data.HasChange(resourceKeyVirtualListenerPoolID) {
		configuration.PoolID = propertyHelper.GetOptionalString(resourceKeyVirtualListenerPoolID, true)
	}

	if data.HasChange(resourceKeyVirtualListenerPersistenceProfileName) {
		persistenceProfile, err := propertyHelper.GetVirtualListenerPersistenceProfile(apiClient)
		if err != nil {
			return err
		}

		configuration.PersistenceProfileID = &persistenceProfile.ID
	}

	if data.HasChange(resourceKeyVirtualListenerIRuleNames) {
		iRuleIDs, err := propertyHelper.GetVirtualListenerIRuleIDs(apiClient)
		if err != nil {
			return err
		}

		configuration.IRuleIDs = &iRuleIDs
	}

	return apiClient.EditVirtualListener(id, *configuration)
}

func resourceVirtualListenerDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVirtualListenerName).(string)
	networkDomainID := data.Get(resourceKeyVirtualListenerNetworkDomainID)

	log.Printf("Delete virtual listener '%s' ('%s') from network domain '%s'...", name, id, networkDomainID)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Delete virtual listener '%s", id)

	return providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		err := apiClient.DeleteVirtualListener(id)
		if err != nil {
			if compute.IsResourceBusyError(err) {
				context.Retry()
			} else {
				context.Fail(err)
			}
		}

		asyncLock.Release()
	})
}
