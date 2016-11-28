package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerNetworkAdapter       = "network_adapter"
	resourceKeyServerNetworkAdapterID     = "id"
	resourceKeyServerNetworkAdapterIndex  = "index"
	resourceKeyServerNetworkAdapterVLANID = "vlan"
	resourceKeyServerNetworkAdapterIPV4   = "ipv4"
	resourceKeyServerNetworkAdapterIPV6   = "ipv6"
	resourceKeyServerNetworkAdapterType   = "type"
)

func schemaServerNetworkAdapter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerNetworkAdapterID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The network adapter's identifier in CloudControl",
				},
				resourceKeyServerNetworkAdapterIndex: &schema.Schema{
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     0,
					Description: "A unique identifier for the network adapter (0 is the primary adapter)",
				},
				resourceKeyServerNetworkAdapterVLANID: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Optional:    true,
					Default:     nil,
					Description: "VLAN ID of the network adapter",
					ForceNew:    true,
				},
				resourceKeyServerNetworkAdapterIPV4: &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Default:     nil,
					Description: "The IPV4 address associated with the network adapter",
				},
				resourceKeyServerNetworkAdapterIPV6: &schema.Schema{
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The IPV6 Address associated the network adapter",
				},
				resourceKeyServerNetworkAdapterType: &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  compute.NetworkAdapterTypeE1000,
					Description: fmt.Sprintf("The type of network adapter (%s or %s)",
						compute.NetworkAdapterTypeE1000,
						compute.NetworkAdapterTypeVMXNET3,
					),
					ValidateFunc: validateNetworkAdapterAdapterType,
				},
			},
		},
		Set: hashServerNetworkAdapter,
		ConflictsWith: []string{
			resourceKeyServerPrimaryAdapterVLAN,
			resourceKeyServerPrimaryAdapterIPv4,
			resourceKeyServerPrimaryAdapterType,
		},
	}
}

// TODO: Define MapStructure-compatible structures to represent configured network adapters.
// TODO: Give these structures methods for reading / writing VirtualMachineNetworkAdapter.

func addNetworkAdapterToServer(apiClient *compute.Client, serverID string, ipv4Address *string, vlanID *string, adapterType *string) error {
	// TODO: Implement (remember to use providerState.AcquireAsyncOperationLock)

	return fmt.Errorf("addNetworkAdapterToServer is not yet implemented")
}

func removeNetworkAdapterFromServer(apiClient *compute.Client, serverID string, networkAdapterID string) error {
	// TODO: Implement (remember to use providerState.AcquireAsyncOperationLock)

	return fmt.Errorf("addNetworkAdapterToServer is not yet implemented")
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

// Validate that the specified value represents a valid network adapter type.
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
	case compute.NetworkAdapterTypeVMXNET3:
		break
	default:
		errors = append(errors,
			fmt.Errorf("Invalid network adapter type '%s'", value),
		)
	}

	return
}

// Create a hash code to represent the property values for a server network adapter.
func hashServerNetworkAdapter(value interface{}) int {
	adapterProperties, ok := value.(map[string]interface{})
	if !ok {
		return -1
	}

	var (
		index       int
		vlanID      string
		ipv4Address string
		adapterType string
	)
	adapterProperty, ok := adapterProperties[resourceKeyServerNetworkAdapterIndex]
	if ok {
		index = adapterProperty.(int)
	}

	adapterProperty, ok = adapterProperties[resourceKeyServerNetworkAdapterVLANID]
	if ok {
		vlanID = adapterProperty.(string)
	}

	adapterProperty, ok = adapterProperties[resourceKeyServerNetworkAdapterIPV4]
	if ok {
		ipv4Address = adapterProperty.(string)
	}

	adapterProperty, ok = adapterProperties[resourceKeyServerNetworkAdapterType]
	if ok {
		adapterType = adapterProperty.(string)
	}

	return schema.HashString(fmt.Sprintf(
		"%d|%s|%s|%s", index, vlanID, ipv4Address, adapterType,
	))
}

// Stop a server.
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
