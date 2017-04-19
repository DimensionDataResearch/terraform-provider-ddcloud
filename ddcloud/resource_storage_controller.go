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
				Optional:    true,
				Default:     compute.StorageControllerAdapterTypeLSILogicParallel,
				ForceNew:    true,
				Description: "The type of storage adapter used to represent the controller",
			},
			resourceKeyStorageControllerDisk: schemaDisk(),
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
		operationDescription := fmt.Sprintf("Add storage controller for bus %d to server '%s'",
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
			"Add storage controller",
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

	log.Printf("Target storage controller '%s' has configuration: %#v",
		targetController.ID,
		targetController,
	)

	data.SetId(targetController.ID)

	configuredDisks := propertyHelper.GetDisks()
	actualDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)

	if configuredDisks.IsEmpty() {
		// No explicitly-configured disks.
		propertyHelper.SetDisks(actualDisks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

		return nil
	}

	return updateStorageControllerDisks(data, providerState)
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
	if server == nil {
		log.Printf("Cannot find server '%s' for controller '%s'.", serverID, controllerID)

		// If the server is deleted, then so is the controller.
		data.SetId("")

		return nil
	}

	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		log.Printf("Cannot find controller '%s' in server '%s'.", controllerID, serverID)

		// Mark as deleted.
		data.SetId("")

		return nil
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
	if server == nil {
		log.Printf("Cannot find server '%s' for controller '%s'; will treat the controller as deleted.", serverID, controllerID)

		// If the server is deleted, then so is the controller.
		data.SetId("")

		return nil
	}

	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		log.Printf("Cannot find controller '%s' in server '%s'; will treat the controller as deleted.", controllerID, serverID)

		// Mark as deleted.
		data.SetId("")

		return nil
	}

	log.Printf("Update storage controller '%s' in server '%s'.",
		controllerID,
		server.Name,
	)

	return updateStorageControllerDisks(data, providerState)
}

// Delete a storage controller resource.
func resourceStorageControllerDelete(data *schema.ResourceData, provider interface{}) error {
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		log.Printf("Cannot find server '%s' for controller '%s'; will treat the controller as deleted.", serverID, controllerID)

		// If the server is deleted, then so is the controller.
		data.SetId("")

		return nil
	}

	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		log.Printf("Cannot find controller '%s' in server '%s'; will treat the controller as deleted.", controllerID, serverID)

		// Mark as deleted.
		data.SetId("")

		return nil
	}

	log.Printf("Delete storage controller '%s' in server '%s'.",
		controllerID,
		server.Name,
	)

	removeDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)
	err = processRemoveStorageControllerDisks(removeDisks, data, providerState)
	if err != nil {
		return err
	}

	if targetController.BusNumber == 0 {
		log.Printf("Controller '%s' is the default adapter; will treat this ddcloud_storage_controller as deleted (but can't actually remove the default adapter).", controllerID)
		data.SetId("") // Treat as deleted.

		return nil
	}

	return nil
}

// When updating a storage controller resource, synchronise the controller's disk attributes with its resource data
func updateStorageControllerDisks(data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	log.Printf("Configure image disks for storage controller '%s' in server '%s'...", controllerID, serverID)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}
	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		return fmt.Errorf("cannot find storage controller '%s' in server '%s'", controllerID, serverID)
	}

	log.Printf("Current state for storage controller '%s': %#v",
		controllerID,
		targetController,
	)

	// Filter disks so we're only looking at ones from this controller.
	actualDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)

	log.Printf("Current disks for storage controller '%s': %#v",
		controllerID,
		actualDisks,
	)

	configuredDisks := propertyHelper.GetDisks()
	log.Printf("Configuration for storage controller '%s' specifies %d disks: %#v.", controllerID, len(configuredDisks), configuredDisks)

	err = validateStorageControllerDisks(configuredDisks)
	if err != nil {
		return err
	}

	if configuredDisks.IsEmpty() {
		// No explicitly-configured disks.
		propertyHelper.SetDisks(actualDisks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

		return nil
	}

	addDisks, modifyDisks, removeDisks := configuredDisks.SplitByAction(actualDisks)
	if addDisks.IsEmpty() && modifyDisks.IsEmpty() && removeDisks.IsEmpty() {
		log.Printf("No post-deploy changes required for disks of server '%s'.", serverID)

		return nil
	}

	// First remove any disks that are no longer required.
	err = processRemoveStorageControllerDisks(removeDisks, data, providerState)
	if err != nil {
		return err
	}

	// Then modify existing disks
	err = processModifyStorageControllerDisks(modifyDisks, data, providerState)
	if err != nil {
		return err
	}

	// Finally, add new disks
	err = processAddStorageControllerDisks(addDisks, data, providerState)
	if err != nil {
		return err
	}

	return nil
}

// Process the collection of disks that need to be added to the server.
func processAddStorageControllerDisks(addDisks models.Disks, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	controllerID := data.Id()
	busNumber := data.Get(resourceKeyStorageControllerBusNumber).(int)
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}
	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		return fmt.Errorf("cannot find storage controller '%s' in server '%s'", controllerID, serverID)
	}

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

// Process the collection of disks whose configuration needs to be modified.
//
// Disk Ids must already be populated.
func processModifyStorageControllerDisks(modifyDisks models.Disks, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}
	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		return fmt.Errorf("cannot find storage controller '%s' in server '%s'", controllerID, serverID)
	}

	actualDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)
	actualDisksBySCSIPath := actualDisks.BySCSIPath()

	for index := range modifyDisks {
		modifyDisk := &modifyDisks[index]
		actualImageDisk := actualDisksBySCSIPath[modifyDisk.SCSIPath()]

		// Can't shrink disk, only grow it.
		if modifyDisk.SizeGB < actualImageDisk.SizeGB {
			return fmt.Errorf(
				"cannot resize disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' from %d to GB to %d (for now, disks can only be expanded)",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}

		// Do we need to expand the disk?
		if modifyDisk.SizeGB > actualImageDisk.SizeGB {
			log.Printf(
				"Expanding disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from %d GB to %d GB)...",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)

			operationDescription := fmt.Sprintf("Expand disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s'",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
			)
			err = providerState.RetryAction(operationDescription, func(context retry.Context) {
				asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
				defer asyncLock.Release()

				response, resizeError := apiClient.ExpandDisk(modifyDisk.ID, modifyDisk.SizeGB)
				if compute.IsResourceBusyError(resizeError) {
					context.Retry()
				} else if resizeError != nil {
					context.Fail(resizeError)
				}
				if response.ResponseCode != compute.ResponseCodeInProgress {
					context.Fail(response.ToError("unexpected response code '%s' when expanding server disk '%s' for server '%s'",
						response.ResponseCode,
						modifyDisk.ID,
						serverID,
					))
				}
			})
			if err != nil {
				return err
			}

			log.Printf(
				"Resizing disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from %d to GB to %d)...",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)

			resource, err := apiClient.WaitForChange(
				compute.ResourceTypeServer,
				serverID,
				"Resize disk",
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			modifyDisk.SizeGB = actualImageDisk.SizeGB

			server := resource.(*compute.Server)
			propertyHelper.SetDisks(
				models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers),
			)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf("storage controller '%s' now has %d disks: %#v.", targetController.ID, len(targetController.Disks), targetController)

			log.Printf(
				"Resized disk '%s' on storage controller '%s' (SCSI bus %d) server '%s' (from %d to GB to %d).",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}

		// Do we need to change the disk speed?
		if modifyDisk.Speed != actualImageDisk.Speed {
			log.Printf(
				"Changing speed of disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from '%s' to '%s')...",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualImageDisk.Speed,
				modifyDisk.Speed,
			)

			operationDescription := fmt.Sprintf("Change speed of disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s'",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
			)
			err = providerState.RetryAction(operationDescription, func(context retry.Context) {
				asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
				defer asyncLock.Release()

				response, resizeError := apiClient.ChangeServerDiskSpeed(serverID, modifyDisk.ID, modifyDisk.Speed)
				if compute.IsResourceBusyError(resizeError) {
					context.Retry()
				} else if resizeError != nil {
					context.Fail(resizeError)
				}
				if response.Result != compute.ResultSuccess {
					context.Fail(response.ToError(
						"Unexpected result '%s' when resizing disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s'.",
						response.Result,
						modifyDisk.ID,
						targetController.ID,
						targetController.BusNumber,
						serverID,
					))
				}
			})
			if err != nil {
				return err
			}

			resource, err := apiClient.WaitForChange(
				compute.ResourceTypeServer,
				serverID,
				"Resize disk",
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			modifyDisk.Speed = actualImageDisk.Speed

			server = resource.(*compute.Server)
			propertyHelper.SetDisks(
				models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers),
			)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf("Resized disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from %d to GB to %d).",
				modifyDisk.ID,
				targetController.ID,
				targetController.Disks,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}
	}

	return nil
}

// Process the collection of disks that need to be removed.
//
// Disk Ids must already be populated.
func processRemoveStorageControllerDisks(removeDisks models.Disks, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}
	targetController := server.SCSIControllers.GetByID(controllerID)
	if targetController == nil {
		return fmt.Errorf("cannot find storage controller '%s' in server '%s'", controllerID, serverID)
	}

	for _, removeDisk := range removeDisks {
		log.Printf("Remove disk '%s' (SCSI unit Id %d) from storage controller '%s' (SCSI bus %d) in server '%s'...",
			removeDisk.ID,
			removeDisk.SCSIUnitID,
			targetController.ID,
			targetController.BusNumber,
			serverID,
		)

		// Can't remove the last disk in a server.
		if server.SCSIControllers.GetDiskCount() == 1 {
			log.Printf("Not removing disk '%s' (SCSI unit Id %d) from storage controller '%s' (SCSI bus %d) in server '%s' because this is the server's last disk.",
				removeDisk.ID,
				removeDisk.SCSIUnitID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
			)

			continue
		}

		operationDescription := fmt.Sprintf("Remove disk '%s' (SCSI unit Id %d) from storage controller '%s' (SCSI bus %d) in server '%s'",
			removeDisk.ID,
			removeDisk.SCSIUnitID,
			targetController.ID,
			targetController.BusNumber,
			serverID,
		)
		err = providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			removeError := apiClient.RemoveDisk(removeDisk.ID)
			if compute.IsResourceBusyError(removeError) {
				context.Retry()
			} else if removeError != nil {
				context.Fail(removeError)
			}
		})
		if err != nil {
			return err
		}

		resource, err := apiClient.WaitForChange(
			compute.ResourceTypeServer,
			serverID,
			"Remove disk",
			resourceUpdateTimeoutServer,
		)
		if err != nil {
			return err
		}

		server = resource.(*compute.Server)
		propertyHelper.SetDisks(
			models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers),
		)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf(
			"Removed disk '%s' (SCSI unit Id %d) from storage controller '%s' (SCSI bus %d) in server '%s'.",
			removeDisk.ID,
			removeDisk.SCSIUnitID,
			targetController.ID,
			targetController.BusNumber,
			serverID,
		)
	}

	return nil
}

func validateStorageControllerDisks(disks models.Disks) error {
	return validateDisks(disks)
}
