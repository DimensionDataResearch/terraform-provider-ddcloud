package models

import (
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/maps"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// ServerDiskAction represents an action to be taken for a ServerDisk.
type ServerDiskAction int

const (
	// ServerDiskActionNone indicates that no action is to be taken.
	ServerDiskActionNone ServerDiskAction = iota

	// ServerDiskActionCreate indicates that the ServerDisk is to be created.
	ServerDiskActionCreate

	// ServerDiskActionUpdate indicates that the ServerDisk is to be updated.
	ServerDiskActionUpdate

	// ServerDiskActionDelete indicates that the ServerDisk is to be deleted.
	ServerDiskActionDelete
)

// ServerDisk represents the Terraform configuration for a ddcloud_server virtual_disk.
type ServerDisk struct {
	ID         *string
	SCSIUnitID int
	SizeGB     int
	Speed      string
	Action     ServerDiskAction
}

// ReadMap populates the ServerDisk with values from the specified map.
func (disk *ServerDisk) ReadMap(diskProperties map[string]interface{}) {
	reader := maps.NewReader(diskProperties)

	disk.ID = reader.GetStringPtr("id")
	disk.SCSIUnitID = reader.GetInt("scsi_unit_id")
	disk.SizeGB = reader.GetInt("size_gb")
	disk.Speed = reader.GetString("speed")
}

// ToMap creates a new map using the values from the ServerDisk.
func (disk *ServerDisk) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	disk.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the ServerDisk.
func (disk *ServerDisk) UpdateMap(diskProperties map[string]interface{}) {
	writer := maps.NewWriter(diskProperties)

	writer.SetStringPtr("id", disk.ID)
	writer.SetInt("scsi_unit_id", disk.SCSIUnitID)
	writer.SetInt("size_gb", disk.SizeGB)
	writer.SetString("speed", disk.Speed)
}

// ReadVirtualMachineDisk populates the ServerDisk with values from the specified VirtualMachineDisk.
func (disk *ServerDisk) ReadVirtualMachineDisk(virtualMachineDisk compute.VirtualMachineDisk) {
	disk.ID = virtualMachineDisk.ID
	disk.SCSIUnitID = virtualMachineDisk.SCSIUnitID
	disk.SizeGB = virtualMachineDisk.SizeGB
	disk.Speed = virtualMachineDisk.Speed
}

// ToVirtualMachineDisk updates a map using values from the ServerDisk.
func (disk *ServerDisk) ToVirtualMachineDisk() compute.VirtualMachineDisk {
	virtualMachineDisk := compute.VirtualMachineDisk{}
	disk.UpdateVirtualMachineDisk(&virtualMachineDisk)

	return virtualMachineDisk
}

// UpdateVirtualMachineDisk updates a CloudControl VirtualMachineDisk using values from the ServerDisk.
func (disk *ServerDisk) UpdateVirtualMachineDisk(virtualMachineDisk *compute.VirtualMachineDisk) {
	virtualMachineDisk.ID = disk.ID
	virtualMachineDisk.SCSIUnitID = disk.SCSIUnitID
	virtualMachineDisk.SizeGB = disk.SizeGB
	virtualMachineDisk.Speed = disk.Speed
}

// NewServerDiskFromMap creates a ServerDisk from the values in the specified map.
func NewServerDiskFromMap(diskProperties map[string]interface{}) ServerDisk {
	disk := ServerDisk{}
	disk.ReadMap(diskProperties)

	return disk
}

// NewServerDiskFromVirtualMachineDisk creates a ServerDisk from the values in the specified CloudControl VirtualMachineDisk.
func NewServerDiskFromVirtualMachineDisk(virtualMachineDisk compute.VirtualMachineDisk) ServerDisk {
	disk := ServerDisk{}
	disk.ReadVirtualMachineDisk(virtualMachineDisk)

	return disk
}

// NewServerDisksFromStateData creates ServerDisks from an array of Terraform state data.
//
// The values in the diskPropertyList are expected to be map[string]interface{}.
func NewServerDisksFromStateData(diskPropertyList []interface{}) ServerDisks {
	disks := make(ServerDisks, len(diskPropertyList))
	for index, data := range diskPropertyList {
		diskProperties := data.(map[string]interface{})
		disks[index] = NewServerDiskFromMap(diskProperties)
	}

	return disks
}

// NewServerDisksFromMaps creates ServerDisks from an array of Terraform value maps.
func NewServerDisksFromMaps(diskPropertyList []map[string]interface{}) ServerDisks {
	disks := make(ServerDisks, len(diskPropertyList))
	for index, data := range diskPropertyList {
		disks[index] = NewServerDiskFromMap(data)
	}

	return disks
}

// NewServerDisksFromVirtualMachineDisks creates ServerDisks from an array of compute.VirtualMachineDisk.
func NewServerDisksFromVirtualMachineDisks(virtualMachineDisks []compute.VirtualMachineDisk) ServerDisks {
	disks := make(ServerDisks, len(virtualMachineDisks))
	for index, virtualMachineDisk := range virtualMachineDisks {
		disks[index] = NewServerDiskFromVirtualMachineDisk(virtualMachineDisk)
	}

	return disks
}
