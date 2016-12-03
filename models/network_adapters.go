package models

import (
	"log"

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
	return !networkAdapters.IsEmpty()
}

// HasAdditional determines whether the NetworkAdapter array includes one or more additional network adapters.
func (networkAdapters NetworkAdapters) HasAdditional() bool {
	return len(networkAdapters) > 1
}

// GetPrimary retrieves the primary network adapter (if present).
//
// The primary network adapter has Index 0.
func (networkAdapters NetworkAdapters) GetPrimary() *NetworkAdapter {
	if networkAdapters.HasPrimary() {
		return &networkAdapters[0]
	}

	return nil
}

// Insert a NetworkAdapter at the specified index.
//
// Returns a new NetworkAdapters.
func (networkAdapters NetworkAdapters) Insert(index int, networkAdapter NetworkAdapter) NetworkAdapters {
	firstSlice := networkAdapters[:index]
	secondSlice := append(
		NetworkAdapters{networkAdapter},
		networkAdapters[index:]...,
	)

	return append(firstSlice, secondSlice...)
}

// Remove the specified NetworkAdapter (by Id).
//
// Returns a new NetworkAdapters.
func (networkAdapters NetworkAdapters) Remove(networkAdapter NetworkAdapter) NetworkAdapters {
	if networkAdapter.ID == "" {
		return networkAdapters
	}

	adapterIndex := -1
	for index, adapter := range networkAdapters {
		if adapter.ID == networkAdapter.ID {
			adapterIndex = index

			break
		}
	}
	if adapterIndex == -1 {
		return networkAdapters
	}

	return networkAdapters.RemoveAt(adapterIndex)
}

// RemoveAt removes the NetworkAdapter at the specified index.
//
// Returns a new NetworkAdapters.
func (networkAdapters NetworkAdapters) RemoveAt(index int) NetworkAdapters {
	return append(
		networkAdapters[0:index],
		networkAdapters[index+1:]...,
	)
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

	for index, networkAdapter := range networkAdapters {
		if index == 0 {
			virtualMachineNetwork.PrimaryAdapter = networkAdapter.ToVirtualMachineNetworkAdapter()
		} else {
			virtualMachineNetwork.AdditionalNetworkAdapters = append(
				virtualMachineNetwork.AdditionalNetworkAdapters,
				networkAdapter.ToVirtualMachineNetworkAdapter(),
			)
		}
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

// ByID creates a map of NetworkAdapter keyed by Id.
func (networkAdapters NetworkAdapters) ByID() map[string]NetworkAdapter {
	networkAdaptersByID := make(map[string]NetworkAdapter)
	for _, networkAdapter := range networkAdapters {
		if networkAdapter.ID == "" {
			continue
		}

		networkAdaptersByID[networkAdapter.ID] = networkAdapter
	}

	return networkAdaptersByID
}

// ByMACAddress creates a map of NetworkAdapter keyed by MAC address.
func (networkAdapters NetworkAdapters) ByMACAddress() map[string]NetworkAdapter {
	networkAdaptersByMACAddress := make(map[string]NetworkAdapter)
	for _, networkAdapter := range networkAdapters {
		if networkAdapter.MACAddress == "" {
			continue
		}

		networkAdaptersByMACAddress[networkAdapter.MACAddress] = networkAdapter
	}

	return networkAdaptersByMACAddress
}

// Subtract the specified NetworkAdapters from the current NetworkAdapters.
func (networkAdapters NetworkAdapters) Subtract(otherNetworkAdapters NetworkAdapters) (remainingNetworkAdapters NetworkAdapters) {
	otherNetworkAdaptersByID := otherNetworkAdapters.ByID()
	for _, networkAdapter := range networkAdapters {
		if networkAdapter.ID == "" {
			continue
		}

		_, ok := otherNetworkAdaptersByID[networkAdapter.ID]
		if !ok {
			remainingNetworkAdapters = append(remainingNetworkAdapters, networkAdapter)
		}
	}

	return
}

// SplitByAction splits the (configured) network adapters by the action to be performed (add, change, or remove).
//
// configuredNetworkAdapters represents the network adapters currently specified in configuration.
// actualNetworkAdapters represents the network adapters in the server, as returned by CloudControl.
func (networkAdapters NetworkAdapters) SplitByAction(actualNetworkAdapters NetworkAdapters) (addNetworkAdapters NetworkAdapters, changeNetworkAdapters NetworkAdapters, removeNetworkAdapters NetworkAdapters) {
	actualNetworkAdaptersByID := actualNetworkAdapters.ByID()
	for _, configuredNetworkAdapter := range networkAdapters {
		actualNetworkAdapter, ok := actualNetworkAdaptersByID[configuredNetworkAdapter.ID]

		// We don't want to see this network adapter when we're looking for networkAdapters that don't appear in the configuration.
		delete(actualNetworkAdaptersByID, configuredNetworkAdapter.ID)

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

	// By process of elimination, any remaining actual network adapters do not appear in the configuration and should be removed.
	for _, unconfiguredNetworkAdapter := range actualNetworkAdaptersByID {
		removeNetworkAdapters = append(removeNetworkAdapters, unconfiguredNetworkAdapter)
	}

	return
}

// NewNetworkAdaptersFromVirtualMachineNetwork creates a new NetworkAdapters array from the specified compute.VirtualMachineNetwork
//
// This allocates index values in the order that adapters are found, and so it only works if there's *no* existing state at all.
func NewNetworkAdaptersFromVirtualMachineNetwork(virtualMachineNetwork compute.VirtualMachineNetwork) (networkAdapters NetworkAdapters) {
	primaryAdapter := NewNetworkAdapterFromVirtualMachineNetworkAdapter(virtualMachineNetwork.PrimaryAdapter)
	networkAdapters = append(networkAdapters, primaryAdapter)

	for _, additionalNetworkAdapter := range virtualMachineNetwork.AdditionalNetworkAdapters {
		additionalAdapter := NewNetworkAdapterFromVirtualMachineNetworkAdapter(additionalNetworkAdapter)

		networkAdapters = append(networkAdapters, additionalAdapter)
	}

	return
}
