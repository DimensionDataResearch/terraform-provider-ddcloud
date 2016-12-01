package models

import "sort"

// Sort NetworkAdapters by Index.
type sortNetworkAdaptersByIndex struct {
	networkAdapters NetworkAdapters
}

var _ sort.Interface = sortNetworkAdaptersByIndex{}

// Len determines the number of NetworkAdapter structures in the NetworkAdapters.
func (sortNetworkAdapters sortNetworkAdaptersByIndex) Len() int {
	return len(sortNetworkAdapters.networkAdapters)
}

// Less determines whether the NetworkAdapter at firstIndex has an Index less than the Index of the NetworkAdapter the number of NetworkAdapter at secondIndex.
func (sortNetworkAdapters sortNetworkAdaptersByIndex) Less(firstIndex int, secondIndex int) bool {
	firstNetworkAdapter := sortNetworkAdapters.networkAdapters[firstIndex]
	secondNetworkAdapter := sortNetworkAdapters.networkAdapters[secondIndex]

	return firstNetworkAdapter.Index < secondNetworkAdapter.Index
}

// Swap the NetworkAdapter at firstIndex with the NetworkAdapter at secondIndex.
func (sortNetworkAdapters sortNetworkAdaptersByIndex) Swap(firstIndex int, secondIndex int) {
	firstNetworkAdapter := sortNetworkAdapters.networkAdapters[firstIndex]
	secondNetworkAdapter := sortNetworkAdapters.networkAdapters[secondIndex]

	sortNetworkAdapters.networkAdapters[firstIndex] = secondNetworkAdapter
	sortNetworkAdapters.networkAdapters[secondIndex] = firstNetworkAdapter
}
