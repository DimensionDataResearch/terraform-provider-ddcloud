package models

import "github.com/DimensionDataResearch/go-dd-cloud-compute/compute"

// ServerDisk represents the Terraform configuration for a ddcloud_server virtual_disk.
type ServerDisk struct {
	ID         string
	SCSIUnitID int
	SizeGB     int
	Speed      string
}

// ReadMap populates the ServerDisk with values from the specified map.
func (disk *ServerDisk) ReadMap(data map[string]interface{}) {
	reader := reader(data)

	disk.ID = reader.String("id")
	disk.SCSIUnitID = reader.Int("scsi_unit_id")
	disk.SizeGB = reader.Int("size_gb")
	disk.Speed = reader.String("speed")
}

// CreateMap creates a new map using the values from the ServerDisk.
func (disk *ServerDisk) CreateMap() map[string]interface{} {
	data := make(map[string]interface{})
	disk.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the ServerDisk.
func (disk *ServerDisk) UpdateMap(data map[string]interface{}) {
	data["id"] = disk.ID
	data["scsi_unit_id"] = disk.SCSIUnitID
	data["size_gb"] = disk.SizeGB
	data["speed"] = disk.Speed
}

// ReadVirtualMachineDisk populates the ServerDisk with values from the specified VirtualMachineDisk.
func (disk *ServerDisk) ReadVirtualMachineDisk(virtualMachineDisk compute.VirtualMachineDisk) {
	if virtualMachineDisk.ID != nil {
		disk.ID = *virtualMachineDisk.ID
	} else {
		disk.ID = ""
	}
	disk.SCSIUnitID = virtualMachineDisk.SCSIUnitID
	disk.SizeGB = virtualMachineDisk.SizeGB
	disk.Speed = virtualMachineDisk.Speed
}

// CreateVirtualMachineDisk updates a map using values from the ServerDisk.
func (disk *ServerDisk) CreateVirtualMachineDisk() *compute.VirtualMachineDisk {
	virtualMachineDisk := &compute.VirtualMachineDisk{}
	disk.UpdateVirtualMachineDisk(virtualMachineDisk)

	return virtualMachineDisk
}

// UpdateVirtualMachineDisk updates a CloudControl VirtualMachineDisk using values from the ServerDisk.
func (disk *ServerDisk) UpdateVirtualMachineDisk(virtualMachineDisk *compute.VirtualMachineDisk) {
	if disk.ID != "" {
		virtualMachineDisk.ID = &disk.ID
	} else {
		virtualMachineDisk.ID = nil
	}
	virtualMachineDisk.SCSIUnitID = disk.SCSIUnitID
	virtualMachineDisk.SizeGB = disk.SizeGB
	virtualMachineDisk.Speed = disk.Speed
}

// ServerDiskFromMap creates a ServerDisk from the values in the specified map.
func ServerDiskFromMap(data map[string]interface{}) *ServerDisk {
	disk := &ServerDisk{}
	disk.ReadMap(data)

	return disk
}

// ServerDiskFromVirtualMachineDisk creates a ServerDisk from the values in the specified CloudControl VirtualMachineDisk.
func ServerDiskFromVirtualMachineDisk(virtualMachinedisk compute.VirtualMachineDisk) *ServerDisk {
	disk := &ServerDisk{}
	disk.ReadVirtualMachineDisk(virtualMachinedisk)

	return disk
}
