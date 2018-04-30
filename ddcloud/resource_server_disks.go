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
	resourceKeyServerDisk          = "disk"
	resourceKeyServerDiskID        = "id"
	resourceKeyServerDiskBusNumber = "scsi_bus_number"
	resourceKeyServerDiskUnitID    = "scsi_unit_id"
	resourceKeyServerDiskSizeGB    = "size_gb"
	resourceKeyServerDiskSpeed     = "speed"
)

func schemaDisk() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Default:     nil,
		Description: "The set of virtual disks attached to a server or storage controller",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerDiskID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The CloudControl identifier for the virtual disk (computed when the disk is first created)",
				},
				resourceKeyServerDiskBusNumber: &schema.Schema{
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The SCSI bus number for the disk",
				},
				resourceKeyServerDiskUnitID: &schema.Schema{
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The SCSI Logical Unit Number (LUN) for the disk",
				},
				resourceKeyServerDiskSizeGB: &schema.Schema{
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The size (in GB) of the disk",
				},
				resourceKeyServerDiskSpeed: &schema.Schema{
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "STANDARD",
					StateFunc:    normalizeSpeed,
					Description:  "The disk speed",
					ValidateFunc: validateDiskSpeed,
				},
			},
		},
	}
}

// When creating a server resource, synchronise the server's disks with its resource data.
//
// If the server is running, then it will be stopped before creating / updating disks, and then restarted.
func createDisks(server *compute.Server, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Id()

	log.Printf("Configuring image disks for server '%s'...", serverID)

	// Since this is the first time, populate image disks.
	configuredDisks := propertyHelper.GetDisks()
	log.Printf("Configuration for server '%s' specifies %d disks: %#v.", serverID, len(configuredDisks), configuredDisks)
	actualDisks := models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers)
	configuredDisks.CaptureIDs(actualDisks)

	err := validateServerDisks(configuredDisks)
	if err != nil {
		return err
	}

	if len(configuredDisks) == 0 {
		log.Printf("Configuration for server '%s' does not specify any disks; the provider will assume all disks are either image-only, or configured via ddcloud_storage_controller.", serverID)

		propertyHelper.SetDisks(actualDisks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(actualDisks), actualDisks)

		return nil
	}

	log.Printf("Configuration for server '%s' specifies %d disks: %#v.", serverID, len(configuredDisks), configuredDisks)

	propertyHelper.SetDisks(configuredDisks)
	propertyHelper.SetPartial(resourceKeyServerDisk)

	apiClient := providerState.Client()

	server, err = apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server with Id '%s'", serverID)
	}

	// After initial server deployment, we only need to handle disks that were part of the original server image (and of those, only ones we need to modify after the initial deployment completed deployment).
	log.Printf("Configure image disks for server '%s'...", serverID)
	actualDisks = models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers)
	addDisks, modifyDisks, _ := configuredDisks.SplitByAction(actualDisks) // Ignore removeDisks since not all disks have been created yet
	if addDisks.IsEmpty() && modifyDisks.IsEmpty() {
		log.Printf("No post-deploy changes required for disks of server '%s'.", serverID)

		return nil
	}

	serverWasStarted := server.Started
	if serverWasStarted {
		log.Printf("Shutting down server '%s' ('%s') before modifying disk configuration...",
			server.Name,
			server.ID,
		)

		err = serverShutdown(providerState, serverID)
		if err != nil {
			return err
		}

		log.Printf("Shutdown complete for server '%s' ('%s').",
			server.Name,
			server.ID,
		)
	}

	err = processModifyDisks(modifyDisks, data, providerState)
	if err != nil {
		return err
	}

	log.Printf("Configure additional disks for server '%s'...", serverID)
	err = processAddDisks(addDisks, data, providerState)
	if err != nil {
		return err
	}

	if serverWasStarted {
		log.Printf("Restarting server '%s' ('%s') after modifying disk configuration...",
			server.Name,
			server.ID,
		)

		err = serverStart(providerState, serverID)
		if err != nil {
			return err
		}

		log.Printf("Restart complete for server '%s' ('%s').",
			server.Name,
			server.ID,
		)
	}

	return nil
}

// When updating a server resource, synchronise the server's image disk attributes with its resource data
// Removes image disks from existingDisksByUnitID as they are processed, leaving only additional disks.
//
// If the server is running, then it will be stopped before creating / updating disks, and then restarted.
func updateDisks(data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Id()

	log.Printf("Configure image disks for server '%s'...", serverID)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}
	actualDisks := models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers)

	configuredDisks := propertyHelper.GetDisks()
	log.Printf("Configuration for server '%s' specifies %d disks: %#v.", serverID, len(configuredDisks), configuredDisks)

	// Detect switch to / from inline disks.
	previouslyConfiguredDisks := propertyHelper.GetOldDisks()
	if previouslyConfiguredDisks.IsEmpty() && !configuredDisks.IsEmpty() {
		// Switched to inline declaration of server disks.
		log.Printf("Disks for server '%s' are now configured inline, but were previously unspecified or configured via one or more ddcloud_storage_controllers.",
			serverID,
		)

		// Can only switch to inline disks if there is a single (default) SCSI controller.
		if len(server.SCSIControllers) > 1 {
			return fmt.Errorf("server '%s' has multiple SCSI controllers, so disks cannot be configured inline",
				serverID,
			)
		}
	} else if !previouslyConfiguredDisks.IsEmpty() && configuredDisks.IsEmpty() {
		// Switched to external declaration of server disks.

		log.Printf("Disks for server '%s' are now unspecified or configured via one or more ddcloud_storage_controllers, but were previously configured inline.",
			serverID,
		)
	}

	err = validateServerDisks(configuredDisks)
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

	serverWasStarted := server.Started
	if serverWasStarted {
		log.Printf("Shutting down server '%s' ('%s') before modifying disk configuration...",
			server.Name,
			server.ID,
		)

		err = serverShutdown(providerState, serverID)
		if err != nil {
			return err
		}

		log.Printf("Shutdown complete for server '%s' ('%s').",
			server.Name,
			server.ID,
		)
	}

	// First remove any disks that are no longer required.
	err = processRemoveDisks(removeDisks, data, providerState)
	if err != nil {
		return err
	}

	// Then modify existing disks
	err = processModifyDisks(modifyDisks, data, providerState)
	if err != nil {
		return err
	}

	// Finally, add new disks
	err = processAddDisks(addDisks, data, providerState)
	if err != nil {
		return err
	}

	if serverWasStarted {
		log.Printf("Restarting server '%s' ('%s') after modifying disk configuration...",
			server.Name,
			server.ID,
		)

		err = serverStart(providerState, serverID)
		if err != nil {
			return err
		}

		log.Printf("Restart complete for server '%s' ('%s').",
			server.Name,
			server.ID,
		)
	}

	return nil
}

// Process the collection of disks that need to be added to the server.
func processAddDisks(addDisks models.Disks, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Id()

	apiClient := providerState.Client()

	for index := range addDisks {
		addDisk := &addDisks[index]

		operationDescription := fmt.Sprintf("Add disk with SCSI unit ID %d to server '%s'",
			addDisk.SCSIUnitID,
			serverID,
		)
		err := providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			var addDiskError error
			addDisk.ID, addDiskError = apiClient.AddDiskToServer(
				serverID,
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

		log.Printf("Adding disk '%s' (%dGB, speed = '%s') with SCSI unit ID %d to server '%s'...",
			addDisk.ID,
			addDisk.SizeGB,
			addDisk.Speed,
			addDisk.SCSIUnitID,
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
		propertyHelper.SetDisks(
			models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers),
		)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

		log.Printf("Added disk '%s' with SCSI unit ID %d to server '%s'.",
			addDisk.ID,
			addDisk.SCSIUnitID,
			serverID,
		)
	}

	return nil
}

// Process the collection of disks whose configuration needs to be modified.
//
// Disk Ids must already be populated.
func processModifyDisks(modifyDisks models.Disks, data *schema.ResourceData, providerState *providerState) error {
	log.Printf("modifyDisks = %#v", modifyDisks)

	propertyHelper := propertyHelper(data)
	serverID := data.Id()

	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}
	actualDisks := models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers)
	actualDisksBySCSIPath := actualDisks.BySCSIPath()

	for index := range modifyDisks {
		modifyDisk := &modifyDisks[index]
		log.Printf("modifyDisk = %#v", modifyDisk)
		actualImageDisk := actualDisksBySCSIPath[modifyDisk.SCSIPath()]

		// Can't shrink disk, only grow it.
		if modifyDisk.SizeGB < actualImageDisk.SizeGB {
			return fmt.Errorf(
				"cannot resize disk '%s' in server '%s' from %d to GB to %d (for now, disks can only be expanded)",
				modifyDisk.ID,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}

		// Do we need to expand the disk?
		if modifyDisk.SizeGB > actualImageDisk.SizeGB {
			log.Printf(
				"Expanding disk '%s' in server '%s' (from %d GB to %d GB)...",
				modifyDisk.ID,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)

			operationDescription := fmt.Sprintf("Expand disk '%s' in server '%s'", modifyDisk.ID, serverID)
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
				"Resizing disk '%s' for server '%s' (from %d to GB to %d)...",
				modifyDisk.ID,
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

			log.Printf("Server '%s' now has %d disks: %#v.", serverID, server.SCSIControllers.GetDiskCount(), server.SCSIControllers)

			log.Printf(
				"Resized disk '%s' for server '%s' (from %d to GB to %d).",
				modifyDisk.ID,
				serverID,
				actualImageDisk.SizeGB,
				modifyDisk.SizeGB,
			)
		}

		// Do we need to change the disk speed?
		if modifyDisk.Speed != actualImageDisk.Speed {
			log.Printf(
				"Changing speed of disk '%s' in server '%s' (from '%s' to '%s')...",
				modifyDisk.ID,
				serverID,
				actualImageDisk.Speed,
				modifyDisk.Speed,
			)

			operationDescription := fmt.Sprintf("Change speed of disk '%s' in server '%s'", modifyDisk.ID, serverID)
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
						"Unexpected result '%s' when resizing server disk '%s' for server '%s'.",
						response.Result,
						modifyDisk.ID,
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

			log.Printf(
				"Resized disk '%s' for server '%s' (from %d to GB to %d).",
				modifyDisk.ID,
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
func processRemoveDisks(removeDisks models.Disks, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Id()

	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("server '%s' has been deleted", serverID)
	}

	for _, removeDisk := range removeDisks {
		log.Printf("Remove disk '%s' (SCSI unit Id %d) from server '%s'...",
			removeDisk.ID,
			removeDisk.SCSIUnitID,
			serverID,
		)

		operationDescription := fmt.Sprintf("Remove disk '%s' from server '%s'", removeDisk.ID, serverID)
		err = providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			removeError := apiClient.RemoveDiskFromServer(removeDisk.ID)
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

		server := resource.(*compute.Server)
		propertyHelper.SetDisks(
			models.NewDisksFromVirtualMachineSCSIControllers(server.SCSIControllers),
		)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf(
			"Removed disk '%s' from server '%s'.",
			removeDisk.ID,
			serverID,
		)
	}

	return nil
}

func hashDiskUnitID(item interface{}) int {
	disk, ok := item.(models.Disk)
	if ok {
		return disk.SCSIUnitID
	}

	virtualMachineDisk, ok := item.(compute.VirtualMachineDisk)
	if ok {
		return virtualMachineDisk.SCSIUnitID
	}

	diskData := item.(map[string]interface{})

	return diskData[resourceKeyServerDiskUnitID].(int)
}

func validateDisks(disks models.Disks) error {
	if disks.IsEmpty() {
		return nil
	}

	disksByUnitID := make(map[int]models.Disk)
	for _, disk := range disks {
		_, duplicate := disksByUnitID[disk.SCSIUnitID]
		if duplicate {
			return fmt.Errorf("multiple disks with SCSI unit ID %d", disk.SCSIUnitID)
		}

		disksByUnitID[disk.SCSIUnitID] = disk
	}

	return nil
}

func validateServerDisks(disks models.Disks) error {
	err := validateDisks(disks)
	if err != nil {
		return err
	}

	for _, disk := range disks {
		// Cannot target non-default SCSI controllers if disks are declared directly on the server.
		if disk.SCSIBusNumber != 0 {
			return fmt.Errorf("unsupported configuration: disk configured to use non-default SCSI bus %d (declare the disk on a ddcloud_storage controller if targeting a SCSI bus other than 0)", disk.SCSIBusNumber)
		}
	}

	return nil
}

func validateDiskSpeed(value interface{}, propertyName string) (messages []string, errors []error) {
	if value == nil {
		return
	}

	adapterType, ok := value.(string)
	if !ok {
		errors = append(errors,
			fmt.Errorf("Unexpected value type '%v'", value),
		)

		return
	}

	switch adapterType {
	case compute.ServerDiskSpeedEconomy:
	case compute.ServerDiskSpeedStandard:
	case compute.ServerDiskSpeedHighPerformance:
		break
	default:
		errors = append(errors,
			fmt.Errorf("invalid disk speed '%s'", value),
		)
	}

	return
}
