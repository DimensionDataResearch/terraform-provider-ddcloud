package models

import "github.com/DimensionDataResearch/go-dd-cloud-compute/compute"

// NetworkAdapters represents an array of NetworkAdapter structures.
type NetworkAdapters []NetworkAdapter

// IsEmpty determines whether the NetworkAdapter array is empty.
func (networkAdapters NetworkAdapters) IsEmpty() bool {
	return len(networkAdapters) == 0
}

// ToVirtualMachineNetworkAdapters converts the NetworkAdapters to an array of compute.VirtualMachineNetworkAdapter.
func (networkAdapters NetworkAdapters) ToVirtualMachineNetworkAdapters() []compute.VirtualMachineNetworkAdapter {
	virtualMachineNetworkAdapters := make([]compute.VirtualMachineNetworkAdapter, len(networkAdapters))
	for index, networkAdapter := range networkAdapters {
		virtualMachineNetworkAdapters[index] = networkAdapter.ToVirtualMachineNetworkAdapter()
	}

	return virtualMachineNetworkAdapters
}

// ToMaps converts the NetworkAdapters to an array of maps.
func (networkAdapters NetworkAdapters) ToMaps() []map[string]interface{} {
	networkAdapterPropertyList := make([]map[string]interface{}, len(networkAdapters))
	for index, networkAdapter := range networkAdapters {
		networkAdapterPropertyList[index] = networkAdapter.ToMap()
	}

	return networkAdapterPropertyList
}

// ByIndex creates a map of NetworkAdapter keyed by index.
func (networkAdapters NetworkAdapters) ByIndex() map[int]NetworkAdapter {
	networkAdaptersByIndex := make(map[int]NetworkAdapter)
	for _, networkAdapter := range networkAdapters {
		networkAdaptersByIndex[networkAdapter.Index] = networkAdapter
	}

	return networkAdaptersByIndex
}
