package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyNicServerID    = "server"
	resourceKeyNicVLANID      = "vlan_id"
	resourceKeyNicPrivateIPV4 = "private_ipv4"
	resourceKeyNicPrivateIPV6 = "private_ipv6"
)

func resourceAdditionalNic() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdditionalNicCreate,
		Exists: resourceAdditionalNicExists,
		Read:   resourceAdditionalNicRead,
		Update: resourceAdditionalNicUpdate,
		Delete: resourceAdditionalNicDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNicServerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the server to which the additional nics needs to be updated",
			},

			resourceKeyNicVLANID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "VLAN ID of the nic",
				ForceNew:    true,
			},
			resourceKeyNicPrivateIPV4: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Private IPV4 address for the nic",
			},
			resourceKeyNicPrivateIPV6: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Private IPV6 Address for the nic",
			},
		},
	}

}

func resourceAdditionalNicCreate(data *schema.ResourceData, provider interface{}) error {
	//propertyHelper := propertyHelper(data)
	apiClient := provider.(*providerState).Client()
	serverID := data.Get(resourceKeyNicServerID).(string)
	ipv4Address := data.Get(resourceKeyNicPrivateIPV4).(string)
	vlanID := data.Get(resourceKeyNicVLANID).(string)

	log.Printf("Configure additional nics for server '%s'...", serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	settings := provider.(*providerState).Settings()

	isStarted := server.Started
	if isStarted {
		if settings.AllowServerReboots {
			err = apiClient.ShutdownServer(serverID)
			if err != nil {
				return err
			}

			_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Shutdown server", serverShutdownTimeout)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Cannot reboot server '%s' because server reboots have been disabled via the 'allow_server_reboot' provider setting or 'DDCLOUD_ALLOW_SERVER_REBOOT' environment variable", serverID)
		}
	}
	log.Printf("create nic in the server id %s", serverID)
	nicID, err := apiClient.AddNicToServer(serverID, ipv4Address, vlanID)

	if err != nil {
		if isStarted {
			err = apiClient.StartServer(serverID)
			if err != nil {
				return err
			}
			_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Start server", serverShutdownTimeout)
			if err != nil {
				return err
			}
		}
		return err
	}

	log.Printf("Adding nic with ID %s to server '%s'...",
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

	data.SetId(nicID) //Nic created
	log.Printf("created the nic with the id %s", nicID)

	if isStarted {
		err = apiClient.StartServer(serverID)
		if err != nil {
			return err
		}
		_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Start server", serverShutdownTimeout)
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

	serverNics := server.Network.AdditionalNetworkAdapters

	nicExists := false
	var serverNic compute.VirtualMachineNetworkAdapter
	for _, nic := range serverNics {
		if *nic.ID == nicID {
			serverNic = nic
			nicExists = true
			break
		}
	}

	if nicExists {
		log.Printf("Nic with the id %s exists", nicID)
	} else {
		log.Printf("Nic with the id %s doesn't exists", nicID)
	}

	if serverNic.ID == nil {
		data.SetId("") // Nic deleted
		return nil
	}
	if err != nil {
		return err
	}
	data.Set(resourceKeyNicPrivateIPV4, serverNic.PrivateIPv4Address)
	data.Set(resourceKeyNicVLANID, serverNic.VLANID)
	data.Set(resourceKeyNicPrivateIPV6, serverNic.PrivateIPv6Address)
	data.Set(resourceKeyNicPrivateIPV4, serverNic.PrivateIPv4Address)

	return nil
}

func resourceAdditionalNicExists(data *schema.ResourceData, provider interface{}) (bool, error) {

	nicExists := false

	serverID := data.Get(resourceKeyNicServerID).(string)

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
	serverNics := server.Network.AdditionalNetworkAdapters
	for _, nic := range serverNics {

		if *nic.ID == nicID {
			nicExists = true
			break
		}
	}
	return nicExists, nil
}

func resourceAdditionalNicRead(data *schema.ResourceData, provider interface{}) error {

	id := data.Id()

	serverID := data.Get(resourceKeyNicServerID).(string)

	apiClient := provider.(*providerState).Client()

	log.Printf("Get the server with the ID %s", serverID)

	server, err := apiClient.GetServer(serverID)

	if server == nil {
		log.Printf("server with the id %s cannot be found", serverID)
	}

	if err != nil {
		return err
	}

	serverNics := server.Network.AdditionalNetworkAdapters

	nicExists := false
	var serverNic compute.VirtualMachineNetworkAdapter
	for _, nic := range serverNics {

		if *nic.ID == id {
			serverNic = nic
			nicExists = true
			break
		}
	}

	if nicExists {
		log.Printf("Nic with the id %s exists", id)
	} else {
		log.Printf("Nic with the id %s doesn't exists", id)
	}

	if serverNic.ID == nil {
		data.SetId("") // Nic deleted
		return nil
	}
	if err != nil {
		return err
	}
	data.Set(resourceKeyNicPrivateIPV4, serverNic.PrivateIPv4Address)
	data.Set(resourceKeyNicVLANID, serverNic.VLANID)
	data.Set(resourceKeyNicPrivateIPV6, serverNic.PrivateIPv6Address)
	data.Set(resourceKeyNicPrivateIPV4, serverNic.PrivateIPv4Address)
	return nil
}

func resourceAdditionalNicUpdate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)
	nicID := data.Id()
	serverID := data.Get(resourceKeyNicServerID).(string)
	privateIPV4 := propertyHelper.GetOptionalString(resourceKeyNicPrivateIPV4, true)

	if data.HasChange(resourceKeyNicPrivateIPV4) {
		log.Printf("changing the ip address of the nic with the id %s to %s", nicID, *privateIPV4)
		apiClient := provider.(*providerState).Client()
		err := updateNicIPAddress(apiClient, serverID, nicID, privateIPV4)
		if err != nil {
			return err
		}
		log.Printf("IP address of the nic with the id %s changed to %s", nicID, *privateIPV4)
	}
	return nil
}

func resourceAdditionalNicDelete(data *schema.ResourceData, provider interface{}) error {
	nicID := data.Id()
	serverID := data.Get(resourceKeyNicServerID).(string)
	apiClient := provider.(*providerState).Client()

	log.Printf("Removing additional nics for server '%s'...", serverID)

	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	settings := provider.(*providerState).Settings()

	isStarted := server.Started
	if isStarted {
		if settings.AllowServerReboots {
			err = apiClient.ShutdownServer(serverID)
			if err != nil {
				return err
			}

			_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Shutdown server", serverShutdownTimeout)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Cannot reboot server '%s' because server reboots have been disabled via the 'allow_server_reboot' provider setting or 'DDCLOUD_ALLOW_SERVER_REBOOT' environment variable", serverID)
		}

	}

	log.Printf("deleting the nic with the id %s", nicID)
	err = apiClient.RemoveNicFromServer(nicID)
	if err != nil {
		if isStarted {
			err = apiClient.StartServer(serverID)
			if err != nil {
				return err
			}
			_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Start server", serverShutdownTimeout)
			if err != nil {
				return err
			}
		}
		return err
	}

	log.Printf("Removing nic with ID %s to server '%s'...",
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

	data.SetId("") //Nic Deleted
	log.Printf("Deleted the nic with the id %s", nicID)

	if isStarted {
		err = apiClient.StartServer(serverID)
		if err != nil {
			return err
		}
		_, err = apiClient.WaitForChange(compute.ResourceTypeServer, serverID, "Start server", serverShutdownTimeout)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateNicIPAddress notifies the compute infrastructure that a Nic's IP address has changed.
func updateNicIPAddress(apiClient *compute.Client, serverID string, nicID string, primaryIPv4 *string) error {
	log.Printf("Update IP address(es) for nic '%s'...", nicID)

	err := apiClient.NotifyServerIPAddressChange(nicID, primaryIPv4, nil)
	if err != nil {
		return err
	}

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", serverID, nicID)
	_, err = apiClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	return err
}
