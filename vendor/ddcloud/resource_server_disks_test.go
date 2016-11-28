package ddcloud

import (
	"testing"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/assert"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// Unit test - splitInitiallyConfiguredDisksByType with no configured disks.
func TestSplitInitiallyConfiguredDisksEmpty(test *testing.T) {
	configuredDisks := []compute.VirtualMachineDisk{}
	actualDisks := []compute.VirtualMachineDisk{}

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(
		configuredDisks,
		actualDisks,
	)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 0, len(imageDisks))
	assert.EqualsInt("AdditionalDisks.Length", 0, len(additionalDisks))
}

// Unit test - splitInitiallyConfiguredDisksByType with both image and additional disks.
func TestSplitInitiallyConfiguredDisksBoth(test *testing.T) {
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

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(
		configuredDisks,
		actualDisks,
	)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 1, len(imageDisks))
	assert.EqualsInt("ImageDisks[0].SCSIUnitID", 0, imageDisks[0].SCSIUnitID)

	assert.EqualsInt("AdditionalDisks.Length", 1, len(additionalDisks))
	assert.EqualsInt("AdditionalDisks[0].SCSIUnitID", 1, additionalDisks[0].SCSIUnitID)
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

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(
		configuredDisks,
		actualDisks,
	)

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

	imageDisks, additionalDisks := splitInitiallyConfiguredDisksByType(
		configuredDisks,
		actualDisks,
	)

	assert := assert.ForTest(test)
	assert.EqualsInt("ImageDisks.Length", 0, len(imageDisks))
	assert.EqualsInt("AdditionalDisks.Length", 2, len(additionalDisks))
}

// Unit test - splitConfiguredDisksByAction with no configured disks.
func TestSplitConfiguredDisksByActionEmpty(test *testing.T) {
	configuredDisks := []compute.VirtualMachineDisk{}
	actualDisks := []compute.VirtualMachineDisk{}

	addDisks, changeDisks, removeDisks := splitConfiguredDisksByAction(
		configuredDisks,
		actualDisks,
	)

	assert := assert.ForTest(test)
	assert.EqualsInt("AddDisks.Length", 0, len(addDisks))
	assert.EqualsInt("ChangeDisks.Length", 0, len(changeDisks))
	assert.EqualsInt("RemoveDisks.Length", 0, len(removeDisks))
}

// Unit test - splitConfiguredDisksByAction with 2 new disks.
func TestSplitConfiguredDisksByActionNew2(test *testing.T) {
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

	addDisks, changeDisks, removeDisks := splitConfiguredDisksByAction(
		configuredDisks,
		actualDisks,
	)

	assert := assert.ForTest(test)
	assert.EqualsInt("AddDisks.Length", 2, len(addDisks))
	assert.EqualsInt("ChangeDisks.Length", 0, len(changeDisks))
	assert.EqualsInt("RemoveDisks.Length", 0, len(removeDisks))
}

// Unit test - splitConfiguredDisksByAction with 2 new disks.
func TestSplitConfiguredDisksByActionNew1Changed2(test *testing.T) {
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
		compute.VirtualMachineDisk{
			SCSIUnitID: 2,
			SizeGB:     20,
			Speed:      "STANDARD",
		},
		compute.VirtualMachineDisk{
			SCSIUnitID: 3,
			SizeGB:     10,
			Speed:      "STANDARD",
		},
	}
	actualDisks := []compute.VirtualMachineDisk{
		compute.VirtualMachineDisk{
			SCSIUnitID: 0,
			SizeGB:     5,
			Speed:      "STANDARD",
		},
		compute.VirtualMachineDisk{
			SCSIUnitID: 1,
			SizeGB:     50,
			Speed:      "STANDARD",
		},
		compute.VirtualMachineDisk{
			SCSIUnitID: 2,
			SizeGB:     20,
			Speed:      "HIGHPERFORMANCE",
		},
	}

	addDisks, changeDisks, removeDisks := splitConfiguredDisksByAction(
		configuredDisks,
		actualDisks,
	)

	assert := assert.ForTest(test)
	assert.EqualsInt("AddDisks.Length", 1, len(addDisks))
	assert.EqualsInt("AddDisks[0].SCSIUnitID", 3, addDisks[0].SCSIUnitID)

	assert.EqualsInt("ChangeDisks.Length", 2, len(changeDisks))
	assert.EqualsInt("ChangeDisks[0].SCSIUnitID", 1, changeDisks[0].SCSIUnitID)
	assert.EqualsInt("ChangeDisks[1].SCSIUnitID", 2, changeDisks[1].SCSIUnitID)

	assert.EqualsInt("RemoveDisks.Length", 0, len(removeDisks))
}
