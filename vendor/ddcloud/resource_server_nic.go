package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyNICServerID    = "server"
	resourceKeyNICVLANID      = "vlan"
	resourceKeyNICPrivateIPV4 = "private_ipv4"
	resourceKeyNICPrivateIPV6 = "private_ipv6"
	resourceKeyNICAdapterType = "adapter_type"
)

func resourceServerNIC() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerNICCreate,
		Exists: resourceServerNICExists,
		Read:   resourceServerNICRead,
		Update: resourceServerNICUpdate,
		Delete: resourceServerNICDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNICServerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the server to which the additional nics needs to be updated",
			},

			resourceKeyNICVLANID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "VLAN ID of the nic",
				ForceNew:    true,
			},
			resourceKeyNICPrivateIPV4: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Private IPV4 address for the nic",
			},
			resourceKeyNICPrivateIPV6: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Private IPV6 Address for the nic",
			},
			resourceKeyNICAdapterType: &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      nil,
				Description:  "The type of network adapter (E1000 or VMXNET3)",
				ValidateFunc: validateNICAdapterType,
			},
		},
	}

}

func resourceServerNICCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Get(resourceKeyNICServerID).(string)
	ipv4Address := data.Get(resourceKeyNICPrivateIPV4).(string)
	vlanID := data.Get(resourceKeyNICVLANID).(string)
	adapterType := propertyHelper.GetOptionalString(resourceKeyNICAdapterType, false)

	log.Printf("Configure additional nics for server '%s'...", serverID)

	providerState := provider.(*providerState)
	serverLock := providerState.GetServerLock(serverID, "resourceServerNICCreate(id = '%s')", serverID)
	serverLock.Lock()
	defer serverLock.Unlock()

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	isStarted := server.Started
	if isStarted {
		err = serverShutdown(providerState, serverID)
		if err != nil {
			return err
		}
	}

	log.Printf("Add network adapter to server '%s'...", serverID)

	// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
	asyncLock := providerState.AcquireAsyncOperationLock("Add network adapter to server '%s'", serverID)
	defer asyncLock.Release()

	var nicID string
	if adapterType != nil {
		nicID, err = apiClient.AddNicWithTypeToServer(serverID, ipv4Address, vlanID, *adapterType)
	} else {
		nicID, err = apiClient.AddNicToServer(serverID, ipv4Address, vlanID)
	}
	if err != nil {
		return err
	}

	// Operation initiated; we no longer need this lock.
	asyncLock.Release()

	log.Printf("Adding network adapter '%s' to server '%s'...",
		nicID,
		serverID,
	)

	_, err = apiClient.WaitForChange(
		compute.ResourceTypeServer,
		serverID,
		"Add nic",
		resourceUpdateTimeoutServer,
	)
	if err != nil {
		return err
	}

	data.SetId(nicID) //NIC created
	log.Printf("created the nic with the id %s", nicID)

	if isStarted {
		err = serverStart(providerState, serverID)
		if err != nil {
			return err
		}
	}

	log.Printf("read the nic with the id %s to set the computed properties", nicID)
	server, err = apiClient.GetServer(serverID)

	if server == nil {
		log.Printf("server with the id %s cannot be found", serverID)
	}

	if err != nil {
		return err
	}

	serverNICs := server.Network.AdditionalNetworkAdapters

	var serverNIC compute.VirtualMachineNetworkAdapter
	for _, nic := range serverNICs {
		if *nic.ID == nicID {
			serverNIC = nic
			break
		}
	}

	if serverNIC.ID == nil {
		log.Printf("NIC with the id %s doesn't exists", nicID)
		data.SetId("") // NIC deleted
		return nil
	}
	if err != nil {
		return err
	}
	data.Set(resourceKeyNICPrivateIPV4, serverNIC.PrivateIPv4Address)
	data.Set(resourceKeyNICVLANID, serverNIC.VLANID)
	data.Set(resourceKeyNICPrivateIPV6, serverNIC.PrivateIPv6Address)
	data.Set(resourceKeyNICPrivateIPV4, serverNIC.PrivateIPv4Address)

	return nil
}

func resourceServerNICExists(data *schema.ResourceData, provider interface{}) (bool, error) {

	nicExists := false

	serverID := data.Get(resourceKeyNICServerID).(string)

	apiClient := provider.(*providerState).Client()

	nicID := data.Id()

	log.Printf("Get the server with the ID %s", serverID)

	server, err := apiClient.GetServer(serverID)

	if server == nil {
		log.Printf("server with the id %s cannot be found", serverID)
	}

	if err != nil {
		return nicExists, err
	}
	serverNICs := server.Network.AdditionalNetworkAdapters
	for _, nic := range serverNICs {

		if *nic.ID == nicID {
			nicExists = true
			break
		}
	}
	return nicExists, nil
}

func resourceServerNICRead(data *schema.ResourceData, provider interface{}) error {

	id := data.Id()

	serverID := data.Get(resourceKeyNICServerID).(string)

	log.Printf("Get the server with the ID %s", serverID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("server with the id %s cannot be found", serverID)
	}

	serverNICs := server.Network.AdditionalNetworkAdapters

	var serverNIC compute.VirtualMachineNetworkAdapter
	for _, nic := range serverNICs {
		if *nic.ID == id {
			serverNIC = nic
			break
		}
	}

	if serverNIC.ID == nil {
		log.Printf("NIC with the id %s doesn't exists", id)
		data.SetId("") // NIC deleted
		return nil
	}

	if err != nil {
		return err
	}
	data.Set(resourceKeyNICPrivateIPV4, serverNIC.PrivateIPv4Address)
	data.Set(resourceKeyNICVLANID, serverNIC.VLANID)
	data.Set(resourceKeyNICPrivateIPV6, serverNIC.PrivateIPv6Address)
	data.Set(resourceKeyNICPrivateIPV4, serverNIC.PrivateIPv4Address)

	return nil
}

func resourceServerNICUpdate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)
	nicID := data.Id()
	serverID := data.Get(resourceKeyNICServerID).(string)
	privateIPV4 := propertyHelper.GetOptionalString(resourceKeyNICPrivateIPV4, true)

	providerState := provider.(*providerState)
	serverLock := providerState.GetServerLock(serverID, "resourceServerNICUpdate(id = '%s', serverID = '%s')", nicID, serverID)
	serverLock.Lock()
	defer serverLock.Unlock()

	if data.HasChange(resourceKeyNICPrivateIPV4) {
		log.Printf("changing the ip address of the nic with the id %s to %s", nicID, *privateIPV4)
		err := updateNICIPAddress(providerState, serverID, nicID, privateIPV4)
		if err != nil {
			return err
		}
		log.Printf("IP address of the nic with the id %s changed to %s", nicID, *privateIPV4)
	}

	return nil
}

func resourceServerNICDelete(data *schema.ResourceData, provider interface{}) error {
	nicID := data.Id()
	serverID := data.Get(resourceKeyNICServerID).(string)

	providerState := provider.(*providerState)
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	log.Printf("Removing additional nics for server '%s'...", serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	serverLock := providerState.GetServerLock(serverID, "resourceServerNICDelete(id = '%s', serverID = '%s')", nicID, serverID)
	serverLock.Lock()
	defer serverLock.Unlock()

	isStarted := server.Started
	if isStarted {
		err = serverShutdown(providerState, serverID)
		if err != nil {
			return err
		}
	}

	log.Printf("Remove network adapter '%s' from server '%s'.", nicID, serverID)

	operationDescription := fmt.Sprintf("Remove network adapter '%s' from server '%s'", nicID, serverID)
	err = providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Remove NIC from server '%s'", serverID)
		defer asyncLock.Release() // Released once the current attempt is complete

		removeError := apiClient.RemoveNicFromServer(nicID)
		if removeError == nil {
			if compute.IsResourceBusyError(removeError) {
				context.Retry()
			} else {
				context.Fail(removeError)
			}
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	log.Printf("Removing network adapter with ID %s from server '%s'...",
		nicID,
		serverID,
	)
	_, err = apiClient.WaitForChange(
		compute.ResourceTypeServer,
		serverID,
		"Remove nic",
		resourceUpdateTimeoutServer,
	)
	if err != nil {
		return err
	}

	data.SetId("") //NIC Deleted

	log.Printf("Removed network adapter with ID %s from server '%s'.",
		nicID,
		serverID,
	)

	if isStarted {
		err = serverStart(providerState, serverID)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateNICIPAddress notifies the compute infrastructure that a NIC's IP address has changed.
func updateNICIPAddress(providerState *providerState, serverID string, nicID string, primaryIPv4 *string) error {
	log.Printf("Update IP address(es) for nic '%s'...", nicID)

	// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
	asyncLock := providerState.AcquireAsyncOperationLock("Update IP address(es) for nic '%s'", nicID)
	defer asyncLock.Release()

	apiClient := providerState.Client()
	err := apiClient.NotifyServerIPAddressChange(nicID, primaryIPv4, nil)
	if err != nil {
		return err
	}

	// Operation initiated; we no longer need this lock.
	asyncLock.Release()

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, nicID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	return err
}

func validateNICAdapterType(value interface{}, propertyName string) (messages []string, errors []error) {
	if value == nil {
		return
	}

	adapterType, ok := value.(string)
	if !ok {
		errors = append(errors,
			fmt.Errorf("Unexpected value type '%v'", value),
		)

		return
	}

	switch adapterType {
	case compute.NetworkAdapterTypeE1000:
	case compute.NetworkAdapterTypeVMXNET3:
		break
	default:
		errors = append(errors,
			fmt.Errorf("Invalid network adapter type '%s'", value),
		)
	}

	return
}

// Shut a server down.
func serverShutdown(providerState *providerState, serverID string) error {
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	if !providerSettings.AllowServerReboots {
		return fmt.Errorf("Cannot shut down server '%s' because server reboots have not been enabled via the 'allow_server_reboot' provider setting or 'DDCLOUD_ALLOW_SERVER_REBOOT' environment variable", serverID)
	}

	operationDescription := fmt.Sprintf("Shut down server '%s'", serverID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Shut down server '%s'", serverID)
		defer asyncLock.Release() // Released when the current attempt is complete.

		shutdownError := apiClient.ShutdownServer(serverID)
		if compute.IsResourceBusyError(shutdownError) {
			context.Retry()
		} else if shutdownError != nil {
			context.Fail(shutdownError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Shut down server", serverShutdownTimeout)
	if err != nil {
		return err
	}

	return nil
}

// Start a server.
func serverStart(providerState *providerState, serverID string) error {
	providerSettings := providerState.Settings()
	apiClient := providerState.Client()

	if !providerSettings.AllowServerReboots {
		return fmt.Errorf("Cannot start server '%s' because server reboots have not been enabled via the 'allow_server_reboot' provider setting or 'DDCLOUD_ALLOW_SERVER_REBOOT' environment variable", serverID)
	}

	operationDescription := fmt.Sprintf("Start server '%s'", serverID)
	err := providerState.Retry().Action(operationDescription, providerSettings.RetryTimeout, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock("Start server '%s'", serverID)
		defer asyncLock.Release() // Released when the current attempt is complete.

		startError := apiClient.StartServer(serverID)
		if compute.IsResourceBusyError(startError) {
			context.Retry()
		} else if startError != nil {
			context.Fail(startError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Start server", serverShutdownTimeout)
	if err != nil {
		return err
	}

	return nil
}
