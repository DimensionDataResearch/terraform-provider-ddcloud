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

/*
 * AF: Note that "terraform plan" produces an incorrect diff when a disk is removed from the ddcloud_storage_controller
* (no idea why, but I'd say it's probably a bug, and should be reported as such).
 *
 * This implementation will still do the correct thing (i.e. remove the disk), but the diff output from "terraform plan" is confusing for the user.
*/

func resourceStorageController() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		Create:        resourceStorageControllerCreate,
		Read:          resourceStorageControllerRead,
		Update:        resourceStorageControllerUpdate,
		Delete:        resourceStorageControllerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStorageControllerImport,
		},

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
		// No explicitly-configured disks so just populate from current controller state.
		propertyHelper.SetDisks(actualDisks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("No disks configured; storage controller '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

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

	configuredDisks := propertyHelper.GetDisks()
	log.Printf("Configuration for storage controller '%s' (bus %d) specifies %d disks: %#v.", controllerID, targetController.BusNumber, len(configuredDisks), configuredDisks)

	oldConfDisks, newConfDisks := data.GetChange(resourceKeyServerDisk)
	log.Printf("Configuration for storage controller '%s' (bus %d) previously specified: %#v.", controllerID, targetController.BusNumber, oldConfDisks)
	log.Printf("Configuration for storage controller '%s' (bus %d) now specifies: %#v.", controllerID, targetController.BusNumber, newConfDisks)

	actualDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)
	log.Printf("Storage controller '%s' currently has %d disks: %#v.", controllerID, len(actualDisks), actualDisks)

	propertyHelper.SetDisks(actualDisks)

	return nil
}

// Update a storage controller resource.
func resourceStorageControllerUpdate(data *schema.ResourceData, provider interface{}) error {
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

	configuredDisks := propertyHelper.GetDisks()
	actualDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)

	if configuredDisks.IsEmpty() {
		// No explicitly-configured disks.
		propertyHelper.SetDisks(actualDisks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("No disks configured; storage controller '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

		return nil
	}

	return updateStorageControllerDisks(data, providerState)
}

// Delete a storage controller resource.
func resourceStorageControllerDelete(data *schema.ResourceData, provider interface{}) error {
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	targetController, err := getStorageController(apiClient, data)
	if err != nil {
		return err
	}

	log.Printf("Delete storage controller '%s' in server '%s'.",
		controllerID,
		serverID,
	)

	removeDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)
	err = processRemoveStorageControllerDisks(removeDisks, data, providerState)
	if err != nil {
		return err
	}

	// We can't remove the SCSI adapter if it still has a disk (i.e. the last disk in the server).
	targetController, err = getStorageController(apiClient, data)
	if err != nil {
		return err
	}
	if len(targetController.Disks) > 0 {
		log.Printf("Treating storage controller '%s' in server '%s' as deleted because it still has one or more disks after disk-removal processing; this indicates that the minimum-disk-count-per-server limit has been exceeded.",
			controllerID,
			serverID,
		)

		return nil
	}

	operationDescription := fmt.Sprintf("Remove storage controller '%s' (SCSI bus %d) from server '%s'",
		targetController.ID,
		targetController.BusNumber,
		serverID,
	)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		var removeSCSIControllerError error
		removeSCSIControllerError = apiClient.RemoveSCSIControllerFromServer(controllerID)
		if compute.IsResourceBusyError(removeSCSIControllerError) {
			context.Retry()
		} else if removeSCSIControllerError != nil {
			context.Fail(removeSCSIControllerError)
		}
	})
	if err != nil {
		return err
	}

	log.Printf("Removing storage controller '%s' (SCSI bus %d) from server '%s'...",
		targetController.ID,
		targetController.BusNumber,
		serverID,
	)

	_, err = apiClient.WaitForChange(
		compute.ResourceTypeServer,
		serverID,
		fmt.Sprintf("Remove SCSI controller '%s' (bus %d)", targetController.ID, targetController.BusNumber),
		resourceUpdateTimeoutServer,
	)
	if err != nil {
		return err
	}

	log.Printf("Removed storage controller '%s' (SCSI bus %d) from server '%s'.",
		targetController.ID,
		targetController.BusNumber,
		serverID,
	)

	return nil
}

// Import data for an existing storage controller.
func resourceStorageControllerImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	storageControllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)
	log.Printf("Import storage controller '%s' in server '%s'.", storageControllerID, serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return
	}
	if server == nil {
		err = fmt.Errorf("Server '%s' not found", serverID)

		return
	}
	storageController := server.SCSIControllers.GetByID(storageControllerID)
	if storageController == nil {
		err = fmt.Errorf("Storage controller '%s' not found in server '%s'", storageControllerID, serverID)

		return
	}

	data.Set(resourceKeyStorageControllerBusNumber, storageController.BusNumber)
	data.Set(resourceKeyStorageControllerAdapterType, storageController.AdapterType)
	propertyHelper(data).SetDisks(
		models.NewDisksFromVirtualMachineSCSIController(*storageController),
	)

	importedData = []*schema.ResourceData{data}

	return
}

// When updating a storage controller resource, synchronise the controller's disk attributes with its resource data
func updateStorageControllerDisks(data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	controllerID := data.Id()
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	log.Printf("Configure disks for storage controller '%s' in server '%s'...", controllerID, serverID)

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
	log.Printf("Storage controller '%s' currently has %d disks: %#v.", controllerID, len(actualDisks), actualDisks)

	configuredDisks := propertyHelper.GetDisks()
	log.Printf("Configuration for storage controller '%s' (bus %d) specifies %d disks: %#v.", controllerID, targetController.BusNumber, len(configuredDisks), configuredDisks)

	err = validateStorageControllerDisks(configuredDisks)
	if err != nil {
		return err
	}

	addDisks, modifyDisks, removeDisks := configuredDisks.SplitByAction(actualDisks)
	if addDisks.IsEmpty() && modifyDisks.IsEmpty() && removeDisks.IsEmpty() {
		log.Printf("No changes required for disks of storage controller '%s' in server '%s'.", controllerID, serverID)

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
	targetController, err := getStorageController(apiClient, data)
	if err != nil {
		return err
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
			fmt.Sprintf("Add disk %d:%d", addDisk.SCSIBusNumber, addDisk.SCSIUnitID),
			resourceUpdateTimeoutServer,
		)
		if err != nil {
			return err
		}

		server := resource.(*compute.Server)
		targetController := server.SCSIControllers.GetByID(controllerID)
		if targetController == nil {
			return fmt.Errorf("cannot find controller '%s' in server '%s'", controllerID, serverID)
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
	targetController, err := getStorageController(apiClient, data)
	if err != nil {
		return err
	}

	actualDisks := models.NewDisksFromVirtualMachineSCSIController(*targetController)
	actualDisksBySCSIPath := actualDisks.BySCSIPath()

	for index := range modifyDisks {
		modifyDisk := &modifyDisks[index]
		actualDisk := actualDisksBySCSIPath[modifyDisk.SCSIPath()]

		// Can't shrink disk, only grow it.
		if modifyDisk.SizeGB < actualDisk.SizeGB {
			return fmt.Errorf(
				"cannot resize disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' from %d to GB to %d (for now, disks can only be expanded)",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}

		// Do we need to expand the disk?
		if modifyDisk.SizeGB > actualDisk.SizeGB {
			log.Printf(
				"Expanding disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from %d GB to %d GB)...",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.SizeGB,
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
				"Expand disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from %d to GB to %d)...",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.SizeGB,
				modifyDisk.SizeGB,
			)

			resource, err := apiClient.WaitForChange(
				compute.ResourceTypeServer,
				serverID,
				fmt.Sprintf("Expand disk %d:%d", modifyDisk.SCSIBusNumber, modifyDisk.SCSIUnitID),
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			modifyDisk.SizeGB = actualDisk.SizeGB

			server := resource.(*compute.Server)
			targetController := server.SCSIControllers.GetByID(controllerID)
			if targetController == nil {
				return fmt.Errorf("cannot find controller '%s' in server '%s'", controllerID, serverID)
			}
			propertyHelper.SetDisks(
				models.NewDisksFromVirtualMachineSCSIController(*targetController),
			)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf("storage controller '%s' now has %d disks: %#v.", targetController.ID, len(targetController.Disks), targetController)

			log.Printf(
				"Expanded disk '%s' on storage controller '%s' (SCSI bus %d) server '%s' (from %d to GB to %d).",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}

		// Do we need to change the disk speed?
		if modifyDisk.Speed != actualDisk.Speed {
			log.Printf(
				"Changing speed of disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from '%s' to '%s')...",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.Speed,
				modifyDisk.Speed,
			)

			operationDescription := fmt.Sprintf("Change speed of disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from '%s' to '%s')",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.Speed,
				modifyDisk.Speed,
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
						"Unexpected result '%s' when changing speed of disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s'.",
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
				fmt.Sprintf("Change speed of disk %d:%d", modifyDisk.SCSIBusNumber, modifyDisk.SCSIUnitID),
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			modifyDisk.Speed = actualDisk.Speed

			server := resource.(*compute.Server)
			targetController := server.SCSIControllers.GetByID(controllerID)
			if targetController == nil {
				return fmt.Errorf("cannot find controller '%s' in server '%s'", controllerID, serverID)
			}
			propertyHelper.SetDisks(
				models.NewDisksFromVirtualMachineSCSIController(*targetController),
			)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf("Changed speed of disk '%s' on storage controller '%s' (SCSI bus %d) in server '%s' (from '%s' to GB to '%s').",
				modifyDisk.ID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
				actualDisk.Speed,
				modifyDisk.Speed,
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
	targetController, err := getStorageController(apiClient, data)
	if err != nil {
		return err
	}

	for _, removeDisk := range removeDisks {
		log.Printf("Remove disk '%s' (SCSI unit Id %d) from storage controller '%s' (SCSI bus %d) in server '%s'...",
			removeDisk.ID,
			removeDisk.SCSIUnitID,
			targetController.ID,
			targetController.BusNumber,
			serverID,
		)

		removingDisk := true
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
			} else if compute.IsExceedsLimitError(removeError) {
				removingDisk = false // It's the server's last disk, so we won't actually perform the removal
			} else if removeError != nil {
				context.Fail(removeError)
			}
		})
		if err != nil {
			return err
		}

		var server *compute.Server
		if removingDisk {
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
		} else {
			log.Printf("Not removing disk '%s' (SCSI unit Id %d) from storage controller '%s' (SCSI bus %d) in server '%s' because this is the server's last disk.",
				removeDisk.ID,
				removeDisk.SCSIUnitID,
				targetController.ID,
				targetController.BusNumber,
				serverID,
			)

			server, err = apiClient.GetServer(serverID)
			if err != nil {
				return err
			}
			if server == nil {
				log.Printf("Cannot find server '%s'; will treat storage controller '%s' as deleted.", serverID, controllerID)

				return nil
			}
		}

		targetController := server.SCSIControllers.GetByID(controllerID)
		if targetController == nil {
			return fmt.Errorf("cannot find controller '%s' in server '%s'", controllerID, serverID)
		}
		propertyHelper.SetDisks(
			models.NewDisksFromVirtualMachineSCSIController(*targetController),
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

func getStorageController(apiClient *compute.Client, data *schema.ResourceData) (*compute.VirtualMachineSCSIController, error) {
	controllerID := data.Id()

	serverStorageControllers, err := getServerStorageControllers(apiClient, data)
	if err != nil {
		return nil, err
	}

	targetController := serverStorageControllers.GetByID(controllerID)
	if targetController == nil {
		serverID := data.Get(resourceKeyStorageControllerServerID).(string)

		return nil, fmt.Errorf("cannot find storage controller '%s' in server '%s'", controllerID, serverID)
	}

	return targetController, nil
}

func getServerStorageControllers(apiClient *compute.Client, data *schema.ResourceData) (compute.VirtualMachineSCSIControllers, error) {
	serverID := data.Get(resourceKeyStorageControllerServerID).(string)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return nil, err
	}
	if server == nil {
		data.SetId("")

		return nil, fmt.Errorf("server '%s' has been deleted", serverID)
	}

	return server.SCSIControllers, nil
}
