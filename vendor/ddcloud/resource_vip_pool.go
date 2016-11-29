package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyVIPPoolName               = "name"
	resourceKeyVIPPoolDescription        = "description"
	resourceKeyVIPPoolLoadBalanceMethod  = "load_balance_method"
	resourceKeyVIPPoolHealthMonitorNames = "health_monitors"
	resourceKeyVIPPoolHealthMonitorIDs   = "health_monitors_id"
	resourceKeyVIPPoolServiceDownAction  = "service_down_action"
	resourceKeyVIPPoolSlowRampTime       = "slow_ramp_time"
	resourceKeyVIPPoolNetworkDomainID    = "networkdomain"
)

func resourceVIPPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceVIPPoolCreate,
		Read:   resourceVIPPoolRead,
		Exists: resourceVIPPoolExists,
		Update: resourceVIPPoolUpdate,
		Delete: resourceVIPPoolDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyVIPPoolName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A name for the VIP pool",
			},
			resourceKeyVIPPoolDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description of the VIP pool",
			},
			resourceKeyVIPPoolLoadBalanceMethod: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     compute.LoadBalanceMethodRoundRobin,
				Description: "The load-balancing method used by the VIP pool",
			},
			resourceKeyVIPPoolHealthMonitorNames: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			resourceKeyVIPPoolHealthMonitorIDs: &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			resourceKeyVIPPoolServiceDownAction: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  compute.ServiceDownActionNone,
				ValidateFunc: func(data interface{}, fieldName string) (messages []string, errors []error) {
					status := data.(string)
					switch status {
					case compute.ServiceDownActionNone:
					case compute.ServiceDownActionDrop:
					case compute.ServiceDownActionReselect:
						return
					default:
						errors = append(errors, fmt.Errorf("Invalid VIP service-down action '%s'.", status))
					}

					return
				},
			},
			resourceKeyVIPPoolSlowRampTime: &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			resourceKeyVIPPoolNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVIPPoolCreate(data *schema.ResourceData, provider interface{}) error {
	networkDomainID := data.Get(resourceKeyVIPPoolNetworkDomainID).(string)
	name := data.Get(resourceKeyVIPPoolName).(string)
	description := data.Get(resourceKeyVIPPoolDescription).(string)

	log.Printf("Create VIP pool '%s' ('%s') in network domain '%s'.", name, description, networkDomainID)

	propertyHelper := propertyHelper(data)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	healthMonitorNames := propertyHelper.GetStringSetItems(resourceKeyVIPPoolHealthMonitorNames)
	healthMonitorIDs := make([]string, len(healthMonitorNames))
	if healthMonitorNames != nil {
		// healthMonitorIDs = []string
		page := compute.DefaultPaging()
		page.PageSize = 50
		healthMonitors, err := apiClient.ListDefaultHealthMonitors(networkDomainID, page)
		if err != nil {
			return err
		}
		for index, healthMonitorName := range healthMonitorNames {
			for _, healthMonitor := range healthMonitors.Items {
				if healthMonitorName == healthMonitor.Name {
					healthMonitorIDs[index] = healthMonitor.ID
				}
			}
		}
	}
	vipPoolID, err := apiClient.CreateVIPPool(compute.NewVIPPoolConfiguration{
		Name:              name,
		Description:       description,
		LoadBalanceMethod: data.Get(resourceKeyVIPPoolLoadBalanceMethod).(string),
		HealthMonitorIDs:  healthMonitorIDs,
		ServiceDownAction: data.Get(resourceKeyVIPPoolServiceDownAction).(string),
		SlowRampTime:      data.Get(resourceKeyVIPPoolSlowRampTime).(int),
		NetworkDomainID:   networkDomainID,
	})
	if err != nil {
		return err
	}

	data.SetId(vipPoolID)
	propertyHelper.SetStringSetItems(resourceKeyVIPPoolHealthMonitorIDs, healthMonitorIDs)
	log.Printf("Successfully created VIP pool '%s'.", vipPoolID)

	return nil
}

func resourceVIPPoolExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()

	log.Printf("Check if VIP pool '%s' exists...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	vipPool, err := apiClient.GetVIPPool(id)
	if err != nil {
		return false, err
	}

	exists := vipPool != nil

	log.Printf("VIP pool '%s' exists: %t.", id, exists)

	return exists, nil
}

func resourceVIPPoolRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()

	log.Printf("Read VIP pool '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	vipPool, err := apiClient.GetVIPPool(id)
	if err != nil {
		return err
	}
	if vipPool == nil {
		data.SetId("") // VIP pool has been deleted

		return nil
	}

	data.Set(resourceKeyVIPPoolName, vipPool.Name)
	data.Set(resourceKeyVIPPoolDescription, vipPool.Description)
	data.Set(resourceKeyVIPPoolLoadBalanceMethod, vipPool.LoadBalanceMethod)

	propertyHelper := propertyHelper(data)

	healthMonitorIDs := make([]string, len(vipPool.HealthMonitors))
	for index, healthMonitor := range vipPool.HealthMonitors {
		healthMonitorIDs[index] = healthMonitor.ID
	}
	propertyHelper.SetStringSetItems(resourceKeyVIPPoolHealthMonitorIDs, healthMonitorIDs)

	return nil
}

func resourceVIPPoolUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	log.Printf("Update VIP pool '%s'...", id)
	configuration := &compute.EditVIPPoolConfiguration{}
	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	propertyHelper := propertyHelper(data)
	if data.HasChange(resourceKeyVIPPoolDescription) {
		configuration.Description = propertyHelper.GetOptionalString(resourceKeyVIPPoolDescription, true)
	}

	if data.HasChange(resourceKeyVIPPoolLoadBalanceMethod) {
		configuration.LoadBalanceMethod = propertyHelper.GetOptionalString(resourceKeyVIPPoolLoadBalanceMethod, false)
	}

	if data.HasChange(resourceKeyVIPPoolHealthMonitorNames) {
		networkDomainID := data.Get(resourceKeyVIPPoolNetworkDomainID).(string)
		healthMonitorNames := propertyHelper.GetStringSetItems(resourceKeyVIPPoolHealthMonitorNames)
		healthMonitorIDs := make([]string, len(healthMonitorNames))
		if healthMonitorNames != nil {
			page := compute.DefaultPaging()
			page.PageSize = 50
			healthMonitors, err := apiClient.ListDefaultHealthMonitors(networkDomainID, page)
			if err != nil {
				return err
			}
			for index, healthMonitorName := range healthMonitorNames {
				for _, healthMonitor := range healthMonitors.Items {
					if healthMonitorName == healthMonitor.Name {
						healthMonitorIDs[index] = healthMonitor.ID
					}
				}
			}
		}
		configuration.HealthMonitorIDs = &healthMonitorIDs
	}

	if data.HasChange(resourceKeyVIPPoolServiceDownAction) {
		configuration.ServiceDownAction = propertyHelper.GetOptionalString(resourceKeyVIPPoolServiceDownAction, false)
	}

	if data.HasChange(resourceKeyVIPPoolSlowRampTime) {
		configuration.SlowRampTime = propertyHelper.GetOptionalInt(resourceKeyVIPPoolSlowRampTime, false)
	}

	return apiClient.EditVIPPool(id, *configuration)
}

func resourceVIPPoolDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyVIPPoolName).(string)
	networkDomainID := data.Get(resourceKeyVIPPoolNetworkDomainID)

	log.Printf("Delete VIP pool '%s' ('%s') from network domain '%s'...", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.DeleteVIPPool(id)
}
