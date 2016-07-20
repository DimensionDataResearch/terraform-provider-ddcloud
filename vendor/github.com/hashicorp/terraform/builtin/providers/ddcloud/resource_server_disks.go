package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyServerDisk       = "disk"
	resourceKeyServerDiskID     = "disk_id"
	resourceKeyServerDiskSizeGB = "size_gb"
	resourceKeyServerDiskUnitID = "scsi_unit_id"
	resourceKeyServerDiskSpeed  = "speed"
)

func schemaServerDisk() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Default:  nil,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerDiskID: &schema.Schema{
					Type:     schema.TypeString,
					Computed: true,
				},
				resourceKeyServerDiskSizeGB: &schema.Schema{
					Type:     schema.TypeInt,
					Required: true,
				},
				resourceKeyServerDiskUnitID: &schema.Schema{
					Type:     schema.TypeInt,
					Required: true,
				},
				resourceKeyServerDiskSpeed: &schema.Schema{
					Type:      schema.TypeString,
					Optional:  true,
					Default:   "STANDARD",
					StateFunc: normalizeSpeed,
				},
			},
		},
		Set: hashDiskUnitID,
	}
}

// When creating a server resource, synchronise the server's disks with its resource data.
// imageDisks refers to the newly-deployed server's collection of disks (i.e. image disks).
func createDisks(imageDisks []compute.VirtualMachineDisk, data *schema.ResourceData, apiClient *compute.Client) (err error) {
	propertyHelper := propertyHelper(data)

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

		return
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

		if configuredImageDisk.SizeGB == actualImageDisk.SizeGB {
			continue // Nothing to do.
		}

		if configuredImageDisk.SizeGB < actualImageDisk.SizeGB {
			// Can't shrink disk, only grow it.
			err = fmt.Errorf(
				"Cannot resize disk '%s' for server '%s' from %d to GB to %d (for now, disks can only be expanded).",
				imageDiskID,
				serverID,
				actualImageDisk.SizeGB,
				configuredImageDisk.SizeGB,
			)

			return
		}

		// We need to expand the disk.
		log.Printf(
			"Expanding disk '%s' for server '%s' (from %d GB to %d GB)...",
			imageDiskID,
			serverID,
			actualImageDisk.SizeGB,
			configuredImageDisk.SizeGB,
		)

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
			imageDiskID,
			serverID,
			actualImageDisk.SizeGB,
			configuredImageDisk.SizeGB,
		)
	}

	// By process of elimination, any remaining disks must be additional disks.
	log.Printf("Configure additional disks for server '%s'...", serverID)

	for additionalDiskUnitID := range configuredDisksByUnitID {
		log.Printf("Add disk with SCSI unit ID %d to server '%s'...", additionalDiskUnitID, serverID)

		configuredAdditionalDisk := configuredDisksByUnitID[additionalDiskUnitID]

		var additionalDiskID string
		additionalDiskID, err = apiClient.AddDiskToServer(
			serverID,
			configuredAdditionalDisk.SCSIUnitID,
			configuredAdditionalDisk.SizeGB,
			configuredAdditionalDisk.Speed,
		)
		if err != nil {
			return
		}

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
			return
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
func updateDisks(data *schema.ResourceData, apiClient *compute.Client) error {
	propertyHelper := propertyHelper(data)

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
		// We don't want to see this disk when we're looking for disks that don't appear in the configuration.
		delete(actualDisksByUnitID, configuredDisk.SCSIUnitID)

		actualDisk, ok := actualDisksByUnitID[configuredDisk.SCSIUnitID]
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

			response, err := apiClient.ResizeServerDisk(serverID, diskID, configuredDisk.SizeGB)
			if err != nil {
				return err
			}
			if response.Result != compute.ResultSuccess {
				return response.ToError("Unexpected result '%s' when resizing server disk '%s' for server '%s'.", response.Result, diskID, serverID)
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
