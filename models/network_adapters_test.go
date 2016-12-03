package models

import (
	"testing"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/assert"
)

// Unit test - given 0 network adapters, subtract 3 network adapters.
func TestSubtractNetworkAdapters_0_3(test *testing.T) {
	previousConfig := NetworkAdapters{}
	currentConfig := NetworkAdapters{
		NetworkAdapter{
			ID:                 "7b8fb12e-9ce6-440e-8a0f-2a139f878967",
			MACAddress:         "00:50:56:a3:79:5e",
			VLANID:             "686bca8d-3cfa-461a-b4ad-88fd77219947",
			PrivateIPv4Address: "192.168.17.20",
			PrivateIPv6Address: "2607:f480:111:1822:7479:fc1b:2e90:b617",
			AdapterType:        "VMXNET3",
		},
		NetworkAdapter{
			ID:                 "aad233e6-8229-4a47-be42-cc0b449eb03f",
			MACAddress:         "00:50:56:a3:5c:79",
			VLANID:             "6f84dce4-1ec6-4992-bf02-df15d4d3dd37",
			PrivateIPv4Address: "192.168.18.20",
			PrivateIPv6Address: "2607:f480:111:1820:2aad:ee0d:f634:8f4a",
			AdapterType:        "E1000",
		},
		NetworkAdapter{
			ID:                 "83fe7621-278c-4f13-82a6-6848a623cd7f",
			MACAddress:         "00:50:56:a3:68:f2",
			VLANID:             "40bb9975-63c6-43fa-96ab-2392df45f923",
			PrivateIPv4Address: "192.168.19.20",
			PrivateIPv6Address: "2607:f480:111:1821:317d:4cf3:5605:465e",
			AdapterType:        "E1000",
		},
	}

	removedAdapters := previousConfig.Subtract(currentConfig)

	assert := assert.ForTest(test)
	assert.EqualsInt("RemovedAdapters.Length", 0, len(removedAdapters))
}

// Unit test - given 3 network adapters, subtract the first and last network adapters.
func TestSubtractNetworkAdapters_3_2_Same(test *testing.T) {
	previousConfig := NetworkAdapters{
		NetworkAdapter{
			ID:                 "7b8fb12e-9ce6-440e-8a0f-2a139f878967",
			MACAddress:         "00:50:56:a3:79:5e",
			VLANID:             "686bca8d-3cfa-461a-b4ad-88fd77219947",
			PrivateIPv4Address: "192.168.17.20",
			PrivateIPv6Address: "2607:f480:111:1822:7479:fc1b:2e90:b617",
			AdapterType:        "VMXNET3",
		},
		NetworkAdapter{
			ID:                 "aad233e6-8229-4a47-be42-cc0b449eb03f",
			MACAddress:         "00:50:56:a3:5c:79",
			VLANID:             "6f84dce4-1ec6-4992-bf02-df15d4d3dd37",
			PrivateIPv4Address: "192.168.18.20",
			PrivateIPv6Address: "2607:f480:111:1820:2aad:ee0d:f634:8f4a",
			AdapterType:        "E1000",
		},
		NetworkAdapter{
			ID:                 "83fe7621-278c-4f13-82a6-6848a623cd7f",
			MACAddress:         "00:50:56:a3:68:f2",
			VLANID:             "40bb9975-63c6-43fa-96ab-2392df45f923",
			PrivateIPv4Address: "192.168.19.20",
			PrivateIPv6Address: "2607:f480:111:1821:317d:4cf3:5605:465e",
			AdapterType:        "E1000",
		},
	}

	currentConfig := NetworkAdapters{
		previousConfig[0],
		previousConfig[2],
	}

	removedAdapters := previousConfig.Subtract(currentConfig)

	assert := assert.ForTest(test)
	assert.EqualsInt("RemovedAdapters.Length", 1, len(removedAdapters))
	assert.EqualsString("RemovedAdapters[0].ID", "aad233e6-8229-4a47-be42-cc0b449eb03f", removedAdapters[0].ID)
}

// Unit test - given 3 network adapters, subtract same 3 network adapters.
func TestSubtractNetworkAdapters_3_3_Same(test *testing.T) {
	previousConfig := NetworkAdapters{
		NetworkAdapter{
			ID:                 "7b8fb12e-9ce6-440e-8a0f-2a139f878967",
			MACAddress:         "00:50:56:a3:79:5e",
			VLANID:             "686bca8d-3cfa-461a-b4ad-88fd77219947",
			PrivateIPv4Address: "192.168.17.20",
			PrivateIPv6Address: "2607:f480:111:1822:7479:fc1b:2e90:b617",
			AdapterType:        "VMXNET3",
		},
		NetworkAdapter{
			ID:                 "aad233e6-8229-4a47-be42-cc0b449eb03f",
			MACAddress:         "00:50:56:a3:5c:79",
			VLANID:             "6f84dce4-1ec6-4992-bf02-df15d4d3dd37",
			PrivateIPv4Address: "192.168.18.20",
			PrivateIPv6Address: "2607:f480:111:1820:2aad:ee0d:f634:8f4a",
			AdapterType:        "E1000",
		},
		NetworkAdapter{
			ID:                 "83fe7621-278c-4f13-82a6-6848a623cd7f",
			MACAddress:         "00:50:56:a3:68:f2",
			VLANID:             "40bb9975-63c6-43fa-96ab-2392df45f923",
			PrivateIPv4Address: "192.168.19.20",
			PrivateIPv6Address: "2607:f480:111:1821:317d:4cf3:5605:465e",
			AdapterType:        "E1000",
		},
	}

	currentConfig := NetworkAdapters{
		previousConfig[0],
		previousConfig[2],
	}

	removedAdapters := previousConfig.Subtract(currentConfig)

	assert := assert.ForTest(test)
	assert.EqualsInt("RemovedAdapters.Length", 1, len(removedAdapters))
	assert.EqualsString("RemovedAdapters[0].ID", "aad233e6-8229-4a47-be42-cc0b449eb03f", removedAdapters[0].ID)
}
