package models

import "github.com/DimensionDataResearch/go-dd-cloud-compute/compute"

// TODO: Consider implementing Disks.CalculateActions([]compute.VirtualMachineDisk)

// Disks represents an array of Disk structures.
type Disks []Disk

// IsEmpty determines whether the Disk array is empty.
func (disks Disks) IsEmpty() bool {
	return len(disks) == 0
}

// ToVirtualMachineDisks converts the Disks to an array of compute.VirtualMachineDisk.
func (disks Disks) ToVirtualMachineDisks() []compute.VirtualMachineDisk {
	virtualMachineDisks := make([]compute.VirtualMachineDisk, len(disks))
	for index, disk := range disks {
		virtualMachineDisks[index] = disk.ToVirtualMachineDisk()
	}

	return virtualMachineDisks
}

// ToMaps converts the Disks to an array of maps.
func (disks Disks) ToMaps() []map[string]interface{} {
	diskPropertyList := make([]map[string]interface{}, len(disks))
	for index, disk := range disks {
		diskPropertyList[index] = disk.ToMap()
	}

	return diskPropertyList
}

// ByUnitID creates a map of Disk keyed by SCSI unit Id.
func (disks Disks) ByUnitID() map[int]Disk {
	disksByUnitID := make(map[int]Disk)
	for _, disk := range disks {
		disksByUnitID[disk.SCSIUnitID] = disk
	}

	return disksByUnitID
}

// ByBusNumber creates a map of Disk keyed by SCSI bus number.
func (disks Disks) ByBusNumber() map[int][]Disk {
	disksByBusNumber := make(map[int][]Disk)
	for _, disk := range disks {
		disksForBusNumber, _ := disksByBusNumber[disk.SCSIBusNumber]
		disksForBusNumber = append(disksForBusNumber, disk)

		disksByBusNumber[disk.SCSIBusNumber] = disksForBusNumber
	}

	return disksByBusNumber
}

// CaptureIDs updates the Disk Ids from the actual disks.
func (disks Disks) CaptureIDs(actualDisks Disks) {
	actualDisksByUnitID := actualDisks.ByUnitID()
	for index := range disks {
		disk := &disks[index]
		actualDisk, ok := actualDisksByUnitID[disk.SCSIUnitID]
		if ok {
			disk.ID = actualDisk.ID
		}
	}
}

// ApplyCurrentConfiguration applies the current configuration, inline, to the old configuration.
//
// Call this function on the old Disks, passing the new Disks.
func (disks *Disks) ApplyCurrentConfiguration(currentDisks Disks) {
	previousDisks := *disks
	currentDisksByID := currentDisks.ByUnitID()

	appliedDisks := make(Disks, 0)
	for index := range previousDisks {
		previousDisk := &previousDisks[index]
		currentDisk, ok := currentDisksByID[previousDisk.SCSIUnitID]
		if !ok {
			continue // Disk no longer configured; leave it out.
		}

		// Update properties from current configuration.
		previousDisk.Speed = currentDisk.Speed
		previousDisk.SizeGB = currentDisk.SizeGB

		appliedDisks = append(appliedDisks, *previousDisk)
	}

	*disks = appliedDisks
}

// SplitByInitialType splits the (initially-configured) disks by whether they represent image disks or additional disks.
//
// configuredDisks represents the disks currently specified in configuration.
// actualDisks represents the disks in the server, as returned by CloudControl.
//
// This function only works right after the server has been deployed (i.e. no post-deployment disk changes (such as AddDiskToServer) have been made).
func (disks Disks) SplitByInitialType(actualDisks Disks) (imageDisks Disks, additionalDisks Disks) {
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
func (disks Disks) SplitByAction(actualDisks Disks) (addDisks Disks, changeDisks Disks, removeDisks Disks) {
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

// NewDisksFromStateData creates Disks from an array of Terraform state data.
//
// The values in the diskPropertyList are expected to be map[string]interface{}.
func NewDisksFromStateData(diskPropertyList []interface{}) Disks {
	disks := make(Disks, len(diskPropertyList))
	for index, data := range diskPropertyList {
		diskProperties := data.(map[string]interface{})
		disks[index] = NewDiskFromMap(diskProperties)
	}

	return disks
}

// NewDisksFromMaps creates Disks from an array of Terraform value maps.
func NewDisksFromMaps(diskPropertyList []map[string]interface{}) Disks {
	disks := make(Disks, len(diskPropertyList))
	for index, data := range diskPropertyList {
		disks[index] = NewDiskFromMap(data)
	}

	return disks
}

// NewDisksFromVirtualMachineDisks creates Disks from an array of compute.VirtualMachineDisk.
//
// AF: Broken by SCSI controller changes
//
func NewDisksFromVirtualMachineDisks(virtualMachineDisks []compute.VirtualMachineDisk) Disks {
	disks := make(Disks, len(virtualMachineDisks))
	for index, virtualMachineDisk := range virtualMachineDisks {
		disks[index] = NewDiskFromVirtualMachineDisk(virtualMachineDisk)
	}

	return disks
}

// NewDisksFromVirtualMachineSCSIControllers creates Disks from an array of compute.VirtualMachineSCSIController.
//
// Populates the SCSI bus number for each disk from its containing controller.
func NewDisksFromVirtualMachineSCSIControllers(virtualMachineSCSIControllers []compute.VirtualMachineSCSIController) Disks {
	var disks Disks

	for _, virtualMachineSCSIController := range virtualMachineSCSIControllers {
		for _, virtualMachineDisk := range virtualMachineSCSIController.Disks {
			disk := NewDiskFromVirtualMachineDisk(virtualMachineDisk)
			disk.SCSIBusNumber = virtualMachineSCSIController.BusNumber

			disks = append(disks, disk)
		}
	}

	return disks
}
