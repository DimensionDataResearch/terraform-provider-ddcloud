package models

import (
	"sort"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// TODO: Consider implementing Disks.CalculateActions([]compute.VirtualMachineDisk)

// Disks represents an array of Disk structures.
type Disks []Disk

// IsEmpty determines whether the Disk array is empty.
func (disks Disks) IsEmpty() bool {
	return len(disks) == 0
}

// SortBySCSIPath sorts the disks by SCSI bus number and then SCSI unit Id.
func (disks Disks) SortBySCSIPath() {
	sorter := &diskSorter{
		Disks: disks,
	}

	sort.Sort(sorter)
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

// BySCSIPath creates a map of Disk keyed by SCSI unit Id.
func (disks Disks) BySCSIPath() map[string]Disk {
	disksBySCSIPath := make(map[string]Disk)
	for _, disk := range disks {
		disksBySCSIPath[disk.SCSIPath()] = disk
	}

	return disksBySCSIPath
}

// CaptureIDs updates the Disk Ids from the actual disks.
func (disks Disks) CaptureIDs(actualDisks Disks) {
	actualDisksBySCSIPath := actualDisks.BySCSIPath()
	for index := range disks {
		disk := &disks[index]
		actualDisk, ok := actualDisksBySCSIPath[disk.SCSIPath()]
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
	currentDisksBySCSIPath := currentDisks.BySCSIPath()

	appliedDisks := make(Disks, 0)
	for index := range previousDisks {
		previousDisk := &previousDisks[index]
		currentDisk, ok := currentDisksBySCSIPath[previousDisk.SCSIPath()]
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
	actualDisksBySCSIPath := actualDisks.BySCSIPath()
	for _, configuredDisk := range disks {
		_, ok := actualDisksBySCSIPath[configuredDisk.SCSIPath()]
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
	actualDisksByUnitSCSIPath := actualDisks.BySCSIPath()
	for _, configuredDisk := range disks {
		actualDisk, ok := actualDisksByUnitSCSIPath[configuredDisk.SCSIPath()]

		// We don't want to see this disk when we're looking for disks that don't appear in the configuration.
		delete(actualDisksByUnitSCSIPath, configuredDisk.SCSIPath())

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
	for unconfiguredDiskUnitID := range actualDisksByUnitSCSIPath {
		unconfiguredDisk := actualDisksByUnitSCSIPath[unconfiguredDiskUnitID]
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

// NewDisksFromVirtualMachineSCSIController creates Disks from a compute.VirtualMachineSCSIController.
func NewDisksFromVirtualMachineSCSIController(virtualMachineSCSIController compute.VirtualMachineSCSIController) (disks Disks) {
	disks = make(Disks, len(virtualMachineSCSIController.Disks))
	for index, virtualMachineDisk := range virtualMachineSCSIController.Disks {
		disks[index] = NewDiskFromVirtualMachineDisk(virtualMachineDisk, virtualMachineSCSIController.BusNumber)
	}
	disks.SortBySCSIPath()

	return
}

// NewDisksFromVirtualMachineSCSIControllers creates Disks from compute.VirtualMachineSCSIControllers.
func NewDisksFromVirtualMachineSCSIControllers(virtualMachineSCSIControllers compute.VirtualMachineSCSIControllers) (disks Disks) {
	for _, virtualMachineSCSIController := range virtualMachineSCSIControllers {
		for _, virtualMachineDisk := range virtualMachineSCSIController.Disks {
			disks = append(disks,
				NewDiskFromVirtualMachineDisk(virtualMachineDisk, virtualMachineSCSIController.BusNumber),
			)
		}
	}
	disks.SortBySCSIPath()

	return
}

// diskSorter sorts disks by SCSI bus number and then by SCSI unit Id
type diskSorter struct {
	Disks Disks
}

func (sorter diskSorter) Len() int {
	return len(sorter.Disks)
}

func (sorter diskSorter) Less(index1 int, index2 int) bool {
	disk1 := sorter.Disks[index1]
	disk1SortKey := disk1.SCSIBusNumber*1000 + disk1.SCSIUnitID

	disk2 := sorter.Disks[index2]
	disk2SortKey := disk2.SCSIBusNumber*1000 + disk2.SCSIUnitID

	return disk1SortKey < disk2SortKey
}

func (sorter diskSorter) Swap(index1 int, index2 int) {
	temp := sorter.Disks[index1]
	sorter.Disks[index1] = sorter.Disks[index2]
	sorter.Disks[index2] = temp
}

var _ sort.Interface = &diskSorter{}
