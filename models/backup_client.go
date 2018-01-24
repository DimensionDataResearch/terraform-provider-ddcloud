package models

import (
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/maps"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// ServerBackupClient represents a backup client assigned to a server.
type ServerBackupClient struct {
	ID                 string
	Type               string
	Description        string
	StoragePolicyName  string
	SchedulePolicyName string
	DownloadURL        string
	Status             string
	Alerting           *BackupClientAlerting
}

// BackupClientAlerting represents the alerting configuration (if any) for a backup client.
type BackupClientAlerting struct {
	Trigger string
	Emails  []string
}

// ReadMap populates the ServerBackupClient with values from the specified map.
func (backupClient *ServerBackupClient) ReadMap(backupClientProperties map[string]interface{}) {
	reader := maps.NewReader(backupClientProperties)

	backupClient.ID = reader.GetString("id")
	backupClient.Type = reader.GetString("type")
	backupClient.Description = reader.GetString("description")
	backupClient.StoragePolicyName = reader.GetString("storage_policy")
	backupClient.SchedulePolicyName = reader.GetString("schedule_policy")
	backupClient.DownloadURL = reader.GetString("download_url")

	rawValue, ok := backupClientProperties["alert"]
	if !ok {
		return
	}

	rawAlertingProperties, ok := rawValue.([]interface{})
	if !ok || len(rawAlertingProperties) == 0 {
		return
	}

	alertingProperties, ok := rawAlertingProperties[0].(map[string]interface{})
	if !ok {
		return
	}

	alertingReader := maps.NewReader(alertingProperties)
	backupClient.Alerting = &BackupClientAlerting{
		Trigger: alertingReader.GetString("trigger"),
		Emails:  alertingReader.GetStringSlice("emails"),
	}
}

// ToMap creates a new map using the values from the ServerBackupClient.
func (backupClient *ServerBackupClient) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	backupClient.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the ServerBackupClient.
func (backupClient *ServerBackupClient) UpdateMap(backupClientProperties map[string]interface{}) {
	writer := maps.NewWriter(backupClientProperties)

	writer.SetString("id", backupClient.ID)
	writer.SetString("type", backupClient.Type)
	writer.SetString("description", backupClient.Description)
	writer.SetString("storage_policy", backupClient.StoragePolicyName)
	writer.SetString("schedule_policy", backupClient.SchedulePolicyName)
	writer.SetString("download_url", backupClient.DownloadURL)
	writer.SetString("status", backupClient.Status)

	if backupClient.Alerting != nil {
		alertingProperties := make(map[string]interface{})
		alertingWriter := maps.NewWriter(alertingProperties)

		alertingWriter.SetString("trigger", backupClient.Alerting.Trigger)
		alertingWriter.SetStringSlice("emails", backupClient.Alerting.Emails)

		backupClientProperties["alert"] = []interface{}{alertingProperties}
	}
}

// ReadBackupClientDetail populates the ServerBackupClient with values from the specified BackupClientDetail.
func (backupClient *ServerBackupClient) ReadBackupClientDetail(backupClientDetail compute.BackupClientDetail) {
	backupClient.ID = backupClientDetail.ID
	backupClient.Type = backupClientDetail.Type
	backupClient.Description = backupClientDetail.Description
	backupClient.StoragePolicyName = backupClientDetail.StoragePolicyName
	backupClient.SchedulePolicyName = backupClientDetail.SchedulePolicyName
	backupClient.DownloadURL = backupClientDetail.DownloadURL
	backupClient.Status = backupClientDetail.Status
	if backupClientDetail.Alerting != nil {
		backupClient.Alerting = &BackupClientAlerting{
			Trigger: backupClientDetail.Alerting.Trigger,
			Emails:  backupClientDetail.Alerting.EmailAddresses, // TODO: Make a copy
		}
	} else {
		backupClient.Alerting = nil
	}
}

// ToBackupClientDetail updates a map using values from the ServerBackupClient.
func (backupClient *ServerBackupClient) ToBackupClientDetail() compute.BackupClientDetail {
	backupClientDetail := compute.BackupClientDetail{}
	backupClient.UpdateBackupClientDetail(&backupClientDetail)

	return backupClientDetail
}

// UpdateBackupClientDetail updates a CloudControl BackupClientDetail using values from the ServerBackupClient.
func (backupClient *ServerBackupClient) UpdateBackupClientDetail(backupClientDetail *compute.BackupClientDetail) {
	backupClientDetail.ID = backupClient.ID
	backupClientDetail.Type = backupClient.Type
	backupClientDetail.Description = backupClient.Description
	backupClientDetail.StoragePolicyName = backupClient.StoragePolicyName
	backupClientDetail.SchedulePolicyName = backupClient.SchedulePolicyName
	backupClientDetail.DownloadURL = backupClient.DownloadURL
	backupClientDetail.Status = backupClient.Status

	if backupClient.Alerting != nil {
		backupClientDetail.Alerting = &compute.BackupClientAlerting{
			Trigger:        backupClient.Alerting.Trigger,
			EmailAddresses: backupClient.Alerting.Emails, // TODO: Make a copy
		}
	} else {
		backupClientDetail.Alerting = nil
	}
}

// NewServerBackupClientFromMap creates a ServerBackupClient from the values in the specified map.
func NewServerBackupClientFromMap(backupClientProperties map[string]interface{}) ServerBackupClient {
	backupClient := ServerBackupClient{}
	backupClient.ReadMap(backupClientProperties)

	return backupClient
}

// NewServerBackupClientFromBackupClientDetail creates a ServerBackupClient from the values in the specified CloudControl BackupClientDetail.
func NewServerBackupClientFromBackupClientDetail(backupClientDetail compute.BackupClientDetail) ServerBackupClient {
	backupClient := ServerBackupClient{}
	backupClient.ReadBackupClientDetail(backupClientDetail)

	return backupClient
}
