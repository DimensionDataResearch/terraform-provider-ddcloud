package models

import "github.com/DimensionDataResearch/go-dd-cloud-compute/compute"

// TODO: Consider implementing SCSIControllers.CalculateActions([]compute.VirtualMachineSCSIController)

// SCSIControllers represents an array of SCSIController structures.
type SCSIControllers []SCSIController

// IsEmpty determines whether the SCSIController array is empty.
func (scsiControllers SCSIControllers) IsEmpty() bool {
	return len(scsiControllers) == 0
}

// ToVirtualMachineSCSIControllers converts the SCSIControllers to an array of compute.VirtualMachineSCSIController.
func (scsiControllers SCSIControllers) ToVirtualMachineSCSIControllers() []compute.VirtualMachineSCSIController {
	virtualMachineSCSIControllers := make([]compute.VirtualMachineSCSIController, len(scsiControllers))
	for index, scsiController := range scsiControllers {
		virtualMachineSCSIControllers[index] = scsiController.ToVirtualMachineSCSIController()
	}

	return virtualMachineSCSIControllers
}

// ToMaps converts the SCSIControllers to an array of maps.
func (scsiControllers SCSIControllers) ToMaps() []map[string]interface{} {
	scsiControllerPropertyList := make([]map[string]interface{}, len(scsiControllers))
	for index, scsiController := range scsiControllers {
		scsiControllerPropertyList[index] = scsiController.ToMap()
	}

	return scsiControllerPropertyList
}

// ByBusNumber creates a map of SCSIController keyed by SCSI bus number.
func (scsiControllers SCSIControllers) ByBusNumber() map[int]SCSIController {
	scsiControllersByBusNumber := make(map[int]SCSIController)
	for _, scsiController := range scsiControllers {
		scsiControllersByBusNumber[scsiController.BusNumber] = scsiController
	}

	return scsiControllersByBusNumber
}

// NewSCSIControllersFromStateData creates SCSIControllers from an array of Terraform state data.
//
// The values in the scsiControllerPropertyList are expected to be map[string]interface{}.
func NewSCSIControllersFromStateData(scsiControllerPropertyList []interface{}) SCSIControllers {
	scsiControllers := make(SCSIControllers, len(scsiControllerPropertyList))
	for index, data := range scsiControllerPropertyList {
		scsiControllerProperties := data.(map[string]interface{})
		scsiControllers[index] = NewSCSIControllerFromMap(scsiControllerProperties)
	}

	return scsiControllers
}

// NewSCSIControllersFromMaps creates SCSIControllers from an array of Terraform value maps.
func NewSCSIControllersFromMaps(scsiControllerPropertyList []map[string]interface{}) SCSIControllers {
	scsiControllers := make(SCSIControllers, len(scsiControllerPropertyList))
	for index, data := range scsiControllerPropertyList {
		scsiControllers[index] = NewSCSIControllerFromMap(data)
	}

	return scsiControllers
}

// NewSCSIControllersFromVirtualMachineSCSIControllers creates SCSIControllers from an array of compute.VirtualMachineSCSIController.
//
// Populates the SCSI bus number for each scsiController from its containing controller.
func NewSCSIControllersFromVirtualMachineSCSIControllers(virtualMachineSCSIControllers []compute.VirtualMachineSCSIController) SCSIControllers {
	var scsiControllers SCSIControllers

	for _, virtualMachineSCSIController := range virtualMachineSCSIControllers {
		scsiControllers = append(scsiControllers,
			NewSCSIControllerFromVirtualMachineSCSIController(virtualMachineSCSIController),
		)
	}

	return scsiControllers
}
