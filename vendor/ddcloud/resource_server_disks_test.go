package ddcloud

import (
	"fmt"
	"testing"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/assert"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// Unit test - splitInitiallyConfiguredDisksByType with both image and additional disks.
func TestSplitInitiallyConfiguredDisks(test *testing.T) {
	configuredDisks := []compute.VirtualMachineDisk{
		compute.VirtualMachineDisk{
			ID:         nil,
			SCSIUnitID: 0,
			SizeGB:     5,
			Speed:      "STANDARD",
		},
		compute.VirtualMachineDisk{
			ID:         nil,
			SCSIUnitID: 1,
			SizeGB:     20,
			Speed:      "STANDARD",
		},
	}
	actualDisks := []compute.VirtualMachineDisk{
		configuredDisks[0],
	}

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(configuredDisks, actualDisks)

	fmt.Printf("ImageDisks = %#v\n", imageDisks)
	fmt.Printf("AdditionalDisks = %#v\n", additionalDisks)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 1, len(imageDisks))
	assert.EqualsInt("ImageDisks[0].SCSIUnitID", 0, imageDisks[0].SCSIUnitID)

	assert.EqualsInt("AdditionalDisks.Length", 1, len(additionalDisks))
	assert.EqualsInt("AdditionalDisks[0].SCSIUnitID", 1, additionalDisks[0].SCSIUnitID)
}

// Unit test - splitInitiallyConfiguredDisksByType with no configured disks.
func TestSplitInitiallyConfiguredDisksEmpty(test *testing.T) {
	configuredDisks := []compute.VirtualMachineDisk{}
	actualDisks := []compute.VirtualMachineDisk{}

	imageDisks, actualDisks := splitInitiallyConfiguredDisksByType(configuredDisks, actualDisks)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 0, len(imageDisks))
	assert.EqualsInt("AdditionalDisks.Length", 0, len(imageDisks))
}

// Unit test - splitInitiallyConfiguredDisksByType with only image disks.
func TestSplitInitiallyConfiguredDisksOnlyImageDisks(test *testing.T) {
	configuredDisks := []compute.VirtualMachineDisk{
		compute.VirtualMachineDisk{
			SCSIUnitID: 0,
			SizeGB:     5,
			Speed:      "STANDARD",
		},
		compute.VirtualMachineDisk{
			SCSIUnitID: 1,
			SizeGB:     20,
			Speed:      "STANDARD",
		},
	}

	// actualDisks = configuredDisks
	actualDisks := configuredDisks[:]

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(configuredDisks, actualDisks)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 2, len(imageDisks))
	assert.EqualsInt("AdditionalDisks.Length", 0, len(additionalDisks))
}

// Unit test - splitInitiallyConfiguredDisksByType with only additional disks.
func TestSplitInitiallyConfiguredDisksOnlyAdditionalDisks(test *testing.T) {
	configuredDisks := []compute.VirtualMachineDisk{
		compute.VirtualMachineDisk{
			SCSIUnitID: 0,
			SizeGB:     5,
			Speed:      "STANDARD",
		},
		compute.VirtualMachineDisk{
			SCSIUnitID: 1,
			SizeGB:     20,
			Speed:      "STANDARD",
		},
	}
	actualDisks := []compute.VirtualMachineDisk{}

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(configuredDisks, actualDisks)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 0, len(imageDisks))
	assert.EqualsInt("AdditionalDisks.Length", 2, len(additionalDisks))
}
