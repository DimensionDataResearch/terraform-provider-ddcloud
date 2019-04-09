package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyNetworkAdapterServerID    = "server"
	resourceKeyNetworkAdapterMACAddress  = "mac"
	resourceKeyNetworkAdapterKey         = "mac"
	resourceKeyNetworkAdapterVLANID      = "vlan"
	resourceKeyNetworkAdapterPrivateIPV4 = "ipv4"
	resourceKeyNetworkAdapterPrivateIPV6 = "ipv6"
	resourceKeyNetworkAdapterType        = "type"
)

func resourceNetworkAdapter() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkAdapterCreate,
		Exists: resourceNetworkAdapterExists,
		Read:   resourceNetworkAdapterRead,
		Update: resourceNetworkAdapterUpdate,
		Delete: resourceNetworkAdapterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNetworkAdapterImport,
		},

		Schema: map[string]*schema.Schema{
			resourceKeyNetworkAdapterServerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the server to which the additional nics needs to be updated",
			},

			resourceKeyNetworkAdapterVLANID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "VLAN ID of the nic",
				ForceNew:    true,
			},
			resourceKeyNetworkAdapterPrivateIPV4: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Private IPV4 address for the nic",
			},
			resourceKeyNetworkAdapterPrivateIPV6: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Private IPV6 Address for the nic",
			},
			resourceKeyNetworkAdapterType: &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      nil,
				Description:  "The type of network adapter (E1000 or VMXNET3)",
				ValidateFunc: validateNetworkAdapterAdapterType,
			},
		},
	}

}

func resourceNetworkAdapterCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)
	serverID := data.Get(resourceKeyNetworkAdapterServerID).(string)
	ipv4Address := data.Get(resourceKeyNetworkAdapterPrivateIPV4).(string)

	vlanID := data.Get(resourceKeyNetworkAdapterVLANID).(string)
	adapterType := propertyHelper.GetOptionalString(resourceKeyNetworkAdapterType, false)

	log.Printf("Configure additional nics for server '%s'...", serverID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server with '%s'", serverID)
	}

	isStarted := server.Started
	if isStarted {
		err = serverShutdown(providerState, serverID)
		if err != nil {
			return err
		}
	}

	log.Printf("Add network adapter to server '%s'...", serverID)

	var networkAdapterID string
	operationDescription := fmt.Sprintf("Add network adapter to server '%s'", serverID)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		var addError error
		if adapterType != nil {
			networkAdapterID, addError = apiClient.AddNicWithTypeToServer(serverID, ipv4Address, vlanID, *adapterType)
		} else {
			networkAdapterID, addError = apiClient.AddNicToServer(serverID, ipv4Address, vlanID)
		}

		if compute.IsResourceBusyError(addError) {
			context.Retry()
		} else if addError != nil {
			context.Fail(addError)
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}
	data.SetId(networkAdapterID)

	log.Printf("Adding network adapter '%s' to server '%s'...",
		networkAdapterID,
		serverID,
	)

	_, err = apiClient.WaitForChange(
		compute.ResourceTypeServer,
		serverID,
		"Add network adapter",
		resourceUpdateTimeoutServer,
	)
	if err != nil {
		return err
	}

	log.Printf("created the nic with the id %s", networkAdapterID)
	if isStarted {
		err = serverStart(providerState, serverID)
		if err != nil {
			return err
		}
	}

	log.Printf("Refresh properties for network adapter '%s' in server '%s'", networkAdapterID, serverID)
	server, err = apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	serverNetworkAdapters := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network)
	serverNetworkAdapter := serverNetworkAdapters.GetByID(networkAdapterID)
	if serverNetworkAdapter == nil {
		data.SetId("") // NetworkAdapter deleted

		return fmt.Errorf("Newly-created network adapter (Id = '%s') not found", networkAdapterID)
	}
	if err != nil {
		return err
	}

	data.Set(resourceKeyNetworkAdapterPrivateIPV4, serverNetworkAdapter.PrivateIPv4Address)
	data.Set(resourceKeyNetworkAdapterVLANID, serverNetworkAdapter.VLANID)
	data.Set(resourceKeyNetworkAdapterPrivateIPV6, serverNetworkAdapter.PrivateIPv6Address)
	data.Set(resourceKeyNetworkAdapterPrivateIPV4, serverNetworkAdapter.PrivateIPv4Address)

	return nil
}

func resourceNetworkAdapterExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	nicID := data.Id()
	serverID := data.Get(resourceKeyNetworkAdapterServerID).(string)

	apiClient := provider.(*providerState).Client()

	log.Printf("Get the server with Id %s", serverID)

	server, err := apiClient.GetServer(serverID)

	if server == nil {
		log.Printf("Server '%s' not found; will treat network adapter '%s' as non-existent.", serverID, nicID)

		return false, nil
	}

	if err != nil {
		return false, err
	}

	serverNetworkAdapters := server.Network.AdditionalNetworkAdapters
	for _, nic := range serverNetworkAdapters {

		if *nic.ID == nicID {
			return true, nil
		}
	}

	return false, nil
}

func resourceNetworkAdapterRead(data *schema.ResourceData, provider interface{}) error {
	networkAdapterID := data.Id()
	serverID := data.Get(resourceKeyNetworkAdapterServerID).(string)

	log.Printf("Get the server with the ID %s", serverID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("Server '%s' not found", serverID)
	}

	serverNetworkAdapter := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network).GetByID(networkAdapterID)
	if serverNetworkAdapter == nil {
		log.Printf("Network adapter with Id '%s' was not found in server %s (will treat as deleted).",
			networkAdapterID,
			serverID,
		)
		data.SetId("") // Deleted.

		return nil
	}

	data.Set(resourceKeyNetworkAdapterPrivateIPV4, serverNetworkAdapter.PrivateIPv4Address)
	data.Set(resourceKeyNetworkAdapterVLANID, serverNetworkAdapter.VLANID)
	data.Set(resourceKeyNetworkAdapterPrivateIPV6, serverNetworkAdapter.PrivateIPv6Address)
	data.Set(resourceKeyNetworkAdapterPrivateIPV4, serverNetworkAdapter.PrivateIPv4Address)

	return nil
}

func resourceNetworkAdapterUpdate(data *schema.ResourceData, provider interface{}) error {
	networkAdapterID := data.Id()
	serverID := data.Get(resourceKeyNetworkAdapterServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("Server '%s' not found", serverID)
	}

	serverNetworkAdapter := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network).GetByID(networkAdapterID)
	if serverNetworkAdapter == nil {
		log.Printf("Network adapter '%s' not found in server '%s'; will treat as deleted.",
			networkAdapterID,
			serverID,
		)
		data.SetId("")

		return nil
	}

	configuredNetworkAdapter := propertyHelper(data).GetNetworkAdapter()
	if data.HasChange(resourceKeyNetworkAdapterPrivateIPV4) {
		err := modifyServerNetworkAdapterIP(providerState, serverID, configuredNetworkAdapter)
		if err != nil {
			return err
		}
	}
	if data.HasChange(resourceKeyNetworkAdapterType) {
		err := modifyServerNetworkAdapterType(providerState, serverID, configuredNetworkAdapter)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceNetworkAdapterDelete(data *schema.ResourceData, provider interface{}) error {
	networkAdapterID := data.Id()
	serverID := data.Get(resourceKeyNetworkAdapterServerID).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	log.Printf("Remove network adapter '%s' from server '%s'.", networkAdapterID, serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("cannot find server '%s'", serverID)
	}

	serverNetworkAdapter := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network).GetByID(networkAdapterID)
	if serverNetworkAdapter == nil {
		log.Printf("Network adapter '%s' not found in server '%s' (will treat as deleted).",
			networkAdapterID,
			serverID,
		)

		return nil
	}

	isStarted := server.Started
	if isStarted {
		err = serverShutdown(providerState, serverID)
		if err != nil {
			return err
		}
	}

	operationDescription := fmt.Sprintf("Remove network adapter '%s' from server '%s'", networkAdapterID, serverID)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release()

		removeError := apiClient.RemoveNicFromServer(networkAdapterID)
		if removeError == nil {
			if compute.IsResourceBusyError(removeError) {
				context.Retry()
			} else {
				context.Fail(removeError)
			}
		}
	})
	if err != nil {
		return err
	}

	log.Printf("Removing network adapter '%s' from server '%s'...",
		networkAdapterID,
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

	data.SetId("") // Resource deleted.

	log.Printf("Removed network adapter '%s' from server '%s'.",
		networkAdapterID,
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

// Import data for an existing network adapter.
func resourceNetworkAdapterImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	networkAdapterID := data.Id()
	serverID := data.Get(resourceKeyNetworkAdapterServerID).(string)
	log.Printf("Import network adapter '%s' in server '%s'.", networkAdapterID, serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return
	}
	if server == nil {
		err = fmt.Errorf("Server '%s' not found", serverID)

		return
	}
	serverNetworkAdapter := models.NewNetworkAdaptersFromVirtualMachineNetwork(server.Network).GetByID(networkAdapterID)
	if serverNetworkAdapter == nil {
		err = fmt.Errorf("Network adapter '%s' not found in server '%s'", networkAdapterID, serverID)

		return
	}

	data.Set(resourceKeyNetworkAdapterType, serverNetworkAdapter.AdapterType)
	data.Set(resourceKeyNetworkAdapterMACAddress, serverNetworkAdapter.MACAddress)
	data.Set(resourceKeyNetworkAdapterVLANID, serverNetworkAdapter.VLANID)
	data.Set(resourceKeyNetworkAdapterPrivateIPV4, serverNetworkAdapter.PrivateIPv4Address)
	data.Set(resourceKeyNetworkAdapterPrivateIPV6, serverNetworkAdapter.PrivateIPv6Address)

	importedData = []*schema.ResourceData{data}

	return
}

func validateNetworkAdapterAdapterType(value interface{}, propertyName string) (messages []string, errors []error) {
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
	case compute.NetworkAdapterTypeE1000E:
	case compute.NetworkAdapterTypeEnhancedVMXNET2:
	case compute.NetworkAdapterTypeFlexiblePCNET32:
	case compute.NetworkAdapterTypeVMXNET3:
		break
	default:
		errors = append(errors,
			fmt.Errorf("invalid network adapter type '%s'", value),
		)
	}

	return
}
