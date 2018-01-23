package ddcloud

import (
	"fmt"
	"log"
	"time"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/validators"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

const (
	resourceKeyServerBackupServerID                 = "server"
	resourceKeyServerBackupServicePlan              = "service_plan"
	resourceKeyServerBackupAssetID                  = "asset_id"
	resourceKeyServerBackupClients                  = "client"
	resourceKeyServerBackupClientID                 = "id"
	resourceKeyServerBackupClientType               = "type"
	resourceKeyServerBackupClientDescription        = "description"
	resourceKeyServerBackupClientStoragePolicyName  = "storage_policy"
	resourceKeyServerBackupClientSchedulePolicyName = "schedule_policy"
	resourceKeyServerBackupClientDownloadURL        = "download_url"
	resourceKeyServerBackupClientAlert              = "alert"
	resourceKeyServerBackupClientAlertTrigger       = "trigger"
	resourceKeyServerBackupClientAlertEmails        = "emails"

	resourceCreateTimeoutServerBackup = 10 * time.Minute
)

func resourceServerBackup() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		Create:        resourceServerBackupCreate,
		Read:          resourceServerBackupRead,
		Update:        resourceServerBackupUpdate,
		Delete:        resourceServerBackupDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceServerBackupImport,
		// },

		Schema: map[string]*schema.Schema{
			resourceKeyServerBackupServerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the target server",
			},
			resourceKeyServerBackupServicePlan: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The backup service plan",
			},
			resourceKeyServerBackupAssetID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server's Cloud Backup asset Id",
			},
			resourceKeyServerBackupClients: &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						resourceKeyServerBackupClientID: &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The backup client's Id",
						},
						resourceKeyServerBackupClientType: &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The backup client type",
						},
						resourceKeyServerBackupClientDescription: &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A description of the backup client",
						},
						resourceKeyServerBackupClientDownloadURL: &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The backup client's download URL",
						},
						resourceKeyServerBackupClientStoragePolicyName: &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the backup client's assigned storage policy",
						},
						resourceKeyServerBackupClientSchedulePolicyName: &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the backup client's assigned schedule policy",
						},
						// TODO: Add resourceKeyServerBackupClientStatus
						resourceKeyServerBackupClientAlert: &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									resourceKeyServerBackupClientAlertTrigger: &schema.Schema{
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "",
										Description:  "If alerts are enabled, one of 'ON_FAILURE', 'ON_SUCCESS', 'ON_SUCCESS_OR_FAILURE'",
										ValidateFunc: validators.StringIsOneOf("ON_FAILURE", "ON_SUCCESS", "ON_SUCCESS_OR_FAILURE"),
									},
									resourceKeyServerBackupClientAlertEmails: &schema.Schema{
										Type:        schema.TypeList,
										Optional:    true,
										MinItems:    1,
										Description: "If alerts are enabled, the email address to which alerts will be sent",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
							Description: "Backup alerting configuration",
						},
					},
				},
				Description: "The server's assigned backup clients",
			},
		},
	}
}

// Create a server backup resource.
func resourceServerBackupCreate(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Get(resourceKeyServerBackupServerID).(string)
	servicePlan := data.Get(resourceKeyServerBackupServicePlan).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	log.Printf("Enabling backup for server '%s'...", serverID)

	operationDescription := fmt.Sprintf("Enable backup for server '%s'.", server.Name)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		enableError := apiClient.EnableServerBackup(serverID, servicePlan)
		if enableError != nil {
			if compute.IsResourceBusyError(enableError) || compute.IsAPIErrorCode(enableError, compute.ResultCodeBackupEnablementInProgressForServer) {
				context.Retry()
			} else if compute.IsAPIErrorCode(enableError, compute.ResultCodeBackupEnabledForServer) {
				// Backup is already enabled; proceed (if there's service plan mismatch, it will be resolved in the next apply-cycle).
			} else {
				context.Fail(enableError)
			}
		}
	})
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForServerBackupStatus(serverID, "enable backup", compute.ResourceStatusNormal, resourceCreateTimeoutServerBackup)
	if err != nil {
		return errors.Wrapf(err, "timed out waiting to enable backup for server '%s'", serverID)
	}

	backupDetails, err := apiClient.GetServerBackupDetails(serverID)
	if err != nil {
		return err
	}
	if backupDetails == nil {
		return fmt.Errorf("cannot find backup details for server '%s'", serverID)
	}

	data.SetId(serverID)
	data.Set(resourceKeyServerBackupAssetID, backupDetails.AssetID)

	propertyHelper := propertyHelper(data)
	backupClients := propertyHelper.GetServerBackupClients()

	if backupClients.IsEmpty() {
		return nil
	}

	log.Printf("Adding backup clients to server '%s'...", serverID)

	return createBackupClients(server, backupClients, data, providerState)
}

// Read a server backup resource.
func resourceServerBackupRead(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Get(resourceKeyServerBackupServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		log.Printf("cannot find server '%s' (will treat as deleted)", serverID)

		data.SetId("")

		return nil
	}

	log.Printf("Read backup details for server '%s'.", server.Name)

	backupDetails, err := apiClient.GetServerBackupDetails(serverID)
	if err != nil {
		return err
	}
	if backupDetails == nil {
		log.Printf("backup is not enabled for server '%s' (will treat as deleted)", serverID)

		data.SetId("")

		return nil
	}

	data.Set(resourceKeyServerBackupAssetID, backupDetails.AssetID)
	data.Set(resourceKeyServerBackupServicePlan, backupDetails.ServicePlan)

	propertyHelper := propertyHelper(data)
	propertyHelper.SetServerBackupClients(
		models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients),
	)

	return nil
}

// Update a server backup resource.
func resourceServerBackupUpdate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Get(resourceKeyServerBackupServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	backupDetails, err := apiClient.GetServerBackupDetails(serverID)
	if err != nil {
		return err
	}
	if backupDetails == nil {
		log.Printf("Backup is not enabled for server '%s' (will treat ddcloud_server_backup resource as deleted).", serverID)

		data.SetId("")

		return nil
	}

	if data.HasChange(resourceKeyServerBackupServicePlan) {
		servicePlan := data.Get(resourceKeyServerBackupServicePlan).(string)

		log.Printf("Change backup service plan for server '%s' to '%s'.",
			serverID,
			servicePlan,
		)

		operationDescription := fmt.Sprintf("Change backup service plan for server '%s'.", server.Name)
		err = providerState.RetryAction(operationDescription, func(context retry.Context) {
			changeServicePlanError := apiClient.ChangeServerBackupServicePlan(serverID, servicePlan)
			if changeServicePlanError != nil {
				if compute.IsResourceBusyError(changeServicePlanError) {
					context.Retry()
				} else {
					context.Fail(changeServicePlanError)
				}
			}
		})
		if err != nil {
			return err
		}
	}

	configuredBackupClients := propertyHelper.GetServerBackupClients()
	actualBackupClients := models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients)
	addedBackupClients, changedBackupClients, removedBackupClients := configuredBackupClients.SplitByAction(actualBackupClients)

	err = deleteBackupClients(server, removedBackupClients, data, providerState)
	if err != nil {
		return err
	}

	err = createBackupClients(server, addedBackupClients, data, providerState)
	if err != nil {
		return err
	}

	err = updateBackupClients(server, changedBackupClients, data, providerState)
	if err != nil {
		return err
	}

	return nil
}

// Delete a server backup resource.
func resourceServerBackupDelete(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Get(resourceKeyServerBackupServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	backupDetails, err := apiClient.GetServerBackupDetails(serverID)
	if err != nil {
		return err
	}
	if backupDetails == nil {
		log.Printf("Backup is not enabled for server '%s' (will treat ddcloud_server_backup resource as deleted).", serverID)

		data.SetId("")

		return nil
	}

	log.Printf("Remove backup clients (if any) for server '%s'.", serverID)

	backupClientsToRemove := models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients)
	err = deleteBackupClients(server, backupClientsToRemove, data, providerState)
	if err != nil {
		return err
	}

	log.Printf("Disable backup for server '%s'.", serverID)

	operationDescription := fmt.Sprintf("Disable backup for server '%s'.", server.Name)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		disableError := apiClient.DisableServerBackup(serverID)
		if disableError != nil {
			if compute.IsResourceBusyError(disableError) {
				context.Retry()
			} else {
				context.Fail(disableError)
			}
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func createBackupClients(server *compute.Server, backupClients models.ServerBackupClients, data *schema.ResourceData, providerState *providerState) error {
	if len(backupClients) == 0 {
		return nil
	}

	propertyHelper := propertyHelper(data)
	apiClient := providerState.Client()

	log.Printf("Add %d backup clients to server '%s'.", len(backupClients), server.ID)

	for index := range backupClients {
		addBackupClient := backupClients[index]

		log.Printf("Adding '%s' backup client to server '%s'...", addBackupClient.Type, server.ID)

		var backupClientID string
		operationDescription := fmt.Sprintf("Add '%s' backup client to server '%s'.", addBackupClient.Type, server.Name)
		err := providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			var addClientError error
			backupClientID, _, addClientError = apiClient.AddServerBackupClient(server.ID, addBackupClient.Type, addBackupClient.SchedulePolicyName, addBackupClient.StoragePolicyName, nil /* TODO: Add alerting configuration */)
			if addClientError != nil {
				if compute.IsResourceBusyError(addClientError) {
					context.Retry()
				} else {
					context.Fail(addClientError)
				}
			}
		})
		if err != nil {
			return errors.Wrapf(err, "failed to add '%s' backup client to server '%s'", addBackupClient.Type, server.ID)
		}

		_, err = apiClient.WaitForServerBackupStatus(server.ID, "add backup client", compute.ResourceStatusNormal, resourceCreateTimeoutServerBackup)
		if err != nil {
			return errors.Wrapf(err, "timed out waiting to add '%s' backup client for server '%s'", addBackupClient.ID, server.ID)
		}

		backupDetails, err := apiClient.GetServerBackupDetails(server.ID)
		if err != nil {
			return err
		}
		if backupDetails == nil {
			return fmt.Errorf("cannot find backup details for server '%s'", server.ID)
		}

		// Persist
		propertyHelper.SetServerBackupClients(
			models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients),
		)
	}

	return nil
}

func updateBackupClients(server *compute.Server, backupClients models.ServerBackupClients, data *schema.ResourceData, providerState *providerState) error {
	if len(backupClients) == 0 {
		return nil
	}

	propertyHelper := propertyHelper(data)
	apiClient := providerState.Client()

	log.Printf("Update %d backup clients for server '%s'.", len(backupClients), server.ID)

	for index := range backupClients {
		backupClient := backupClients[index]

		log.Printf("Modifying '%s' backup client of server '%s'...", backupClient.Type, server.ID)

		operationDescription := fmt.Sprintf("Modify backup client '%s' of server '%s'.", backupClient.ID, server.Name)
		err := providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			var changeClientError error
			_, changeClientError = apiClient.ModifyServerBackupClient(server.ID, backupClient.ID,
				backupClient.SchedulePolicyName,
				backupClient.StoragePolicyName,
				nil, /* TODO: Add alerting configuration */
			)
			if changeClientError != nil {
				if compute.IsResourceBusyError(changeClientError) {
					context.Retry()
				} else {
					context.Fail(changeClientError)
				}
			}
		})
		if err != nil {
			return err
		}

		_, err = apiClient.WaitForServerBackupStatus(server.ID, "modify backup client", compute.ResourceStatusNormal, resourceCreateTimeoutServerBackup)
		if err != nil {
			return errors.Wrapf(err, "timed out waiting to modify backup client '%s' of server '%s'", backupClient.ID, server.ID)
		}

		backupDetails, err := apiClient.GetServerBackupDetails(server.ID)
		if err != nil {
			return err
		}
		if backupDetails == nil {
			return fmt.Errorf("cannot find backup details for server '%s'", server.ID)
		}

		// Persist
		propertyHelper.SetServerBackupClients(
			models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients),
		)
	}

	return nil
}

func deleteBackupClients(server *compute.Server, backupClients models.ServerBackupClients, data *schema.ResourceData, providerState *providerState) error {
	if len(backupClients) == 0 {
		return nil
	}

	propertyHelper := propertyHelper(data)
	apiClient := providerState.Client()

	log.Printf("Remove %d backup clients from server '%s'.", len(backupClients), server.ID)

	for index := range backupClients {
		backupClient := backupClients[index]

		log.Printf("Removing '%s' backup client '%s' from server '%s'...", backupClient.Type, backupClient.ID, server.ID)

		operationDescription := fmt.Sprintf("Remove '%s' backup client '%s' from server '%s'.", backupClient.Type, backupClient.ID, server.Name)
		err := providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			log.Printf("Attempting to cancel all outstanding jobs for '%s' backup client '%s' of server '%s'...", backupClient.Type, backupClient.ID, server.ID)
			cancelJobsError := apiClient.CancelBackupClientJobs(server.ID, backupClient.ID)
			if cancelJobsError != nil {
				if compute.IsAPIErrorCode(cancelJobsError, compute.ResultCodeBackupClientNotFound) {
					// Client has never actually been installed on the target server; it's safe to remove it.
				} else {
					if compute.IsResourceBusyError(cancelJobsError) {
						context.Retry()
					} else {
						context.Fail(cancelJobsError)
					}

					return
				}
			}

			log.Printf("Attempting to remove '%s' backup client '%s' from server '%s'...", backupClient.Type, backupClient.ID, server.ID)
			removeClientError := apiClient.RemoveServerBackupClient(server.ID, backupClient.ID)
			if removeClientError != nil {
				if compute.IsResourceBusyError(removeClientError) {
					context.Retry()
				} else {
					context.Fail(removeClientError)
				}
			}
		})
		if err != nil {
			return err
		}

		_, err = apiClient.WaitForServerBackupStatus(server.ID, "remove backup client", compute.ResourceStatusNormal, resourceCreateTimeoutServerBackup)
		if err != nil {
			return errors.Wrapf(err, "timed out waiting to remove '%s' backup client '%s' of server '%s'", backupClient.Type, backupClient.ID, server.ID)
		}

		backupDetails, err := apiClient.GetServerBackupDetails(server.ID)
		if err != nil {
			return err
		}
		if backupDetails == nil {
			return fmt.Errorf("cannot find backup details for server '%s'", server.ID)
		}

		// Persist
		propertyHelper.SetServerBackupClients(
			models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients),
		)
	}

	return nil
}
