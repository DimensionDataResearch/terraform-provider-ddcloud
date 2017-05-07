package ddcloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerName                 = "name"
	resourceKeyServerDescription          = "description"
	resourceKeyServerAdminPassword        = "admin_password"
	resourceKeyServerImage                = "image"
	resourceKeyServerImageType            = "image_type"
	resourceKeyServerOSType               = "os_type"
	resourceKeyServerOSFamily             = "os_family"
	resourceKeyServerNetworkDomainID      = "networkdomain"
	resourceKeyServerMemoryGB             = "memory_gb"
	resourceKeyServerCPUCount             = "cpu_count"
	resourceKeyServerCPUCoreCount         = "cores_per_cpu"
	resourceKeyServerCPUSpeed             = "cpu_speed"
	resourceKeyServerPrimaryAdapterVLAN   = "primary_adapter_vlan"
	resourceKeyServerPrimaryAdapterIPv4   = "primary_adapter_ipv4"
	resourceKeyServerPrimaryAdapterIPv6   = "primary_adapter_ipv6"
	resourceKeyServerPublicIPv4           = "public_ipv4"
	resourceKeyServerPrimaryDNS           = "dns_primary"
	resourceKeyServerSecondaryDNS         = "dns_secondary"
	resourceKeyServerAutoStart            = "auto_start"
	resourceKeyServerGuestOSCustomization = "guest_os_customization"

	// Obsolete properties
	resourceKeyServerOSImageID          = "os_image_id"
	resourceKeyServerOSImageName        = "os_image_name"
	resourceKeyServerCustomerImageID    = "customer_image_id"
	resourceKeyServerCustomerImageName  = "customer_image_name"
	resourceKeyServerPrimaryAdapterType = "primary_adapter_type"

	resourceCreateTimeoutServer = 30 * time.Minute
	resourceUpdateTimeoutServer = 10 * time.Minute
	resourceDeleteTimeoutServer = 15 * time.Minute
	serverShutdownTimeout       = 5 * time.Minute
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 4,
		Create:        resourceServerCreate,
		Read:          resourceServerRead,
		Update:        resourceServerUpdate,
		Delete:        resourceServerDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyServerName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the server",
			},
			resourceKeyServerDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the server",
			},
			resourceKeyServerAdminPassword: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Default:     "",
				Description: "The initial administrative password (if applicable) for the deployed server",
			},
			resourceKeyServerMemoryGB: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The amount of memory (in GB) allocated to the server",
			},
			resourceKeyServerCPUCount: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The number of CPUs allocated to the server",
			},
			resourceKeyServerCPUCoreCount: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The number of cores per CPU allocated to the server",
			},
			resourceKeyServerCPUSpeed: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The speed (quality-of-service) for CPUs allocated to the server",
			},
			resourceKeyServerImage: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name or Id of the image from which the server is created",
			},
			resourceKeyServerImageType: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     serverImageTypeAuto,
				Description: "The type of image from which the server is created (default is auto-detect)",
				ValidateFunc: func(value interface{}, key string) (warnings []string, errors []error) {
					imageType := value.(string)

					switch imageType {
					case serverImageTypeOS:
					case serverImageTypeCustomer:
					case serverImageTypeAuto:
						return
					default:
						errors = append(errors,
							fmt.Errorf("invalid image type '%s'", imageType),
						)
					}

					return
				},
			},
			resourceKeyServerOSType: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server operating system type",
			},
			resourceKeyServerOSFamily: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server operating system family",
			},
			resourceKeyServerDisk: schemaDisk(),
			resourceKeyServerNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the server is deployed",
			},
			resourceKeyServerPrimaryNetworkAdapter:    schemaServerPrimaryNetworkAdapter(),
			resourceKeyServerAdditionalNetworkAdapter: schemaServerAdditionalNetworkAdapter(),
			resourceKeyServerPrimaryAdapterVLAN: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Default:     nil,
				Description: "The Id of the VLAN to which the server's primary network adapter will be attached (the first available IPv4 address will be allocated)",
			},
			resourceKeyServerPrimaryAdapterIPv4: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Default:     nil,
				Description: "The IPv4 address for the server's primary network adapter",
			},
			resourceKeyServerPublicIPv4: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Default:     nil,
				Description: "The server's public IPv4 address (if any)",
			},
			resourceKeyServerPrimaryAdapterIPv6: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv6 address of the server's primary network adapter",
			},
			resourceKeyServerPrimaryDNS: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     "",
				Description: "The IP address of the server's primary DNS server",
			},
			resourceKeyServerSecondaryDNS: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     "",
				Description: "The IP address of the server's secondary DNS server",
			},
			resourceKeyServerAutoStart: &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should the server be started automatically once it has been deployed",
			},
			resourceKeyServerTag: schemaServerTag(),

			// Obsolete properties
			resourceKeyServerPrimaryAdapterType: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Default:     nil,
				Description: "The type of the server's primary network adapter (E1000 or VMXNET3)",
				Removed:     "This property has been removed because it is not exposed via the CloudControl API and will not be available until the provider uses the new (v2.4) API",
			},
			resourceKeyServerOSImageID: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The Id of the OS (built-in) image from which the server is created",
				Removed:     fmt.Sprintf("This property has been removed; use %s instead.", resourceKeyServerImage),
			},
			resourceKeyServerOSImageName: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The name of the OS (built-in) image from which the server is created",
				Removed:     fmt.Sprintf("This property has been removed; use %s instead.", resourceKeyServerImage),
			},
			resourceKeyServerCustomerImageID: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The Id of the customer (custom) image from which the server is created",
				Removed:     fmt.Sprintf("This property has been removed; use %s instead.", resourceKeyServerImage),
			},
			resourceKeyServerCustomerImageName: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The name of the customer (custom) image from which the server is created",
				Removed:     fmt.Sprintf("This property has been removed; use %s instead.", resourceKeyServerImage),
			},
		},
		MigrateState: resourceServerMigrateState,
	}
}

// Create a server resource.
func resourceServerCreate(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Create server '%s' in network domain '%s' (description = '%s').", name, networkDomainID, description)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	networkDomain, err := apiClient.GetNetworkDomain(networkDomainID)
	if err != nil {
		return err
	}

	if networkDomain == nil {
		return fmt.Errorf("no network domain was found with Id '%s'", networkDomainID)
	}

	dataCenterID := networkDomain.DatacenterID
	log.Printf("Server will be deployed in data centre '%s'.", dataCenterID)

	configuredImage := data.Get(resourceKeyServerImage).(string)
	configuredImageType := data.Get(resourceKeyServerImageType).(string)
	image, err := resolveServerImage(configuredImage, configuredImageType, dataCenterID, apiClient)
	if err != nil {
		return err
	}
	if image == nil {
		return fmt.Errorf("an unexpected error occurred while resolving the configured server image")
	}

	if image.RequiresCustomization() {
		return deployCustomizedServer(data, provider)
	}

	return deployUncustomizedServer(data, provider)
}

// Read a server resource.
func resourceServerRead(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	id := data.Id()
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Read server '%s' (Id = '%s') in network domain '%s' (description = '%s').", name, id, networkDomainID, description)

	apiClient := provider.(*providerState).Client()
	server, err := apiClient.GetServer(id)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("Server '%s' has been deleted.", id)

		// Mark as deleted.
		data.SetId("")

		return nil
	}
	data.Set(resourceKeyServerName, server.Name)
	data.Set(resourceKeyServerDescription, server.Description)
	data.Set(resourceKeyServerMemoryGB, server.MemoryGB)
	data.Set(resourceKeyServerCPUCount, server.CPU.Count)
	data.Set(resourceKeyServerCPUCoreCount, server.CPU.CoresPerSocket)
	data.Set(resourceKeyServerCPUSpeed, server.CPU.Speed)

	captureServerNetworkConfiguration(server, data, false)

	var publicIPv4Address string
	publicIPv4Address, err = findPublicIPv4Address(apiClient,
		networkDomainID,
		*server.Network.PrimaryAdapter.PrivateIPv4Address,
	)
	if err != nil {
		return err
	}
	if !isEmpty(publicIPv4Address) {
		data.Set(resourceKeyServerPublicIPv4, publicIPv4Address)
	} else {
		data.Set(resourceKeyServerPublicIPv4, nil)
	}

	err = readServerTags(data, apiClient)
	if err != nil {
		return err
	}

	propertyHelper.SetDisks(
		models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers),
	)

	return nil
}

// Update a server resource.
func resourceServerUpdate(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Id()

	log.Printf("Update server '%s'.", serverID)

	providerState := provider.(*providerState)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("Server '%s' has been deleted.", serverID)
		data.SetId("")

		return nil
	}

	data.Partial(true)

	propertyHelper := propertyHelper(data)

	var name, description *string
	if data.HasChange(resourceKeyServerName) {
		name = propertyHelper.GetOptionalString(resourceKeyServerName, true)
	}

	if data.HasChange(resourceKeyServerDescription) {
		description = propertyHelper.GetOptionalString(resourceKeyServerDescription, true)
	}

	if name != nil || description != nil {
		log.Printf("Server name / description change detected.")

		err = apiClient.EditServerMetadata(serverID, name, description)
		if err != nil {
			return err
		}

		if name != nil {
			data.SetPartial(resourceKeyServerName)
		}
		if description != nil {
			data.SetPartial(resourceKeyServerDescription)
		}
	}

	var memoryGB, cpuCount, cpuCoreCount *int
	var cpuSpeed *string
	if data.HasChange(resourceKeyServerMemoryGB) {
		memoryGB = propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	}
	if data.HasChange(resourceKeyServerCPUCount) {
		cpuCount = propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	}
	if data.HasChange(resourceKeyServerCPUCoreCount) {
		cpuCoreCount = propertyHelper.GetOptionalInt(resourceKeyServerCPUCoreCount, false)
	}
	if data.HasChange(resourceKeyServerCPUSpeed) {
		cpuSpeed = propertyHelper.GetOptionalString(resourceKeyServerCPUSpeed, false)
	}

	if memoryGB != nil || cpuCount != nil || cpuCoreCount != nil || cpuSpeed != nil {
		log.Printf("Server CPU / memory configuration change detected.")

		err = updateServerConfiguration(apiClient, server, memoryGB, cpuCount, cpuCoreCount, cpuSpeed)
		if err != nil {
			return err
		}

		if data.HasChange(resourceKeyServerMemoryGB) {
			data.SetPartial(resourceKeyServerMemoryGB)
		}

		if data.HasChange(resourceKeyServerCPUCount) {
			data.SetPartial(resourceKeyServerCPUCount)
		}
	}

	if data.HasChange(resourceKeyServerPrimaryNetworkAdapter) {
		actualNetworkAdapters := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network)
		actualPrimaryNetworkAdapter := models.NewNetworkAdapterFromVirtualMachineNetworkAdapter(server.Network.PrimaryAdapter)

		configuredPrimaryNetworkAdapter := propertyHelper.GetServerNetworkAdapters().GetPrimary()

		log.Printf("Configured primary network adapter = %#v", configuredPrimaryNetworkAdapter)
		log.Printf("Actual primary network adapter     = %#v", actualPrimaryNetworkAdapter)

		if configuredPrimaryNetworkAdapter.PrivateIPv4Address != actualPrimaryNetworkAdapter.PrivateIPv4Address {
			err = modifyServerNetworkAdapterIP(providerState, serverID, *configuredPrimaryNetworkAdapter)
			if err != nil {
				return err
			}
		}

		if configuredPrimaryNetworkAdapter.AdapterType != actualPrimaryNetworkAdapter.AdapterType {
			err = modifyServerNetworkAdapterType(providerState, serverID, *configuredPrimaryNetworkAdapter)
			if err != nil {
				return err
			}
		}

		// Capture updated public IPv4 address (if any).
		var publicIPv4Address string
		publicIPv4Address, err = findPublicIPv4Address(apiClient,
			server.Network.NetworkDomainID,
			*server.Network.PrimaryAdapter.PrivateIPv4Address,
		)
		if err != nil {
			return err
		}
		if !isEmpty(publicIPv4Address) {
			data.Set(resourceKeyServerPublicIPv4, publicIPv4Address)
		} else {
			data.Set(resourceKeyServerPublicIPv4, nil)
		}

		// Persist final state.
		server, err = apiClient.GetServer(serverID)
		if err != nil {
			return err
		}
		if server == nil {
			return fmt.Errorf("cannot find server with Id '%s'", serverID)
		}

		actualNetworkAdapters = models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network)
		propertyHelper.SetServerNetworkAdapters(actualNetworkAdapters, true)
	}

	if data.HasChange(resourceKeyServerTag) {
		err = applyServerTags(data, apiClient, providerState.Settings())
		if err != nil {
			return err
		}

		data.SetPartial(resourceKeyServerTag)
	}

	if data.HasChange(resourceKeyServerDisk) {
		err = updateDisks(data, providerState)
		if err != nil {
			return err
		}
	}

	data.Partial(false)

	return nil
}

// Delete a server resource.
func resourceServerDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyServerName).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Delete server '%s' ('%s') in network domain '%s'.", id, name, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(id)
	if err != nil {
		return err
	}
	if server == nil {
		log.Printf("Server '%s' not found; will treat the server as having already been deleted.", id)

		return nil
	}

	if server.Started {
		log.Printf("Server '%s' is currently running. The server will be powered off.", id)
		err = serverPowerOff(providerState, id)
		if err != nil {
			return err
		}
	}

	operationDescription := fmt.Sprintf("Delete server '%s'", id)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		deleteError := apiClient.DeleteServer(id)
		if compute.IsResourceBusyError(deleteError) {
			context.Retry()
		} else if deleteError != nil {
			context.Fail(deleteError)
		}
	})
	if err != nil {
		return err
	}

	log.Printf("Server '%s' is being deleted...", id)

	return apiClient.WaitForDelete(compute.ResourceTypeServer, id, resourceDeleteTimeoutServer)
}

// TODO: Refactor deployCustomizedServer / deployUncustomizedServer and move common logic to shared functions.

// Deploy a server with guest OS customisation.
func deployCustomizedServer(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	adminPassword := data.Get(resourceKeyServerAdminPassword).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)
	primaryDNS := data.Get(resourceKeyServerPrimaryDNS).(string)
	secondaryDNS := data.Get(resourceKeyServerSecondaryDNS).(string)
	autoStart := data.Get(resourceKeyServerAutoStart).(bool)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	networkDomain, err := apiClient.GetNetworkDomain(networkDomainID)
	if err != nil {
		return err
	}

	if networkDomain == nil {
		return fmt.Errorf("no network domain was found with Id '%s'", networkDomainID)
	}

	dataCenterID := networkDomain.DatacenterID
	log.Printf("Server will be deployed in data centre '%s' with guest OS customisation.", dataCenterID)

	propertyHelper := propertyHelper(data)
	configuredImage := data.Get(resourceKeyServerImage).(string)
	configuredImageType := data.Get(resourceKeyServerImageType).(string)
	image, err := resolveServerImage(configuredImage, configuredImageType, dataCenterID, apiClient)
	if err != nil {
		return err
	}
	if image == nil {
		return fmt.Errorf("an unexpected error occurred while resolving the configured server image")
	}

	log.Printf("Server will be deployed from %s image named '%s' (Id = '%s').",
		compute.ImageTypeName(image.GetType()),
		image.GetName(),
		image.GetID(),
	)
	deploymentConfiguration := compute.ServerDeploymentConfiguration{
		Name:                  name,
		Description:           description,
		AdministratorPassword: adminPassword,
		Start: autoStart,
	}
	err = validateAdminPassword(deploymentConfiguration.AdministratorPassword, image)
	if err != nil {
		return err
	}
	image.ApplyTo(&deploymentConfiguration)

	operatingSystem := image.GetOS()
	data.Set(resourceKeyServerOSType, operatingSystem.DisplayName)
	data.SetPartial(resourceKeyServerOSType)
	data.Set(resourceKeyServerOSFamily, operatingSystem.Family)
	data.SetPartial(resourceKeyServerOSFamily)

	// Validate disk configuration.
	configuredDisks := propertyHelper.GetDisks()
	err = validateServerDisks(configuredDisks)
	if err != nil {
		return err
	}

	// Image disk speeds
	configuredDisksBySCSIPath := configuredDisks.BySCSIPath()
	for controllerIndex := range deploymentConfiguration.SCSIControllers {
		deploymentSCSIController := &deploymentConfiguration.SCSIControllers[controllerIndex]
		for diskIndex := range deploymentSCSIController.Disks {
			deploymentDisk := &deploymentSCSIController.Disks[diskIndex]

			configuredDisk, ok := configuredDisksBySCSIPath[models.SCSIPath(deploymentSCSIController.BusNumber, deploymentDisk.SCSIUnitID)]
			if ok {
				deploymentDisk.Speed = configuredDisk.Speed
			}
		}
	}

	// Memory and CPU
	memoryGB := propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	if memoryGB != nil {
		deploymentConfiguration.MemoryGB = *memoryGB
	} else {
		data.Set(resourceKeyServerMemoryGB, deploymentConfiguration.MemoryGB)
	}

	cpuCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	if cpuCount != nil {
		deploymentConfiguration.CPU.Count = *cpuCount
	} else {
		data.Set(resourceKeyServerCPUCount, deploymentConfiguration.CPU.Count)
	}

	cpuCoreCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCoreCount, false)
	if cpuCoreCount != nil {
		deploymentConfiguration.CPU.CoresPerSocket = *cpuCoreCount
	} else {
		data.Set(resourceKeyServerCPUCoreCount, deploymentConfiguration.CPU.CoresPerSocket)
	}

	cpuSpeed := propertyHelper.GetOptionalString(resourceKeyServerCPUSpeed, false)
	if cpuSpeed != nil {
		deploymentConfiguration.CPU.Speed = *cpuSpeed
	} else {
		data.Set(resourceKeyServerCPUSpeed, deploymentConfiguration.CPU.Speed)
	}

	// Network
	deploymentConfiguration.Network = compute.VirtualMachineNetwork{
		NetworkDomainID: networkDomainID,
	}

	// Initial configuration for network adapters.
	networkAdapters := propertyHelper.GetServerNetworkAdapters()
	networkAdapters.UpdateVirtualMachineNetwork(&deploymentConfiguration.Network)

	deploymentConfiguration.PrimaryDNS = primaryDNS
	deploymentConfiguration.SecondaryDNS = secondaryDNS

	log.Printf("Server deployment configuration: %+v", deploymentConfiguration)
	log.Printf("Server CPU deployment configuration: %+v", deploymentConfiguration.CPU)

	var serverID string
	operationDescription := fmt.Sprintf("Deploy server '%s'", name)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		var deployError error
		serverID, deployError = apiClient.DeployServer(deploymentConfiguration)
		if compute.IsResourceBusyError(deployError) {
			context.Retry()
		} else if deployError != nil {
			context.Fail(deployError)
		}
	})
	if err != nil {
		return err
	}
	data.SetId(serverID)

	log.Printf("Server '%s' is being provisioned...", name)
	resource, err := apiClient.WaitForDeploy(compute.ResourceTypeServer, serverID, resourceCreateTimeoutServer)
	if err != nil {
		return err
	}

	// Capture additional properties that may only be available after deployment.
	server := resource.(*compute.Server)
	err = captureCreatedServerProperties(data, provider, server, networkAdapters)
	if err != nil {
		return err
	}

	data.Partial(true)

	err = applyServerTags(data, apiClient, providerState.Settings())
	if err != nil {
		return err
	}
	data.SetPartial(resourceKeyServerTag)

	err = createDisks(server.SCSIControllers, data, providerState)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

// Deploy a server without guest OS customisation.
func deployUncustomizedServer(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)
	autoStart := data.Get(resourceKeyServerAutoStart).(bool)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	networkDomain, err := apiClient.GetNetworkDomain(networkDomainID)
	if err != nil {
		return err
	}

	if networkDomain == nil {
		return fmt.Errorf("no network domain was found with Id '%s'", networkDomainID)
	}

	dataCenterID := networkDomain.DatacenterID
	log.Printf("Server will be deployed in data centre '%s' without guest OS customisation.", dataCenterID)

	propertyHelper := propertyHelper(data)
	configuredImage := data.Get(resourceKeyServerImage).(string)
	configuredImageType := data.Get(resourceKeyServerImageType).(string)
	image, err := resolveServerImage(configuredImage, configuredImageType, dataCenterID, apiClient)
	if err != nil {
		return err
	}
	if image == nil {
		return fmt.Errorf("an unexpected error occurred while resolving the configured server image")
	}

	log.Printf("Server will be deployed from %s image named '%s' (Id = '%s').",
		compute.ImageTypeName(image.GetType()),
		image.GetName(),
		image.GetID(),
	)
	deploymentConfiguration := compute.UncustomizedServerDeploymentConfiguration{
		Name:        name,
		Description: description,
		Start:       autoStart,
	}
	image.ApplyToUncustomized(&deploymentConfiguration)

	operatingSystem := image.GetOS()
	data.Set(resourceKeyServerOSType, operatingSystem.DisplayName)
	data.SetPartial(resourceKeyServerOSType)
	data.Set(resourceKeyServerOSFamily, operatingSystem.Family)
	data.SetPartial(resourceKeyServerOSFamily)

	// Validate disk configuration.
	configuredDisks := propertyHelper.GetDisks()
	err = validateServerDisks(configuredDisks)
	if err != nil {
		return err
	}

	// Image disk speeds
	configuredDisksBySCSIPath := configuredDisks.BySCSIPath()
	for diskIndex := range deploymentConfiguration.Disks {
		deploymentDisk := &deploymentConfiguration.Disks[diskIndex]

		configuredDisk, ok := configuredDisksBySCSIPath[models.SCSIPath(0, deploymentDisk.SCSIUnitID)]
		if ok {
			deploymentDisk.Speed = configuredDisk.Speed
		}
	}

	// Memory and CPU
	memoryGB := propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	if memoryGB != nil {
		deploymentConfiguration.MemoryGB = *memoryGB
	} else {
		data.Set(resourceKeyServerMemoryGB, deploymentConfiguration.MemoryGB)
	}

	cpuCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	if cpuCount != nil {
		deploymentConfiguration.CPU.Count = *cpuCount
	} else {
		data.Set(resourceKeyServerCPUCount, deploymentConfiguration.CPU.Count)
	}

	cpuCoreCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCoreCount, false)
	if cpuCoreCount != nil {
		deploymentConfiguration.CPU.CoresPerSocket = *cpuCoreCount
	} else {
		data.Set(resourceKeyServerCPUCoreCount, deploymentConfiguration.CPU.CoresPerSocket)
	}

	cpuSpeed := propertyHelper.GetOptionalString(resourceKeyServerCPUSpeed, false)
	if cpuSpeed != nil {
		deploymentConfiguration.CPU.Speed = *cpuSpeed
	} else {
		data.Set(resourceKeyServerCPUSpeed, deploymentConfiguration.CPU.Speed)
	}

	// Network
	deploymentConfiguration.Network = compute.VirtualMachineNetwork{
		NetworkDomainID: networkDomainID,
	}

	// Initial configuration for network adapters.
	networkAdapters := propertyHelper.GetServerNetworkAdapters()
	networkAdapters.UpdateVirtualMachineNetwork(&deploymentConfiguration.Network)

	log.Printf("Server deployment configuration: %+v", deploymentConfiguration)
	log.Printf("Server CPU deployment configuration: %+v", deploymentConfiguration.CPU)

	var serverID string
	operationDescription := fmt.Sprintf("Deploy uncustomised server '%s'", name)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		var deployError error
		serverID, deployError = apiClient.DeployUncustomizedServer(deploymentConfiguration)
		if compute.IsResourceBusyError(deployError) {
			context.Retry()
		} else if deployError != nil {
			context.Fail(deployError)
		}
	})
	if err != nil {
		return err
	}
	data.SetId(serverID)

	log.Printf("Server '%s' is being provisioned...", name)
	resource, err := apiClient.WaitForDeploy(compute.ResourceTypeServer, serverID, resourceCreateTimeoutServer)
	if err != nil {
		return err
	}

	// Capture additional properties that may only be available after deployment.
	server := resource.(*compute.Server)
	err = captureCreatedServerProperties(data, provider, server, networkAdapters)
	if err != nil {
		return err
	}

	data.Partial(true)

	err = applyServerTags(data, apiClient, providerState.Settings())
	if err != nil {
		return err
	}
	data.SetPartial(resourceKeyServerTag)

	err = createDisks(server.SCSIControllers, data, providerState)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

// Capture additional properties that may only be available after deployment.
func captureCreatedServerProperties(data *schema.ResourceData, provider interface{}, server *compute.Server, networkAdapters models.NetworkAdapters) error {
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	propertyHelper := propertyHelper(data)

	networkAdapters.CaptureIDs(server.Network)
	propertyHelper.SetServerNetworkAdapters(networkAdapters, true)
	captureServerNetworkConfiguration(server, data, true)

	publicIPv4Address, err := findPublicIPv4Address(apiClient,
		networkDomainID,
		*server.Network.PrimaryAdapter.PrivateIPv4Address,
	)
	if err != nil {
		return err
	}
	if !isEmpty(publicIPv4Address) {
		data.Set(resourceKeyServerPublicIPv4, publicIPv4Address)
	} else {
		data.Set(resourceKeyServerPublicIPv4, nil)
	}
	data.SetPartial(resourceKeyServerPublicIPv4)

	return nil
}

func findPublicIPv4Address(apiClient *compute.Client, networkDomainID string, privateIPv4Address string) (publicIPv4Address string, err error) {
	page := compute.DefaultPaging()
	for {
		var natRules *compute.NATRules
		natRules, err = apiClient.ListNATRules(networkDomainID, page)
		if err != nil {
			return
		}
		if natRules.IsEmpty() {
			break // We're done
		}

		for _, natRule := range natRules.Rules {
			if natRule.InternalIPAddress == privateIPv4Address {
				return natRule.ExternalIPAddress, nil
			}
		}

		page.Next()
	}

	return
}

func validateAdminPassword(adminPassword string, image compute.Image) error {
	switch image.GetType() {
	case compute.ImageTypeOS:
		// Admin password is always mandatory for OS images.
		if adminPassword == "" {
			return fmt.Errorf("Must specify an initial admin password when deploying an OS image")
		}
	case compute.ImageTypeCustomer:
		imageOS := image.GetOS()

		// Admin password cannot be supplied for Linux customer images.
		if imageOS.Family == "UNIX" && adminPassword != "" {
			return fmt.Errorf("cannot specify an initial admin password when deploying a Linux OS image")
		}

		// Admin password is only mandatory for some types of Windows images
		if imageOS.Family != "WINDOWS" {
			return nil
		}

		if adminPassword == "" {
			// Mandatory for Windows Server 2008.
			if strings.HasPrefix(imageOS.ID, "WIN2008") {

				return fmt.Errorf("Must specify an initial admin password when deploying a customer image for Windows Server 2008")
			}

			// Mandatory for Windows Server 2012 R2.
			if strings.HasPrefix(imageOS.ID, "WIN2012R2") {
				return fmt.Errorf("Must specify an initial admin password when deploying a customer image for Windows Server 2012 R2")
			}

			// Mandatory for Windows Server 2012.
			if strings.HasPrefix(imageOS.ID, "WIN2012") {
				return fmt.Errorf("Must specify an initial admin password when deploying a customer image for Windows Server 2012")
			}
		}

	default:
		return fmt.Errorf("Unknown image type (%d)", image.GetType())
	}

	return nil
}

// Start a server.
//
// Respects providerSettings.AllowServerReboots.
func serverStart(providerState *providerState, serverID string) error {
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	if !providerSettings.AllowServerReboots {
		return fmt.Errorf("cannot start server '%s' because server reboots have not been enabled via the 'allow_server_reboot' provider setting or 'DDCLOUD_ALLOW_SERVER_REBOOT' environment variable", serverID)
	}

	operationDescription := fmt.Sprintf("Start server '%s'", serverID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		startError := apiClient.StartServer(serverID)
		if compute.IsResourceBusyError(startError) {
			context.Retry()
		} else if startError != nil {
			context.Fail(startError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Start server", serverShutdownTimeout)
	if err != nil {
		return err
	}

	return nil
}

// Gracefully stop a server.
//
// Respects providerSettings.AllowServerReboots.
func serverShutdown(providerState *providerState, serverID string) error {
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	if !providerSettings.AllowServerReboots {
		return fmt.Errorf("cannot shut down server '%s' because server reboots have not been enabled via the 'allow_server_reboot' provider setting or 'DDCLOUD_ALLOW_SERVER_REBOOT' environment variable", serverID)
	}

	operationDescription := fmt.Sprintf("Shut down server '%s'", serverID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		shutdownError := apiClient.ShutdownServer(serverID)
		if compute.IsResourceBusyError(shutdownError) {
			context.Retry()
		} else if shutdownError != nil {
			context.Fail(shutdownError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Shut down server", serverShutdownTimeout)
	if err != nil {
		return err
	}

	return nil
}

// Forcefully stop a server.
//
// Does not respect providerSettings.AllowServerReboots.
func serverPowerOff(providerState *providerState, serverID string) error {
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Power off server '%s'", serverID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		powerOffError := apiClient.PowerOffServer(serverID)
		if compute.IsResourceBusyError(powerOffError) {
			context.Retry()
		} else if powerOffError != nil {
			context.Fail(powerOffError)
		}
	})
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Power off server", serverShutdownTimeout)
	if err != nil {
		return err
	}

	return nil
}
