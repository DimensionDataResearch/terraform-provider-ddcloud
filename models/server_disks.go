package models

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// TODO: Consider implementing ServerDisks.CalculateActions([]compute.VirtualMachineDisk)

// ServerDisks represents an array of ServerDisk structures.
type ServerDisks []ServerDisk

// ToVirtualMachineDisks converts the ServerDisks to an array of compute.VirtualMachineDisk.
func (disks ServerDisks) ToVirtualMachineDisks() []compute.VirtualMachineDisk {
	virtualMachineDisks := make([]compute.VirtualMachineDisk, len(disks))
	for index, disk := range disks {
		virtualMachineDisks[index] = disk.ToVirtualMachineDisk()
	}

	return virtualMachineDisks
}

// ToMaps converts the ServerDisks to an array of maps.
func (disks ServerDisks) ToMaps() []map[string]interface{} {
	diskPropertyList := make([]map[string]interface{}, len(disks))
	for index, disk := range disks {
		diskPropertyList[index] = disk.ToMap()
	}

	return diskPropertyList
}

// ByUnitID creates a map of ServerDisk keyed by SCSI unit Id.
func (disks ServerDisks) ByUnitID() map[int]ServerDisk {
	disksByUnitID := make(map[int]ServerDisk)
	for _, disk := range disks {
		disksByUnitID[disk.SCSIUnitID] = disk
	}

	return disksByUnitID
}

// SplitByInitialType splits the (initially-configured) disks by whether they represent image disks or additional disks.
//
// configuredDisks represents the disks currently specified in configuration.
// actualDisks represents the disks in the server, as returned by CloudControl.
//
// This function only works right after the server has been deployed (i.e. no post-deployment disk changes (such as AddDiskToServer) have been made).
func (disks ServerDisks) SplitByInitialType(actualDisks ServerDisks) (imageDisks ServerDisks, additionalDisks ServerDisks) {
	actualDisksByUnitID := actualDisks.ByUnitID()
	for _, configuredDisk := range disks {
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

// SplitByAction splits the (configured) server disks by the action to be performed (add, change, or remove).
//
// configuredDisks represents the disks currently specified in configuration.
// actualDisks represents the disks in the server, as returned by CloudControl.
func (disks ServerDisks) SplitByAction(actualDisks ServerDisks) (addDisks ServerDisks, changeDisks ServerDisks, removeDisks ServerDisks) {
	actualDisksByUnitID := actualDisks.ByUnitID()
	for _, configuredDisk := range disks {
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
		removeDisks = append(removeDisks, unconfiguredDisk)
	}

	return
}
