package main

import (
	"compute-api/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

// resourcePropertyHelper provides commonly-used functionality for working with Terraform's schema.ResourceData.
type resourcePropertyHelper struct {
	data *schema.ResourceData
}

func propertyHelper(data *schema.ResourceData) resourcePropertyHelper {
	return resourcePropertyHelper{data}
}

func (helper resourcePropertyHelper) GetOptionalString(key string, allowEmpty bool) *string {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case string:
		if len(typedValue) > 0 || allowEmpty {
			return &typedValue
		}
	}

	return nil
}

func (helper resourcePropertyHelper) GetOptionalInt(key string, allowZero bool) *int {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case int:
		if typedValue != 0 || allowZero {
			return &typedValue
		}
	}

	return nil
}

func (helper resourcePropertyHelper) GetOptionalBool(key string) *bool {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case bool:
		return &typedValue
	default:
		return nil
	}
}

func (helper resourcePropertyHelper) GetVirtualMachineDisks(key string) []compute.VirtualMachineDisk {
	items := helper.data.Get(key).([]interface{})

	disks := make([]compute.VirtualMachineDisk, len(items))
	for index, item := range items {
		diskProperties := item.(map[string]interface{})
		disks[index] = parseDiskProperties(diskProperties)
	}

	return disks
}

// Parse a compute.VirtualMachineDisk from the specified disk property map.
func parseDiskProperties(diskProperties map[string]interface{}) compute.VirtualMachineDisk {
	disk := &compute.VirtualMachineDisk{}

	if unitID, ok := diskProperties[resourceKeyServerAdditionalDiskUnitID]; ok {
		disk.SCSIUnitID = unitID.(int)
	}

	if sizeGB, ok := diskProperties[resourceKeyServerAdditionalDiskSizeGB]; ok {
		disk.SizeGB = sizeGB.(int)
	}

	if speed, ok := diskProperties[resourceKeyServerAdditionalDiskSpeed]; ok {
		disk.Speed = speed.(string)
	}

	return *disk
}
