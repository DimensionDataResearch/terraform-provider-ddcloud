package models

import (
	"fmt"
	"sort"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// ServerBackupClients represents an array of ServerBackupClient structures.
type ServerBackupClients []ServerBackupClient

// IsEmpty determines whether the ServerBackupClient array is empty.
func (clients ServerBackupClients) IsEmpty() bool {
	return len(clients) == 0
}

// SortByType sorts the backup clients by type.
func (clients ServerBackupClients) SortByType() {
	sorter := &backupClientSorter{
		ServerBackupClients: clients,
	}

	sort.Sort(sorter)
}

// ToBackupClientDetails converts the ServerBackupClients to an array of compute.BackupClientDetail.
func (clients ServerBackupClients) ToBackupClientDetails() []compute.BackupClientDetail {
	virtualMachineServerBackupClients := make([]compute.BackupClientDetail, len(clients))
	for index, client := range clients {
		virtualMachineServerBackupClients[index] = client.ToBackupClientDetail()
	}

	return virtualMachineServerBackupClients
}

// ToMaps converts the ServerBackupClients to an array of maps.
func (clients ServerBackupClients) ToMaps() []map[string]interface{} {
	clientPropertyList := make([]map[string]interface{}, len(clients))
	for index, client := range clients {
		clientPropertyList[index] = client.ToMap()
	}

	return clientPropertyList
}

// ByID creates a map of ServerBackupClient keyed by client Id.
func (clients ServerBackupClients) ByID() map[string]ServerBackupClient {
	clientsByID := make(map[string]ServerBackupClient)
	for _, client := range clients {
		clientsByID[client.ID] = client
	}

	return clientsByID
}

// ByType creates a map of ServerBackupClient keyed by client type.
func (clients ServerBackupClients) ByType() map[string]ServerBackupClient {
	clientsByType := make(map[string]ServerBackupClient)
	for _, client := range clients {
		clientsByType[client.Type] = client
	}

	return clientsByType
}

// CaptureIDs updates the ServerBackupClient Ids from the actual clients.
func (clients ServerBackupClients) CaptureIDs(actualServerBackupClients ServerBackupClients) {
	actualServerBackupClientsByType := actualServerBackupClients.ByType()
	for index := range clients {
		client := &clients[index]
		actualServerBackupClient, ok := actualServerBackupClientsByType[client.Type]
		if ok {
			client.ID = actualServerBackupClient.ID
		}
	}
}

// ApplyCurrentConfiguration applies the current configuration, inline, to the old configuration.
//
// Call this function on the old ServerBackupClients, passing the new ServerBackupClients.
func (clients *ServerBackupClients) ApplyCurrentConfiguration(currentServerBackupClients ServerBackupClients) {
	previousServerBackupClients := *clients
	currentServerBackupClientsByID := currentServerBackupClients.ByID()

	appliedServerBackupClients := make(ServerBackupClients, 0)
	for index := range previousServerBackupClients {
		previousServerBackupClient := &previousServerBackupClients[index]
		currentServerBackupClient, ok := currentServerBackupClientsByID[previousServerBackupClient.ID]
		if !ok {
			continue // ServerBackupClient no longer configured; leave it out.
		}

		// Update properties from current configuration.
		previousServerBackupClient.SchedulePolicyName = currentServerBackupClient.SchedulePolicyName
		previousServerBackupClient.StoragePolicyName = currentServerBackupClient.StoragePolicyName

		appliedServerBackupClients = append(appliedServerBackupClients, *previousServerBackupClient)
	}

	*clients = appliedServerBackupClients
}

// SplitByAction splits the (configured) server clients by the action to be performed (add, change, or remove).
//
// configuredServerBackupClients represents the clients currently specified in configuration.
// actualServerBackupClients represents the clients in the server, as returned by CloudControl.
func (clients ServerBackupClients) SplitByAction(actualServerBackupClients ServerBackupClients) (addServerBackupClients ServerBackupClients, changeServerBackupClients ServerBackupClients, removeServerBackupClients ServerBackupClients) {
	actualServerBackupClientsByType := actualServerBackupClients.ByType()
	for _, configuredServerBackupClient := range clients {
		actualServerBackupClient, ok := actualServerBackupClientsByType[configuredServerBackupClient.Type]

		// We don't want to see this client when we're looking for clients that don't appear in the configuration.
		delete(actualServerBackupClientsByType, configuredServerBackupClient.Type)

		if ok {
			// Existing client.
			var changed bool

			changed = changed || configuredServerBackupClient.SchedulePolicyName != actualServerBackupClient.SchedulePolicyName
			changed = changed || configuredServerBackupClient.StoragePolicyName != actualServerBackupClient.StoragePolicyName

			if configuredServerBackupClient.Alerting != nil && actualServerBackupClient.Alerting != nil {
				// Cheaty string comparison.
				configuredAlerting := fmt.Sprintf("%#v", *configuredServerBackupClient.Alerting)
				actualAlerting := fmt.Sprintf("%#v", *actualServerBackupClient.Alerting)

				changed = changed || configuredAlerting != actualAlerting
			} else {
				changed = changed || configuredServerBackupClient.Alerting != actualServerBackupClient.Alerting
			}

			if changed {
				changeServerBackupClients = append(changeServerBackupClients, configuredServerBackupClient)
			}
		} else {
			// New client.
			addServerBackupClients = append(addServerBackupClients, configuredServerBackupClient)
		}
	}

	// By process of elimination, any remaining actual clients do not appear in the configuration and should be removed.
	for unconfiguredServerBackupClientType := range actualServerBackupClientsByType {
		unconfiguredServerBackupClient := actualServerBackupClientsByType[unconfiguredServerBackupClientType]
		removeServerBackupClients = append(removeServerBackupClients, unconfiguredServerBackupClient)
	}

	return
}

// NewServerBackupClientsFromStateData creates ServerBackupClients from an array of Terraform state data.
//
// The values in the clientPropertyList are expected to be map[string]interface{}.
func NewServerBackupClientsFromStateData(clientPropertyList []interface{}) ServerBackupClients {
	clients := make(ServerBackupClients, len(clientPropertyList))
	for index, data := range clientPropertyList {
		clientProperties := data.(map[string]interface{})
		clients[index] = NewServerBackupClientFromMap(clientProperties)
	}

	return clients
}

// NewServerBackupClientsFromMaps creates ServerBackupClients from an array of Terraform value maps.
func NewServerBackupClientsFromMaps(clientPropertyList []map[string]interface{}) ServerBackupClients {
	clients := make(ServerBackupClients, len(clientPropertyList))
	for index, data := range clientPropertyList {
		clients[index] = NewServerBackupClientFromMap(data)
	}

	return clients
}

// NewServerBackupClientsFromBackupClientDetails creates ServerBackupClients from compute.BackupClientDetails.
func NewServerBackupClientsFromBackupClientDetails(backupClientDetails []compute.BackupClientDetail) (clients ServerBackupClients) {
	for _, backupClientDetail := range backupClientDetails {
		clients = append(clients,
			NewServerBackupClientFromBackupClientDetail(backupClientDetail),
		)
	}
	clients.SortByType()

	return
}

// backupClientSorter sorts backup clients by type
type backupClientSorter struct {
	ServerBackupClients ServerBackupClients
}

func (sorter backupClientSorter) Len() int {
	return len(sorter.ServerBackupClients)
}

func (sorter backupClientSorter) Less(index1 int, index2 int) bool {
	client1 := sorter.ServerBackupClients[index1]
	client2 := sorter.ServerBackupClients[index2]

	return client1.Type < client2.Type
}

func (sorter backupClientSorter) Swap(index1 int, index2 int) {
	temp := sorter.ServerBackupClients[index1]
	sorter.ServerBackupClients[index1] = sorter.ServerBackupClients[index2]
	sorter.ServerBackupClients[index2] = temp
}

var _ sort.Interface = &backupClientSorter{}
