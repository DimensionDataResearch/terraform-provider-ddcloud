package models

import (
	"log"
	"sort"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// NetworkAdapters represents an array of NetworkAdapter structures.
type NetworkAdapters []NetworkAdapter

// IsEmpty determines whether the NetworkAdapter array is empty.
func (networkAdapters NetworkAdapters) IsEmpty() bool {
	return len(networkAdapters) == 0
}

// HasPrimary determines whether the NetworkAdapter array includes a primary network adapter.
func (networkAdapters NetworkAdapters) HasPrimary() bool {
	for _, networkAdapter := range networkAdapters {
		if networkAdapter.Index == 0 {
			return true
		}
	}

	return false
}

// HasAdditional determines whether the NetworkAdapter array includes one or more additional network adapters.
func (networkAdapters NetworkAdapters) HasAdditional() bool {
	for _, networkAdapter := range networkAdapters {
		if networkAdapter.Index != 0 {
			return true
		}
	}

	return false
}

// HasContiguousIndexes determines whether each NetworkAdapter in the NetworkAdapters has an index that is 1 greater than the previous NetworkAdapter.
func (networkAdapters NetworkAdapters) HasContiguousIndexes() bool {
	if len(networkAdapters) < 2 {
		return true
	}

	sortedAdapters := networkAdapters[:] // Copy
	sortedAdapters.SortByIndex()

	index := sortedAdapters[0].Index
	for _, sortedAdapter := range sortedAdapters {
		if sortedAdapter.Index != index {
			return false
		}

		index++
	}

	return true
}

// GetPrimary retrieves the primary network adapter (if present).
//
// The primary network adapter has Index 0.
func (networkAdapters NetworkAdapters) GetPrimary() *NetworkAdapter {
	for index := range networkAdapters {
		if networkAdapters[index].Index == 0 {
			return &networkAdapters[index]
		}
	}

	return nil
}

// SortByIndex sorts the NetworkAdapters by NetworkAdapter.Index.
func (networkAdapters NetworkAdapters) SortByIndex() {
	sorter := &sortNetworkAdaptersByIndex{
		networkAdapters: networkAdapters,
	}
	sort.Sort(sorter)
}

// ToVirtualMachineNetworkAdapters converts the NetworkAdapters to an array of compute.VirtualMachineNetworkAdapter.
func (networkAdapters NetworkAdapters) ToVirtualMachineNetworkAdapters() []compute.VirtualMachineNetworkAdapter {
	virtualMachineNetworkAdapters := make([]compute.VirtualMachineNetworkAdapter, len(networkAdapters))
	for index, networkAdapter := range networkAdapters {
		virtualMachineNetworkAdapters[index] = networkAdapter.ToVirtualMachineNetworkAdapter()
	}

	return virtualMachineNetworkAdapters
}

// ToVirtualMachineNetwork converts the NetworkAdapters to a compute.VirtualMachineNetwork.
func (networkAdapters NetworkAdapters) ToVirtualMachineNetwork() compute.VirtualMachineNetwork {
	virtualMachineNetwork := compute.VirtualMachineNetwork{}
	networkAdapters.UpdateVirtualMachineNetwork(&virtualMachineNetwork)

	return virtualMachineNetwork
}

// UpdateVirtualMachineNetwork updates a compute.VirtualMachineNetwork with values from the NetworkAdapters.
func (networkAdapters NetworkAdapters) UpdateVirtualMachineNetwork(virtualMachineNetwork *compute.VirtualMachineNetwork) {
	if networkAdapters.IsEmpty() {
		return
	}

	for _, networkAdapter := range networkAdapters {
		if networkAdapter.Index == 0 {
			virtualMachineNetwork.PrimaryAdapter = networkAdapter.ToVirtualMachineNetworkAdapter()
		} else {
			virtualMachineNetwork.AdditionalNetworkAdapters = append(
				virtualMachineNetwork.AdditionalNetworkAdapters,
				networkAdapter.ToVirtualMachineNetworkAdapter(),
			)
		}
	}
}

// InitializeIndexes sets the Index field on each NetworkAdapter to its index in the array.
func (networkAdapters NetworkAdapters) InitializeIndexes() {
	for index := range networkAdapters {
		networkAdapter := &networkAdapters[index]
		networkAdapter.Index = index
	}
}

// CaptureIDs sets the ID field on each NetworkAdapter based on the ID of the compute.VirtualMachineNetworkAdapter with the same index.
func (networkAdapters NetworkAdapters) CaptureIDs(virtualMachineNetwork compute.VirtualMachineNetwork) {
	if networkAdapters.IsEmpty() {
		return
	}

	actualNetworkAdapter := virtualMachineNetwork.PrimaryAdapter
	networkAdapter := &networkAdapters[0]
	networkAdapter.ID = *actualNetworkAdapter.ID

	for index, actualNetworkAdapter := range virtualMachineNetwork.AdditionalNetworkAdapters {
		networkAdapter = &networkAdapters[index+1]
		networkAdapter.ID = *actualNetworkAdapter.ID
	}
}

// CaptureIndexes updates the Index field on each NetworkAdapter with its corresponding index in the compute.VirtualMachineNetwork.
func (networkAdapters NetworkAdapters) CaptureIndexes(virtualMachineNetwork compute.VirtualMachineNetwork) {
	if networkAdapters.IsEmpty() {
		return
	}

	actualAdapterIndexesByID := make(map[string]int)
	actualAdapter := virtualMachineNetwork.PrimaryAdapter
	actualAdapterIndexesByID[*actualAdapter.ID] = 0
	for index, actualAdapter := range virtualMachineNetwork.AdditionalNetworkAdapters {
		actualAdapter = virtualMachineNetwork.AdditionalNetworkAdapters[index]
		actualAdapterIndexesByID[*actualAdapter.ID] = index + 1
	}

	for index := range networkAdapters {
		networkAdapter := &networkAdapters[index]
		if networkAdapter.ID == "" {
			continue
		}

		actualAdapterIndex, ok := actualAdapterIndexesByID[networkAdapter.ID]
		if ok {
			networkAdapter.Index = actualAdapterIndex
		}
	}
}

// ReadVirtualMachineNetwork updates each NetworkAdapter with values from the corresponding compute.VirtualMachineNetworkAdapter (if one is found with the same Id).
func (networkAdapters NetworkAdapters) ReadVirtualMachineNetwork(virtualMachineNetwork compute.VirtualMachineNetwork) {
	actualNetworkAdaptersByID := NewNetworkAdaptersFromVirtualMachineNetwork(virtualMachineNetwork).ByID()

	for index := range networkAdapters {
		networkAdapter := &networkAdapters[index]
		if networkAdapter.ID == "" {
			continue
		}

		actualNetworkAdapter, ok := actualNetworkAdaptersByID[networkAdapter.ID]
		if !ok {
			log.Printf("No configuration found for primary network adapter '%s'", networkAdapter.ID)

			continue
		}

		networkAdapter.ReadNetworkAdapter(actualNetworkAdapter)
	}
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

// ByID creates a map of NetworkAdapter keyed by Id.
func (networkAdapters NetworkAdapters) ByID() map[string]NetworkAdapter {
	networkAdaptersByIndex := make(map[string]NetworkAdapter)
	for _, networkAdapter := range networkAdapters {
		if networkAdapter.ID == "" {
			continue
		}

		networkAdaptersByIndex[networkAdapter.ID] = networkAdapter
	}

	return networkAdaptersByIndex
}

// SplitByAction splits the (configured) network adapters by the action to be performed (add, change, or remove).
//
// configuredNetworkAdapters represents the network adapters currently specified in configuration.
// actualNetworkAdapters represents the network adapters in the server, as returned by CloudControl.
func (networkAdapters NetworkAdapters) SplitByAction(actualNetworkAdapters NetworkAdapters) (addNetworkAdapters NetworkAdapters, changeNetworkAdapters NetworkAdapters, removeNetworkAdapters NetworkAdapters) {
	actualNetworkAdaptersByIndex := actualNetworkAdapters.ByIndex()
	for _, configuredNetworkAdapter := range networkAdapters {
		actualNetworkAdapter, ok := actualNetworkAdaptersByIndex[configuredNetworkAdapter.Index]

		// We don't want to see this networkAdapter when we're looking for networkAdapters that don't appear in the configuration.
		delete(actualNetworkAdaptersByIndex, configuredNetworkAdapter.Index)

		if ok {
			// Existing network adapter.
			if configuredNetworkAdapter.PrivateIPv4Address != actualNetworkAdapter.PrivateIPv4Address {
				changeNetworkAdapters = append(changeNetworkAdapters, configuredNetworkAdapter)
			}
		} else {
			// New networkAdapter.
			addNetworkAdapters = append(addNetworkAdapters, configuredNetworkAdapter)
		}
	}

	// By process of elimination, any remaining actual networkAdapters do not appear in the configuration and should be removed.
	for unconfiguredNetworkAdapterIndex := range actualNetworkAdaptersByIndex {
		unconfiguredNetworkAdapter := actualNetworkAdaptersByIndex[unconfiguredNetworkAdapterIndex]
		removeNetworkAdapters = append(removeNetworkAdapters, unconfiguredNetworkAdapter)
	}

	return
}

// NewNetworkAdaptersFromVirtualMachineNetwork creates a new NetworkAdapters array from the specified compute.VirtualMachineNetwork
//
// This allocates index values in the order that adapters are found, and so it only works if there's *no* existing state at all.
func NewNetworkAdaptersFromVirtualMachineNetwork(virtualMachineNetwork compute.VirtualMachineNetwork) (networkAdapters NetworkAdapters) {
	primaryAdapter := NewNetworkAdapterFromVirtualMachineNetworkAdapter(virtualMachineNetwork.PrimaryAdapter)
	primaryAdapter.Index = 0
	networkAdapters = append(networkAdapters, primaryAdapter)

	for index, additionalNetworkAdapter := range virtualMachineNetwork.AdditionalNetworkAdapters {
		additionalAdapter := NewNetworkAdapterFromVirtualMachineNetworkAdapter(additionalNetworkAdapter)
		additionalAdapter.Index = index

		networkAdapters = append(networkAdapters, additionalAdapter)
	}

	return
}
