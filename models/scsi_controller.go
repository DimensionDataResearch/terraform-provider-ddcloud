package models

import (
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/maps"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// SCSIController represents the Terraform configuration for a ddcloud_server scsi_controller.
type SCSIController struct {
	ID          string
	BusNumber   int
	AdapterType string
}

// ReadMap populates the SCSIController with values from the specified map.
func (scsiController *SCSIController) ReadMap(scsiControllerProperties map[string]interface{}) {
	reader := maps.NewReader(scsiControllerProperties)

	scsiController.ID = reader.GetString("id")
	scsiController.BusNumber = reader.GetInt("bus_number")
	scsiController.AdapterType = reader.GetString("adapter_type")
}

// ToMap creates a new map using the values from the SCSIController.
func (scsiController *SCSIController) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	scsiController.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the SCSIController.
func (scsiController *SCSIController) UpdateMap(scsiControllerProperties map[string]interface{}) {
	writer := maps.NewWriter(scsiControllerProperties)

	writer.SetString("id", scsiController.ID)
	writer.SetInt("bus_number", scsiController.BusNumber)
	writer.SetString("adapter_type", scsiController.AdapterType)
}

// ReadVirtualMachineSCSIController populates the SCSIController with values from the specified VirtualMachineSCSIController.
func (scsiController *SCSIController) ReadVirtualMachineSCSIController(virtualMachineSCSIController compute.VirtualMachineSCSIController) {
	scsiController.ID = virtualMachineSCSIController.ID
	scsiController.BusNumber = virtualMachineSCSIController.BusNumber
	scsiController.AdapterType = virtualMachineSCSIController.AdapterType
}

// ToVirtualMachineSCSIController updates a map using values from the SCSIController.
func (scsiController *SCSIController) ToVirtualMachineSCSIController() compute.VirtualMachineSCSIController {
	virtualMachineSCSIController := compute.VirtualMachineSCSIController{}
	scsiController.UpdateVirtualMachineSCSIController(&virtualMachineSCSIController)

	return virtualMachineSCSIController
}

// UpdateVirtualMachineSCSIController updates a CloudControl VirtualMachineSCSIController using values from the SCSIController.
func (scsiController *SCSIController) UpdateVirtualMachineSCSIController(virtualMachineSCSIController *compute.VirtualMachineSCSIController) {
	virtualMachineSCSIController.ID = scsiController.ID
	virtualMachineSCSIController.BusNumber = scsiController.BusNumber
	virtualMachineSCSIController.AdapterType = scsiController.AdapterType
}

// NewSCSIControllerFromMap creates a SCSIController from the values in the specified map.
func NewSCSIControllerFromMap(scsiControllerProperties map[string]interface{}) SCSIController {
	scsiController := SCSIController{}
	scsiController.ReadMap(scsiControllerProperties)

	return scsiController
}

// NewSCSIControllerFromVirtualMachineSCSIController creates a SCSIController from the values in the specified CloudControl VirtualMachineSCSIController.
func NewSCSIControllerFromVirtualMachineSCSIController(virtualMachineSCSIController compute.VirtualMachineSCSIController) SCSIController {
	scsiController := SCSIController{}
	scsiController.ReadVirtualMachineSCSIController(virtualMachineSCSIController)

	return scsiController
}
