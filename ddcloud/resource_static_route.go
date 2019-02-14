package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyStaticRouteName                      = "name"
	resourceKeyStaticRouteDescription               = "description"
	resourceKeyStaticRouteNetworkdomain             = "networkdomain"
	resourceKeyStaticRouteIpVersion                 = "ip_version"
	resourceKeyStaticRouteDestinationNetworkAddress = "destination_network_address"
	resourceKeyStaticRouteDestinationPrefixSize     = "destination_prefix_size"
	resourceKeyStaticRouteNextHopAddress            = "next_hop_address"
	resourceKeyStaticRouteState                     = "state"
	resourceKeyStaticRouteDataCenter                = "data_center"
)

func resourceStaticRoute() *schema.Resource {

	return &schema.Resource{
		Exists: resourceStaticRouteExists,
		Create: resourceStaticRouteCreate,
		Read:   resourceStaticRouteRead,
		Update: resourceStaticRouteUpdate,
		Delete: resourceStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStaticRouteImport,
		},
		Schema: map[string]*schema.Schema{
			resourceKeyStaticRouteName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for static route",
			},
			resourceKeyStaticRouteDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description for static route",
			},
			resourceKeyStaticRouteNetworkdomain: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network domain ID",
			},
			resourceKeyStaticRouteIpVersion: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Either IPv4 or IPv6",
			},
			resourceKeyStaticRouteDestinationNetworkAddress: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Destination address",
			},
			resourceKeyStaticRouteDestinationPrefixSize: &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Integer prefix defining the size of the network",
			},
			resourceKeyStaticRouteNextHopAddress: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Gateway address in the form of an INET gateway, CPNC gateway or an address on an Attached VLAN in the same Network Domain",
			},
			resourceKeyStaticRouteState: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "State of Static Route",
			},
			resourceKeyStaticRouteDataCenter: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data center where Static Route reside",
			},
		},
	}
}

func resourceStaticRouteExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	log.Println("resourceStaticRouteExists")

	client := provider.(*providerState).Client()
	id := data.Id()
	staticRoute, _ := client.GetStaticRoute(id)
	exists := staticRoute != nil

	return exists, nil
}

func resourceStaticRouteCreate(data *schema.ResourceData, provider interface{}) error {
	log.Println("resourceStaticRouteCreate")
	var networkDomainId, name, description, ipVersion, destinationNetworkAddress, nextHopAddress string
	var destinationPrefixSize int

	networkDomainId = data.Get(resourceKeyStaticRouteNetworkdomain).(string)
	name = data.Get(resourceKeyStaticRouteName).(string)
	description = data.Get(resourceKeyStaticRouteDescription).(string)
	ipVersion = data.Get(resourceKeyStaticRouteIpVersion).(string)
	destinationNetworkAddress = data.Get(resourceKeyStaticRouteDestinationNetworkAddress).(string)
	destinationPrefixSize = data.Get(resourceKeyStaticRouteDestinationPrefixSize).(int)
	nextHopAddress = data.Get(resourceKeyStaticRouteNextHopAddress).(string)

	log.Printf("Create static route with networkDomainId:%s, name: %s, description: %s, ipVersion: %s, "+
		"destinationNetworkAddress: %s, destinationPrefixSize: %d, nextHopAddress: %s",
		networkDomainId, name, description, ipVersion, destinationNetworkAddress, destinationPrefixSize, nextHopAddress)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	var staticRouteID string
	operationDescription := fmt.Sprintf("Create static route '%s'", name)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Create static route '%s'", name)
		defer asyncLock.Release()
		log.Println("resourceStaticRouteCreate defer  asyncLock.Release")
		var deployError error
		staticRouteID, deployError = apiClient.CreateStaticRoute(networkDomainId, name, description, ipVersion,
			destinationNetworkAddress, destinationPrefixSize, nextHopAddress)
		log.Println("resourceStaticRouteCreate defer  apiClient.CreateStaticRoute")
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

	data.SetId(staticRouteID)

	log.Printf("Static Route '%s' is being provisioned...", staticRouteID)

	resource, err := apiClient.WaitForDeploy(compute.ResourceTypeStaticRoutes, staticRouteID, resourceCreateTimeoutVLAN)
	if err != nil {
		return err
	}

	data.Partial(true)

	// Capture additional properties that are only available after deployment.
	staticRoute := resource.(*compute.StaticRoute)

	data.Set(resourceKeyStaticRouteState, staticRoute.State)
	data.SetPartial(resourceKeyStaticRouteState)

	data.Set(resourceKeyStaticRouteDataCenter, staticRoute.DataCenter)
	data.SetPartial(resourceKeyStaticRouteDataCenter)

	err = applyNetworkDomainDefaultFirewallRules(data, apiClient)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

func resourceStaticRouteRead(data *schema.ResourceData, provider interface{}) error {
	log.Println("resourceStaticRouteRead")

	var networkDomainId, name, description, ipVersion, destinationNetworkAddress, nextHopAddress string
	var destinationPrefixSize int

	id := data.Id()
	networkDomainId = data.Get(resourceKeyStaticRouteNetworkdomain).(string)
	name = data.Get(resourceKeyStaticRouteName).(string)
	description = data.Get(resourceKeyStaticRouteDescription).(string)
	ipVersion = data.Get(resourceKeyStaticRouteIpVersion).(string)
	destinationNetworkAddress = data.Get(resourceKeyStaticRouteDestinationNetworkAddress).(string)
	destinationPrefixSize = data.Get(resourceKeyStaticRouteDestinationPrefixSize).(int)
	nextHopAddress = data.Get(resourceKeyStaticRouteNextHopAddress).(string)

	log.Printf("Reading static route. networkDomainId:%s, name: %s, description: %s, ipVersion: %s, "+
		"destinationNetworkAddress: %s, destinationPrefixSize: %d, nextHopAddress: %s",
		networkDomainId, name, description, ipVersion, destinationNetworkAddress, destinationPrefixSize, nextHopAddress)

	apiClient := provider.(*providerState).Client()

	staticRoute, err := apiClient.GetStaticRoute(id)

	if err != nil {
		return err
	}

	data.Partial(true)

	if staticRoute != nil {
		data.Set(resourceKeyStaticRouteNetworkdomain, staticRoute.NetworkDomainId)
		data.SetPartial(resourceKeyStaticRouteNetworkdomain)

		data.Set(resourceKeyStaticRouteName, staticRoute.Name)
		data.SetPartial(resourceKeyStaticRouteName)

		data.Set(resourceKeyStaticRouteDescription, staticRoute.Description)
		data.SetPartial(resourceKeyStaticRouteDescription)

		data.Set(resourceKeyStaticRouteIpVersion, staticRoute.IpVersion)
		data.SetPartial(resourceKeyStaticRouteIpVersion)

		data.Set(resourceKeyStaticRouteDestinationNetworkAddress, staticRoute.DestinationNetworkAddress)
		data.SetPartial(resourceKeyStaticRouteDestinationNetworkAddress)

		data.Set(resourceKeyStaticRouteDestinationPrefixSize, staticRoute.DestinationPrefixSize)
		data.SetPartial(resourceKeyStaticRouteDestinationPrefixSize)

		data.Set(resourceKeyStaticRouteNextHopAddress, staticRoute.NextHopAddress)
		data.SetPartial(resourceKeyStaticRouteNextHopAddress)
	} else {
		data.SetId("") // Mark resource as deleted
	}

	data.Partial(false)
	return nil
}

func resourceStaticRouteUpdate(data *schema.ResourceData, provider interface{}) error {
	log.Println("resourceStaticRouteUpdate")

	// update step 1: delete
	log.Println("resourceStaticRouteUpdate: delete resource")
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	id := data.Id()
	name := data.Get(resourceKeyStaticRouteName).(string)

	operationDescriptionDel := fmt.Sprintf("Deleting static route '%s'", name)

	var errDel error
	errDel = providerState.RetryAction(operationDescriptionDel, func(context retry.Context) {

		asyncLock := providerState.AcquireAsyncOperationLock("Delete static route: '%s'", id)
		defer asyncLock.Release()

		deleteErr := apiClient.DeleteStaticRoute(id)

		if compute.IsResourceBusyError(deleteErr) {
			context.Retry()
		} else if errDel != nil {
			context.Fail(deleteErr)
		}

		asyncLock.Release()
	})

	if errDel != nil {
		return errDel
	}

	// Update step 2: re-create
	log.Println("resourceStaticRouteUpdate: Re-Create resource")
	var networkDomainId, description, ipVersion, destinationNetworkAddress, nextHopAddress string
	var destinationPrefixSize int

	networkDomainId = data.Get(resourceKeyStaticRouteNetworkdomain).(string)
	name = data.Get(resourceKeyStaticRouteName).(string)
	description = data.Get(resourceKeyStaticRouteDescription).(string)
	ipVersion = data.Get(resourceKeyStaticRouteIpVersion).(string)
	destinationNetworkAddress = data.Get(resourceKeyStaticRouteDestinationNetworkAddress).(string)
	destinationPrefixSize = data.Get(resourceKeyStaticRouteDestinationPrefixSize).(int)
	nextHopAddress = data.Get(resourceKeyStaticRouteNextHopAddress).(string)

	log.Printf("Create static route with networkDomainId:%s, name: %s, description: %s, ipVersion: %s, "+
		"destinationNetworkAddress: %s, destinationPrefixSize: %d, nextHopAddress: %s",
		networkDomainId, name, description, ipVersion, destinationNetworkAddress, destinationPrefixSize, nextHopAddress)

	var staticRouteID string
	operationDescription := fmt.Sprintf("Create static route '%s'", name)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Create static route '%s'", name)
		defer asyncLock.Release()
		log.Println("resourceStaticRouteCreate defer  asyncLock.Release")
		var deployError error
		staticRouteID, deployError = apiClient.CreateStaticRoute(networkDomainId, name, description, ipVersion,
			destinationNetworkAddress, destinationPrefixSize, nextHopAddress)
		log.Println("resourceStaticRouteCreate defer  apiClient.CreateStaticRoute")
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

	data.SetId(staticRouteID)

	log.Printf("Static Route '%s' is being provisioned...", staticRouteID)

	resource, err := apiClient.WaitForDeploy(compute.ResourceTypeStaticRoutes, staticRouteID, resourceCreateTimeoutVLAN)
	if err != nil {
		return err
	}

	data.Partial(true)

	// Capture additional properties that are only available after deployment.
	staticRoute := resource.(*compute.StaticRoute)

	data.Set(resourceKeyStaticRouteState, staticRoute.State)
	data.SetPartial(resourceKeyStaticRouteState)

	data.Set(resourceKeyStaticRouteDataCenter, staticRoute.DataCenter)
	data.SetPartial(resourceKeyStaticRouteDataCenter)

	err = applyNetworkDomainDefaultFirewallRules(data, apiClient)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

func resourceStaticRouteDelete(data *schema.ResourceData, provider interface{}) error {
	log.Println("resourceStaticRouteDelete")
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	id := data.Id()
	name := data.Get(resourceKeyStaticRouteName).(string)

	operationDescription := fmt.Sprintf("Deleting static route '%s'", name)

	var err error
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {

		asyncLock := providerState.AcquireAsyncOperationLock("Delete static route: '%s'", id)
		defer asyncLock.Release()

		deleteErr := apiClient.DeleteStaticRoute(id)

		if compute.IsResourceBusyError(deleteErr) {
			context.Retry()
		} else if err != nil {
			context.Fail(deleteErr)
		}

		asyncLock.Release()
	})

	if err != nil {
		return err
	}

	return nil
}

func resourceStaticRouteImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	log.Println("resourceStaticRouteImport")
	return
}
