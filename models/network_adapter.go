package models

import (
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/maps"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// NetworkAdapter represents the Terraform configuration for a ddcloud_server network_adapter.
type NetworkAdapter struct {
	ID                 string
	Index              int
	VLANID             string
	PrivateIPv4Address string
	PrivateIPv6Address string
	AdapterType        string
}

// ReadMap populates the NetworkAdapter with values from the specified map.
func (networkAdapter *NetworkAdapter) ReadMap(networkAdapterProperties map[string]interface{}) {
	reader := maps.NewReader(networkAdapterProperties)

	networkAdapter.ID = reader.GetString("id")
	networkAdapter.Index = reader.GetInt("index")
	networkAdapter.VLANID = reader.GetString("vlan")
	networkAdapter.PrivateIPv4Address = reader.GetString("ipv4")
	networkAdapter.PrivateIPv6Address = reader.GetString("ipv6")
	networkAdapter.AdapterType = reader.GetString("type")
}

// ToMap creates a new map using the values from the NetworkAdapter.
func (networkAdapter *NetworkAdapter) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	networkAdapter.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the NetworkAdapter.
func (networkAdapter *NetworkAdapter) UpdateMap(networkAdapterProperties map[string]interface{}) {
	writer := maps.NewWriter(networkAdapterProperties)

	writer.SetString("id", networkAdapter.ID)
	writer.SetInt("index", networkAdapter.Index)
	writer.SetString("vlan", networkAdapter.VLANID)
	writer.SetString("ipv4", networkAdapter.PrivateIPv4Address)
	writer.SetString("ipv6", networkAdapter.PrivateIPv6Address)
	writer.SetString("type", networkAdapter.AdapterType)
}

// ReadVirtualMachineNetworkAdapter populates the NetworkAdapter with values from the specified VirtualMachineNetworkAdapter.
func (networkAdapter *NetworkAdapter) ReadVirtualMachineNetworkAdapter(virtualMachineNetworkAdapter compute.VirtualMachineNetworkAdapter) {
	networkAdapter.ID = ptrToString(virtualMachineNetworkAdapter.ID)
	networkAdapter.VLANID = ptrToString(virtualMachineNetworkAdapter.VLANID)
	networkAdapter.PrivateIPv4Address = ptrToString(virtualMachineNetworkAdapter.PrivateIPv4Address)
	networkAdapter.PrivateIPv6Address = ptrToString(virtualMachineNetworkAdapter.PrivateIPv6Address)
	networkAdapter.AdapterType = ptrToString(virtualMachineNetworkAdapter.AdapterType)
}

// ToVirtualMachineNetworkAdapter updates a map using values from the NetworkAdapter.
func (networkAdapter *NetworkAdapter) ToVirtualMachineNetworkAdapter() compute.VirtualMachineNetworkAdapter {
	virtualMachineNetworkAdapter := compute.VirtualMachineNetworkAdapter{}
	networkAdapter.UpdateVirtualMachineNetworkAdapter(&virtualMachineNetworkAdapter)

	return virtualMachineNetworkAdapter
}

// UpdateVirtualMachineNetworkAdapter updates a CloudControl VirtualMachineNetworkAdapter using values from the NetworkAdapter.
func (networkAdapter *NetworkAdapter) UpdateVirtualMachineNetworkAdapter(virtualMachineNetworkAdapter *compute.VirtualMachineNetworkAdapter) {
	virtualMachineNetworkAdapter.ID = stringToPtr(networkAdapter.ID)
	virtualMachineNetworkAdapter.VLANID = stringToPtr(networkAdapter.VLANID)
	virtualMachineNetworkAdapter.PrivateIPv4Address = stringToPtr(networkAdapter.PrivateIPv4Address)
	virtualMachineNetworkAdapter.PrivateIPv6Address = stringToPtr(networkAdapter.PrivateIPv6Address)
	virtualMachineNetworkAdapter.AdapterType = stringToPtr(networkAdapter.AdapterType)
}

// NewNetworkAdapterFromMap creates a NetworkAdapter from the values in the specified map.
func NewNetworkAdapterFromMap(networkAdapterProperties map[string]interface{}) NetworkAdapter {
	networkAdapter := NetworkAdapter{}
	networkAdapter.ReadMap(networkAdapterProperties)

	return networkAdapter
}

// NewNetworkAdapterFromVirtualMachineNetworkAdapter creates a NetworkAdapter from the values in the specified CloudControl VirtualMachineNetworkAdapter.
func NewNetworkAdapterFromVirtualMachineNetworkAdapter(virtualMachineNetworkAdapter compute.VirtualMachineNetworkAdapter) NetworkAdapter {
	networkAdapter := NetworkAdapter{}
	networkAdapter.ReadVirtualMachineNetworkAdapter(virtualMachineNetworkAdapter)

	return networkAdapter
}

// NewNetworkAdaptersFromStateData creates NetworkAdapters from an array of Terraform state data.
//
// The values in the networkAdapterPropertyList are expected to be map[string]interface{}.
func NewNetworkAdaptersFromStateData(networkAdapterPropertyList []interface{}) NetworkAdapters {
	networkAdapters := make(NetworkAdapters, len(networkAdapterPropertyList))
	for index, data := range networkAdapterPropertyList {
		networkAdapterProperties := data.(map[string]interface{})
		networkAdapters[index] = NewNetworkAdapterFromMap(networkAdapterProperties)
	}

	return networkAdapters
}

// NewNetworkAdaptersFromMaps creates NetworkAdapters from an array of Terraform value maps.
func NewNetworkAdaptersFromMaps(networkAdapterPropertyList []map[string]interface{}) NetworkAdapters {
	networkAdapters := make(NetworkAdapters, len(networkAdapterPropertyList))
	for index, data := range networkAdapterPropertyList {
		networkAdapters[index] = NewNetworkAdapterFromMap(data)
	}

	return networkAdapters
}

// NewNetworkAdaptersFromVirtualMachineNetworkAdapters creates NetworkAdapters from an array of compute.VirtualMachineNetworkAdapter.
func NewNetworkAdaptersFromVirtualMachineNetworkAdapters(virtualMachineNetworkAdapters []compute.VirtualMachineNetworkAdapter) NetworkAdapters {
	networkAdapters := make(NetworkAdapters, len(virtualMachineNetworkAdapters))
	for index, virtualMachineNetworkAdapter := range virtualMachineNetworkAdapters {
		networkAdapters[index] = NewNetworkAdapterFromVirtualMachineNetworkAdapter(virtualMachineNetworkAdapter)
	}

	return networkAdapters
}
