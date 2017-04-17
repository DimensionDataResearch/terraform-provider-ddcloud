package ddcloud

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

func newStringSet() *schema.Set {
	return &schema.Set{
		F: func(item interface{}) int {
			str := item.(string)

			return schema.HashString(str)
		},
	}
}

func intToPtr(value int) *int {
	return &value
}

func stringToPtr(value string) *string {
	return &value
}

func isEmpty(value string) bool {
	return len(value) == 0
}

func diskCount(scsiControllers []compute.VirtualMachineSCSIController) (count int) {
	for _, scsiController := range scsiControllers {
		count += len(scsiController.Disks)
	}

	return
}
