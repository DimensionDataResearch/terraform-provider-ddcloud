package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerDisk       = "disk"
	resourceKeyServerDiskID     = "disk_id"
	resourceKeyServerDiskUnitID = "scsi_unit_id"
	resourceKeyServerDiskSizeGB = "size_gb"
	resourceKeyServerDiskSpeed  = "speed"
	// TODO: Consider adding "disk_type" property ("image" or "additional")
)

func schemaServerDisk() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Computed:    true,
		Default:     nil,
		Description: "The set of virtual disks attached to the server",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerDiskID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The CloudControl identifier for the virtual disk (computed when the disk is first created)",
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
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "STANDARD",
					StateFunc:   normalizeSpeed,
					Description: "The disk speed",
				},
			},
		},
		Set: hashDisk,
	}
}

// When creating a server resource, synchronise the server's disks with its resource data.
// imageDisks refers to the newly-deployed server's collection of disks (i.e. image disks).
func createDisks(imageDisks []compute.VirtualMachineDisk, data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)
	apiClient := providerState.Client()

	serverID := data.Id()

	log.Printf("Configuring image disks for server '%s'...", serverID)

	configuredDisks := propertyHelper.GetServerDisks()
	log.Printf("Configuration for server '%s' specifies %d disks: %#v.", serverID, len(configuredDisks), configuredDisks)

	if len(configuredDisks) == 0 {
		// Since this is the first time, populate image disks.
		var serverDisks []compute.VirtualMachineDisk
		for _, disk := range configuredDisks {
			serverDisks = append(serverDisks, disk)
		}

		propertyHelper.SetServerDisks(serverDisks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(serverDisks), serverDisks)

		return nil
	}

	var server *compute.Server
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("Cannot find server with Id '%s'", serverID)
	}

	// First, handle disks that were part of the original server image.
	log.Printf("Configure image disks for server '%s'...", serverID)

	configuredDisksByUnitID := getDisksByUnitID(configuredDisks)
	for _, actualImageDisk := range imageDisks {
		configuredImageDisk, ok := configuredDisksByUnitID[actualImageDisk.SCSIUnitID]
		if !ok {
			// This is not an image disk.
			log.Printf("No configuration was found for disk with SCSI unit Id %d for server '%s'; this disk will be treated as an additional disk.", actualImageDisk.SCSIUnitID, serverID)

			continue
		}

		// This is an image disk, so we don't want to see it when we're configuring additional disks
		delete(configuredDisksByUnitID, configuredImageDisk.SCSIUnitID)

		imageDiskID := *actualImageDisk.ID

		if configuredImageDisk.SizeGB < actualImageDisk.SizeGB {
			// Can't shrink disk, only grow it.
			return fmt.Errorf(
				"Cannot resize disk '%s' for server '%s' from %d to GB to %d (for now, disks can only be expanded).",
				imageDiskID,
				serverID,
				actualImageDisk.SizeGB,
				configuredImageDisk.SizeGB,
			)
		} else if configuredImageDisk.SizeGB > actualImageDisk.SizeGB {
			// We need to expand the disk.
			log.Printf(
				"Expanding disk '%s' for server '%s' (from %d GB to %d GB)...",
				imageDiskID,
				serverID,
				actualImageDisk.SizeGB,
				configuredImageDisk.SizeGB,
			)

			// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
			asyncLock := providerState.AcquireAsyncOperationLock("Expand disk '%s' for server '%s'", imageDiskID, serverID)
			defer asyncLock.Release()

			response, err := apiClient.ResizeServerDisk(serverID, imageDiskID, configuredImageDisk.SizeGB)
			if err != nil {
				return err
			}
			if response.Result != compute.ResultSuccess {
				return response.ToError(
					"Unexpected result '%s' when resizing server disk '%s' for server '%s'.",
					response.Result,
					imageDiskID,
					serverID,
				)
			}

			// Operation initiated; we no longer need this lock.
			asyncLock.Release()

			resource, err := apiClient.WaitForChange(
				compute.ResourceTypeServer,
				serverID,
				"Resize disk",
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			server = resource.(*compute.Server)
			propertyHelper.SetServerDisks(server.Disks)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf(
				"Resized disk '%s' for server '%s' (from %d to GB to %d).",
				imageDiskID,
				serverID,
				actualImageDisk.SizeGB,
				configuredImageDisk.SizeGB,
			)
		}

		if configuredImageDisk.Speed != actualImageDisk.Speed {
			// We need to change the disk speed.
			log.Printf(
				"Changing speed of disk '%s' in server '%s' (from '%s' to '%s')...",
				imageDiskID,
				serverID,
				actualImageDisk.Speed,
				configuredImageDisk.Speed,
			)

			// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
			asyncLock := providerState.AcquireAsyncOperationLock("Change speed of disk '%s' in server '%s'", imageDiskID, serverID)
			defer asyncLock.Release()

			response, err := apiClient.ChangeServerDiskSpeed(serverID, imageDiskID, configuredImageDisk.Speed)
			if err != nil {
				return err
			}
			if response.Result != compute.ResultSuccess {
				return response.ToError(
					"Unexpected result '%s' when changing speed of server disk '%s' in server '%s'.",
					response.Result,
					imageDiskID,
					serverID,
				)
			}

			// Operation initiated; we no longer need this lock.
			asyncLock.Release()

			resource, err := apiClient.WaitForChange(
				compute.ResourceTypeServer,
				serverID,
				"Change disk speed",
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			server = resource.(*compute.Server)
			propertyHelper.SetServerDisks(server.Disks)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf(
				"Changed speed of disk '%s' in server '%s' (from '%s' to '%s').",
				imageDiskID,
				serverID,
				actualImageDisk.Speed,
				configuredImageDisk.Speed,
			)
		}

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(server.Disks), server.Disks)
	}

	// By process of elimination, any remaining disks must be additional disks.
	log.Printf("Configure additional disks for server '%s'...", serverID)

	for additionalDiskUnitID := range configuredDisksByUnitID {
		log.Printf("Add disk with SCSI unit Id %d to server '%s'...", additionalDiskUnitID, serverID)

		configuredAdditionalDisk := configuredDisksByUnitID[additionalDiskUnitID]

		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Add disk with SCSI unit Id %d to server '%s'", additionalDiskUnitID, serverID)
		defer asyncLock.Release()

		var additionalDiskID string
		additionalDiskID, err = apiClient.AddDiskToServer(
			serverID,
			configuredAdditionalDisk.SCSIUnitID,
			configuredAdditionalDisk.SizeGB,
			configuredAdditionalDisk.Speed,
		)
		if err != nil {
			return err
		}

		// Operation initiated; we no longer need this lock.
		asyncLock.Release()

		log.Printf("Adding disk '%s' (%dGB, speed = '%s') with SCSI unit ID %d to server '%s'...",
			additionalDiskID,
			configuredAdditionalDisk.SizeGB,
			configuredAdditionalDisk.Speed,
			configuredAdditionalDisk.SCSIUnitID,
			serverID,
		)

		configuredAdditionalDisk.ID = &additionalDiskID

		var resource compute.Resource
		resource, err = apiClient.WaitForChange(
			compute.ResourceTypeServer,
			serverID,
			"Add disk",
			resourceUpdateTimeoutServer,
		)
		if err != nil {
			return err
		}

		server := resource.(*compute.Server)
		propertyHelper.SetServerDisks(server.Disks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(server.Disks), server.Disks)

		log.Printf("Added disk '%s' with SCSI unit ID %d to server '%s'.",
			additionalDiskID,
			configuredAdditionalDisk.SCSIUnitID,
			serverID,
		)
	}

	return nil
}

// When updating a server resource, synchronise the server's image disk attributes with its resource data
// Removes image disks from existingDisksByUnitID as they are processed, leaving only additional disks.
func updateDisks(data *schema.ResourceData, providerState *providerState) error {
	propertyHelper := propertyHelper(data)

	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	serverID := data.Id()

	log.Printf("Configure image disks for server '%s'...", serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		data.SetId("")

		return fmt.Errorf("Server '%s' has been deleted.", serverID)
	}

	configuredDisks := propertyHelper.GetServerDisks()
	log.Printf("Configuration for server '%s' specifies %d disks: %#v.", serverID, len(configuredDisks), configuredDisks)

	if len(configuredDisks) == 0 {
		// No explicitly-configured disks.
		propertyHelper.SetServerDisks(server.Disks)
		propertyHelper.SetPartial(resourceKeyServerDisk)

		log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(server.Disks), server.Disks)

		return nil
	}

	actualDisksByUnitID := getDisksByUnitID(server.Disks)
	for _, configuredDisk := range configuredDisks {
		actualDisk, ok := actualDisksByUnitID[configuredDisk.SCSIUnitID]

		// We don't want to see this disk when we're looking for disks that don't appear in the configuration.
		delete(actualDisksByUnitID, configuredDisk.SCSIUnitID)

		if ok {

			diskID := *actualDisk.ID

			// Existing disk.
			log.Printf("Examining existing disk '%s' with SCSI unit Id %d in server '%s'...", diskID, actualDisk.SCSIUnitID, serverID)

			if configuredDisk.SizeGB == actualDisk.SizeGB && configuredDisk.Speed == actualDisk.Speed {
				log.Printf("Disk '%s' with SCSI unit Id %d in server '%s' is up-to-date; nothing to do.", diskID, actualDisk.SCSIUnitID, serverID)

				continue // Nothing to do.
			}

			// We don't support changing disk speed yet.

			// Currently we can't shrink a disk, only grow it.
			if configuredDisk.SizeGB < actualDisk.SizeGB {
				log.Printf("Disk '%s' with SCSI unit Id %d in server '%s' is larger than the size specified in the server configuration; this is currently unsupported.", diskID, actualDisk.SCSIUnitID, serverID)

				return fmt.Errorf(
					"Cannot shrink disk '%s' for server '%s' from %d to GB to %d (for now, disks can only be expanded).",
					diskID,
					serverID,
					actualDisk.SizeGB,
					configuredDisk.SizeGB,
				)
			}

			// We need to expand the disk.
			log.Printf(
				"Expanding disk '%s' for server '%s' (from %d GB to %d GB)...",
				diskID,
				serverID,
				actualDisk.SizeGB,
				configuredDisk.SizeGB,
			)

			operationDescription := fmt.Sprintf("Expand disk '%s' in server '%s'", diskID, serverID)
			err = providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
				// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
				asyncLock := providerState.AcquireAsyncOperationLock("Expand disk '%s' in server '%s'", diskID, serverID)
				defer asyncLock.Release() // Release when the current attempt completes.

				response, err := apiClient.ResizeServerDisk(serverID, diskID, configuredDisk.SizeGB)
				if err != nil {
					context.Fail(err)
				} else if response.Result == compute.ResultResourceBusy {
					context.Retry()
				} else if response.Result != compute.ResultSuccess {
					context.Fail(
						response.ToError("Unexpected result '%s' when resizing server disk '%s' for server '%s'.", response.Result, diskID, serverID),
					)
				}

				asyncLock.Release()
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

			server := resource.(*compute.Server)

			propertyHelper.SetServerDisks(server.Disks)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(server.Disks), server.Disks)

			log.Printf(
				"Resized disk '%s' for server '%s' (from %d to GB to %d).",
				diskID,
				serverID,
				actualDisk.SizeGB,
				configuredDisk.SizeGB,
			)
		} else {
			// New disk.
			log.Printf("Adding disk with SCSI unit ID %d to server '%s'...", configuredDisk.SCSIUnitID, serverID)

			// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
			asyncLock := providerState.AcquireAsyncOperationLock("Add disk with SCSI unit Id %d to server '%s'", configuredDisk.SCSIUnitID, serverID)
			defer asyncLock.Release()

			var diskID string
			diskID, err = apiClient.AddDiskToServer(
				serverID,
				configuredDisk.SCSIUnitID,
				configuredDisk.SizeGB,
				configuredDisk.Speed,
			)
			if err != nil {
				return err
			}

			// Operation initiated; we no longer need this lock.
			asyncLock.Release()

			log.Printf("New disk '%s' has SCSI unit ID %d in server '%s'...", diskID, configuredDisk.SCSIUnitID, serverID)

			var resource compute.Resource
			resource, err = apiClient.WaitForChange(
				compute.ResourceTypeServer,
				serverID,
				"Add disk",
				resourceUpdateTimeoutServer,
			)
			if err != nil {
				return err
			}

			server := resource.(*compute.Server)
			propertyHelper.SetServerDisks(server.Disks)
			propertyHelper.SetPartial(resourceKeyServerDisk)

			log.Printf("Server '%s' now has %d disks: %#v.", serverID, len(server.Disks), server.Disks)

			log.Printf(
				"Added disk '%s' with SCSI unit ID %d to server '%s'.",
				diskID,
				configuredDisk.SCSIUnitID,
				serverID,
			)
		}
	}

	// By process of elimination, any remaining actual disks do not appear in the configuration and should be removed.
	for unconfiguredDiskUnitID := range actualDisksByUnitID {
		unconfiguredDisk := actualDisksByUnitID[unconfiguredDiskUnitID]

		unconfiguredDiskID := *unconfiguredDisk.ID

		log.Printf(
			"Disk '%s' does not appear in the configuration for server '%s' and will be removed.",
			unconfiguredDiskID,
			serverID,
		)

		// TODO: Implement server disk removal.
	}

	return nil
}

func getDisksByUnitID(disks []compute.VirtualMachineDisk) map[int]*compute.VirtualMachineDisk {
	disksByUnitID := make(map[int]*compute.VirtualMachineDisk)
	for index := range disks {
		disk := disks[index]
		disksByUnitID[disk.SCSIUnitID] = &disk
	}

	return disksByUnitID
}

func mergeAdditionalDisks(disksByUnitID map[int]*compute.VirtualMachineDisk, additionalDisks []compute.VirtualMachineDisk) {
	for _, disk := range additionalDisks {
		disksByUnitID[disk.SCSIUnitID] = &disk
	}
}

func hashDiskUnitID(item interface{}) int {
	disk, ok := item.(compute.VirtualMachineDisk)
	if ok {
		return disk.SCSIUnitID
	}

	diskData := item.(map[string]interface{})

	return diskData[resourceKeyServerDiskUnitID].(int)
}

func hashDisk(item interface{}) int {
	disk, ok := item.(compute.VirtualMachineDisk)
	if ok {
		return schema.HashString(fmt.Sprintf(
			"%d/%d/%s",
			disk.SCSIUnitID,
			disk.SizeGB,
			disk.Speed,
		))
	}

	diskData := item.(map[string]interface{})

	return schema.HashString(fmt.Sprintf(
		"%d/%d/%s",
		diskData[resourceKeyServerDiskUnitID].(int),
		diskData[resourceKeyServerDiskSizeGB].(int),
		diskData[resourceKeyServerDiskSpeed].(string),
	))
}

// Split an array of initially-configured disks by whether they represent image disks or additional disks.
//
// configuredDisks represents the disks currently specified in configuration.
// actualDisks represents the disks in the server, as returned by CloudControl.
//
// This function only works right after the server has been deployed (i.e. no post-deployment disk changes (such as AddDiskToServer) have been made).
func splitInitiallyConfiguredDisksByType(configuredDisks []compute.VirtualMachineDisk, actualDisks []compute.VirtualMachineDisk) (imageDisks []compute.VirtualMachineDisk, additionalDisks []compute.VirtualMachineDisk) {
	actualDisksByUnitID := getDisksByUnitID(actualDisks)
	for _, configuredDisk := range configuredDisks {
		_, ok := actualDisksByUnitID[configuredDisk.SCSIUnitID]
		if ok {
			// This is an image disk.
			imageDisks = append(imageDisks, configuredDisk)
		} else {
			// This is an additional disk
			additionalDisks = append(additionalDisks, configuredDisk)
		}
	}

	return
}

// Split an array of configured server disks by the action to be performed (add, change, or remove).
//
// configuredDisks represents the disks currently specified in configuration.
// actualDisks represents the disks in the server, as returned by CloudControl.
func splitConfiguredDisksByAction(configuredDisks []compute.VirtualMachineDisk, actualDisks []compute.VirtualMachineDisk) (addDisks []compute.VirtualMachineDisk, changeDisks []compute.VirtualMachineDisk, removeDisks []compute.VirtualMachineDisk) {
	actualDisksByUnitID := getDisksByUnitID(actualDisks)
	for _, configuredDisk := range configuredDisks {
		actualDisk, ok := actualDisksByUnitID[configuredDisk.SCSIUnitID]

		// We don't want to see this disk when we're looking for disks that don't appear in the configuration.
		delete(actualDisksByUnitID, configuredDisk.SCSIUnitID)

		if ok {
			// Existing disk.
			if configuredDisk.SizeGB != actualDisk.SizeGB {
				changeDisks = append(changeDisks, configuredDisk)
			} else if configuredDisk.Speed != actualDisk.Speed {
				changeDisks = append(changeDisks, configuredDisk)
			}
		} else {
			// New disk.
			addDisks = append(addDisks, configuredDisk)
		}
	}

	// By process of elimination, any remaining actual disks do not appear in the configuration and should be removed.
	for unconfiguredDiskUnitID := range actualDisksByUnitID {
		unconfiguredDisk := actualDisksByUnitID[unconfiguredDiskUnitID]
		removeDisks = append(removeDisks, *unconfiguredDisk)
	}

	return
}
