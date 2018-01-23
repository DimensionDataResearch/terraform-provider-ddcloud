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

	data.Partial(true)
	data.Set(resourceKeyServerBackupAssetID, backupDetails.AssetID)
	data.SetPartial(resourceKeyServerBackupAssetID)

	propertyHelper := propertyHelper(data)
	backupClients := propertyHelper.GetServerBackupClients()

	if backupClients.IsEmpty() {
		return nil
	}

	log.Printf("Adding backup clients to server '%s'...", serverID)

	data.SetPartial(resourceKeyServerBackupClients)

	for index := range backupClients {
		backupClient := backupClients[index]

		log.Printf("Adding '%s' backup client to server '%s'...", backupClient.Type, serverID)

		var clientID string
		operationDescription := fmt.Sprintf("Enable backup for server '%s'.", server.Name)
		err = providerState.RetryAction(operationDescription, func(context retry.Context) {
			asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
			defer asyncLock.Release()

			var addClientError error
			clientID, _, addClientError = apiClient.AddServerBackupClient(serverID, backupClient.Type, backupClient.SchedulePolicyName, backupClient.StoragePolicyName, nil /* TODO: Add alerting configuration */)
			if addClientError != nil {
				if compute.IsResourceBusyError(addClientError) {
					context.Retry()
				} else {
					context.Fail(addClientError)
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

		// Update computed properties
		serverBackupClientsByID := models.NewServerBackupClientsFromBackupClientDetails(backupDetails.Clients).ByID()
		backupClients[index] = serverBackupClientsByID[clientID]

		// Persist
		propertyHelper.SetServerBackupClients(backupClients)
	}

	data.Partial(false)

	return nil
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
		data.SetId("")

		return fmt.Errorf("cannot find server '%s' (will treat as deleted)", serverID)
	}

	log.Printf("Read backup details for server '%s'.", server.Name)

	backupDetails, err := apiClient.GetServerBackupDetails(serverID)
	if err != nil {
		return err
	}
	if backupDetails == nil {
		return fmt.Errorf("cannot find backup details for server '%s'", serverID)
	}

	data.Set(resourceKeyServerBackupAssetID, backupDetails.AssetID)
	data.Set(resourceKeyServerBackupServicePlan, backupDetails.ServicePlan)

	return nil
}

// Update a server backup resource.
func resourceServerBackupUpdate(data *schema.ResourceData, provider interface{}) error {
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

	// TODO: Add / remove agents as required.

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

	// TODO: Remove backup clients (if any).

	log.Printf("Disable backup for server '%s'.", serverID)

	backupDetails, err := apiClient.GetServerBackupDetails(serverID)
	if err != nil {
		return err
	}
	if backupDetails == nil {
		log.Printf("Backup is not enabled for server '%s' (will treat ddcloud_server_backup resource as deleted).", serverID)

		data.SetId("")

		return nil
	}

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
