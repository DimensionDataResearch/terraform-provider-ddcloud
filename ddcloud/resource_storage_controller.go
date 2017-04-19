package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyStorageControllerServerID    = "server"
	resourceKeyStorageControllerBusNumber   = "scsi_bus_number"
	resourceKeyStorageControllerAdapterType = "adapter_type"
	resourceKeyStorageControllerDisk        = "disk"

	resourceKeyStorageControllerDiskID     = "id"
	resourceKeyStorageControllerDiskUnitID = "scsi_unit_id"
	resourceKeyStorageControllerDiskSizeGB = "size_gb"
	resourceKeyStorageControllerDiskSpeed  = "speed"
)

func resourceStorageController() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		Create:        resourceStorageControllerCreate,
		Read:          resourceStorageControllerRead,
		Update:        resourceStorageControllerUpdate,
		Delete:        resourceStorageControllerDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyStorageControllerServerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the server that the controller is attached to",
			},
			resourceKeyStorageControllerBusNumber: &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The controller's SCSI bus number",
			},
			resourceKeyStorageControllerAdapterType: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of storage adapter used to represent the controller",
			},
			resourceKeyStorageControllerDisk: &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The set of virtual disks attached to the storage controller",
				Elem:        schemaDisk(),
			},
		},
	}
}

// Create a storage controller resource.
func resourceStorageControllerCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	serverID := data.Get(resourceKeyStorageControllerServerID).(string)
	adapterType := data.Get(resourceKeyStorageControllerAdapterType).(string)
	busNumber := data.Get(resourceKeyStorageControllerBusNumber).(int)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	log.Printf("Create storage controller for SCSI bus %d in server '%s'.",
		busNumber,
		server.Name,
	)

	var targetController *compute.VirtualMachineSCSIController
	if busNumber == 0 {
		// Default controller (always present)
		targetController = server.SCSIControllers.GetByBusNumber(busNumber)
		if targetController == nil {
			return fmt.Errorf("cannot find controller for bus %d in server '%s'", busNumber, serverID)
		}

		log.Printf("This controller is the default controller; will treat as already-created.")
	} else {
		operationDescription := fmt.Sprintf("Add SCSI controller for bus %d to server '%s'",
			busNumber,
			serverID,
		)
		err = providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			addControllerError := apiClient.AddSCSIControllerToServer(serverID, adapterType, busNumber)
			if compute.IsResourceBusyError(addControllerError) {
				context.Retry()
			} else if addControllerError != nil {
				context.Fail(addControllerError)
			}
		})

		server, err = apiClient.GetServer(serverID)
		if err != nil {
			return err
		}
		if server == nil {
			return fmt.Errorf("cannot find server '%s'", serverID)
		}

		resource, err := apiClient.WaitForChange(
			compute.ResourceTypeServer,
			serverID,
			"Add SCSI controller",
			resourceUpdateTimeoutServer,
		)
		if err != nil {
			return err
		}

		server = resource.(*compute.Server)
		targetController = server.SCSIControllers.GetByBusNumber(busNumber)
		if targetController == nil {
			return fmt.Errorf("cannot find controller for bus %d in server '%s'", busNumber, serverID)
		}
	}

	data.SetId(targetController.ID)
	propertyHelper.SetDisks(
		models.NewDisksFromVirtualMachineSCSIController(*targetController),
	)

	addDisks := propertyHelper.GetDisks()
	for index := range addDisks {
		addDisk := &addDisks[index]

		operationDescription := fmt.Sprintf("Add disk with SCSI unit ID %d to controller '%s' (for bus %d) in server '%s'",
			addDisk.SCSIUnitID,
			targetController.ID,
			busNumber,
			serverID,
		)
		err := providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			var addDiskError error
			addDisk.ID, addDiskError = apiClient.AddDiskToSCSIController(
				targetController.ID,
				addDisk.SCSIUnitID,
				addDisk.SizeGB,
				addDisk.Speed,
			)
			if compute.IsResourceBusyError(addDiskError) {
				context.Retry()
			} else if addDiskError != nil {
				context.Fail(addDiskError)
			}
		})
		if err != nil {
			return err
		}

		log.Printf("Adding disk '%s' (%dGB, speed = '%s') with SCSI unit ID %d to controller '%s' (for bus %d) in server '%s'...",
			addDisk.ID,
			addDisk.SizeGB,
			addDisk.Speed,
			addDisk.SCSIUnitID,
			targetController.ID,
			busNumber,
			serverID,
		)

		resource, err := apiClient.WaitForChange(
			compute.ResourceTypeServer,
			serverID,
			"Add disk",
			resourceUpdateTimeoutServer,
		)
		if err != nil {
			return err
		}

		server := resource.(*compute.Server)
		targetController := server.SCSIControllers.GetByBusNumber(busNumber)
		if targetController == nil {
			return fmt.Errorf("cannot find controller for bus %d in server '%s'", busNumber, serverID)
		}
		propertyHelper.SetDisks(
			models.NewDisksFromVirtualMachineSCSIController(*targetController),
		)
		propertyHelper.SetPartial(resourceKeyStorageControllerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

		log.Printf("Added disk '%s' with SCSI unit ID %d to to controller '%s' (for bus %d) in server '%s'.",
			addDisk.ID,
			addDisk.SCSIUnitID,
			targetController.ID,
			busNumber,
			serverID,
		)
	}

	return nil
}

// Read a storage controller resource.
func resourceStorageControllerRead(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server != nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		return fmt.Errorf("cannot find controller '%s' in server '%s'", controllerID, serverID)
	}

	log.Printf("Read storage controller '%s' in server '%s'.",
		controllerID,
		server.Name,
	)

	propertyHelper.SetDisks(
		models.NewDisksFromVirtualMachineSCSIController(*targetController),
	)

	return nil
}

// Update a storage controller resource.
func resourceStorageControllerUpdate(data *schema.ResourceData, provider interface{}) error {
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	log.Printf("Update storage controller '%s' in server '%s'.",
		controllerID,
		server.Name,
	)

	return fmt.Errorf("update not implemented yet for ddcloud_storage_controller")
}

// Delete a storage controller resource.
func resourceStorageControllerDelete(data *schema.ResourceData, provider interface{}) error {
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	var (
		server *compute.Server
		err    error
	)
	server, err = apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	log.Printf("Delete storage controller '%s' in server '%s'.",
		controllerID,
		server.Name,
	)

	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		return fmt.Errorf("cannot find controller '%s' in server '%s'", controllerID, serverID)
	}
	if targetController.BusNumber == 0 {
		log.Printf("Controller '%s' is the default adapter and will not be deleted.", controllerID)
		data.SetId("") // Treat as deleted.

		return nil
	}

	return fmt.Errorf("delete not implemented yet for ddcloud_storage_controller")
}
