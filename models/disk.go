package models

import (
	"fmt"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/maps"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// SCSIPath builds a path representing the specified SCSI bus number and logical unit ID.
func SCSIPath(busNumber int, unitID int) string {
	return fmt.Sprintf("%d/%d", busNumber, unitID)
}

// Disk represents the Terraform configuration for a ddcloud_server disk.
type Disk struct {
	ID            string
	SCSIBusNumber int
	SCSIUnitID    int
	SizeGB        int
	Speed         string
}

// SCSIPath builds a path representing the disk's SCSI bus number and logical unit ID.
func (disk Disk) SCSIPath() string {
	return fmt.Sprintf("%d/%d", disk.SCSIBusNumber, disk.SCSIUnitID)
}

// ReadMap populates the Disk with values from the specified map.
func (disk *Disk) ReadMap(diskProperties map[string]interface{}) {
	reader := maps.NewReader(diskProperties)

	disk.ID = reader.GetString("id")
	disk.SCSIBusNumber = reader.GetInt("scsi_bus_number")
	disk.SCSIUnitID = reader.GetInt("scsi_unit_id")
	disk.SizeGB = reader.GetInt("size_gb")
	disk.Speed = reader.GetString("speed")
}

// ToMap creates a new map using the values from the Disk.
func (disk *Disk) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	disk.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the Disk.
func (disk *Disk) UpdateMap(diskProperties map[string]interface{}) {
	writer := maps.NewWriter(diskProperties)

	writer.SetString("id", disk.ID)
	writer.SetInt("scsi_bus_number", disk.SCSIBusNumber)
	writer.SetInt("scsi_unit_id", disk.SCSIUnitID)
	writer.SetInt("size_gb", disk.SizeGB)
	writer.SetString("speed", disk.Speed)
}

// ReadVirtualMachineDisk populates the Disk with values from the specified VirtualMachineDisk.
func (disk *Disk) ReadVirtualMachineDisk(virtualMachineDisk compute.VirtualMachineDisk) {
	disk.ID = virtualMachineDisk.ID
	disk.SCSIUnitID = virtualMachineDisk.SCSIUnitID
	disk.SizeGB = virtualMachineDisk.SizeGB
	disk.Speed = virtualMachineDisk.Speed
}

// ToVirtualMachineDisk updates a map using values from the Disk.
func (disk *Disk) ToVirtualMachineDisk() compute.VirtualMachineDisk {
	virtualMachineDisk := compute.VirtualMachineDisk{}
	disk.UpdateVirtualMachineDisk(&virtualMachineDisk)

	return virtualMachineDisk
}

// UpdateVirtualMachineDisk updates a CloudControl VirtualMachineDisk using values from the Disk.
func (disk *Disk) UpdateVirtualMachineDisk(virtualMachineDisk *compute.VirtualMachineDisk) {
	virtualMachineDisk.ID = disk.ID
	virtualMachineDisk.SCSIUnitID = disk.SCSIUnitID
	virtualMachineDisk.SizeGB = disk.SizeGB
	virtualMachineDisk.Speed = disk.Speed
}

// NewDiskFromMap creates a Disk from the values in the specified map.
func NewDiskFromMap(diskProperties map[string]interface{}) Disk {
	disk := Disk{}
	disk.ReadMap(diskProperties)

	return disk
}

// NewDiskFromVirtualMachineDisk creates a Disk from the values in the specified CloudControl VirtualMachineDisk.
func NewDiskFromVirtualMachineDisk(virtualMachineDisk compute.VirtualMachineDisk, scsiBusNumber int) Disk {
	disk := Disk{
		SCSIBusNumber: scsiBusNumber,
	}
	disk.ReadVirtualMachineDisk(virtualMachineDisk)

	return disk
}
