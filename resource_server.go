package main

import (
	"compute-api/compute"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyServerName            = "name"
	resourceKeyServerDescription     = "description"
	resourceKeyServerAdminPassword   = "admin_password"
	resourceKeyServerNetworkDomainID = "networkdomain"
	resourceKeyServerMemoryGB        = "memory_gb"
	resourceKeyServerCPUCount        = "cpu_count"
	resourceKeyServerOSImageID       = "osimage_id"
	resourceKeyServerOSImageName     = "osimage_name"
	resourceKeyServerPrimaryVLAN     = "primary_adapter_vlan"
	resourceKeyServerPrimaryIPv4     = "primary_adapter_ipv4"
	resourceKeyServerPrimaryIPv6     = "primary_adapter_ipv6"
	resourceKeyServerPrimaryDNS      = "dns_primary"
	resourceKeyServerSecondaryDNS    = "dns_secondary"
	resourceKeyServerAutoStart       = "auto_start"
	resourceCreateTimeoutServer      = 30 * time.Minute
	resourceUpdateTimeoutServer      = 10 * time.Minute
	resourceDeleteTimeoutServer      = 15 * time.Minute
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyServerName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			resourceKeyServerDescription: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerAdminPassword: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			resourceKeyServerMemoryGB: &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Computed: true,
				Default:  nil,
			},
			resourceKeyServerCPUCount: &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Computed: true,
				Default:  nil,
			},
			resourceKeyServerNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyServerPrimaryVLAN: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerPrimaryIPv4: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerPrimaryIPv6: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			resourceKeyServerPrimaryDNS: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerSecondaryDNS: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerOSImageID: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerOSImageName: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			resourceKeyServerAutoStart: &schema.Schema{
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Default:  false,
			},
		},
	}
}

// Create a server resource.
func resourceServerCreate(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	adminPassword := data.Get(resourceKeyServerAdminPassword).(string)
	memoryGB := data.Get(resourceKeyServerMemoryGB).(*int)
	cpuCount := data.Get(resourceKeyServerCPUCount).(*int)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)
	primaryVLAN := data.Get(resourceKeyServerPrimaryVLAN).(string)
	primaryIPv4 := data.Get(resourceKeyServerPrimaryIPv4).(string)
	primaryDNS := data.Get(resourceKeyServerPrimaryDNS).(string)
	secondaryDNS := data.Get(resourceKeyServerSecondaryDNS).(string)
	osImageID := data.Get(resourceKeyServerOSImageID).(string)
	osImageName := data.Get(resourceKeyServerOSImageName).(string)
	autoStart := data.Get(resourceKeyServerAutoStart).(bool)

	log.Printf("Create server '%s' in network domain '%s' (description = '%s').", name, networkDomainID, description)

	providerClient := provider.(*compute.Client)

	networkDomain, err := providerClient.GetNetworkDomain(networkDomainID)
	if err != nil {
		return err
	}

	if networkDomain == nil {
		return fmt.Errorf("No network domain was found with Id '%s'.", networkDomainID)
	}

	dataCenterID := networkDomain.DatacenterID
	log.Printf("Server will be deployed in data centre '%s'.", dataCenterID)

	// Retrieve image details.
	var osImage *compute.OSImage
	if len(osImageID) > 0 {
		// TODO: Look up OS image by Id (first, implement in compute API client).

		return fmt.Errorf("Specifying osimage_id is not supported yet.")
	} else if len(osImageName) > 0 {
		log.Printf("Looking up OS image '%s' by name...", osImageName)

		osImage, err = providerClient.FindOSImage(osImageName, dataCenterID)
		if err != nil {
			return err
		}

		if osImage == nil {
			log.Printf("Warning - unable to find an OS image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", osImageName, dataCenterID, networkDomainID)

			return fmt.Errorf("Unable to find an OS image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", osImageName, dataCenterID, networkDomainID)
		}

		log.Printf("Server will be deployed from OS image with Id '%s'.", osImage.ID)
		data.Set(resourceKeyServerOSImageID, osImage.ID)
	} else {
		return fmt.Errorf("Must specify either osimage_id or osimage_name.")
	}

	deploymentConfiguration := compute.ServerDeploymentConfiguration{
		Name:                  name,
		Description:           description,
		AdministratorPassword: adminPassword,
		Start: autoStart,
	}
	err = deploymentConfiguration.ApplyImage(osImage)
	if err != nil {
		return err
	}

	// Memory
	if memoryGB != nil {
		deploymentConfiguration.MemoryGB = *memoryGB
	} else {
		data.Set(resourceKeyServerMemoryGB, *memoryGB)
	}

	// CPU
	if cpuCount != nil {
		deploymentConfiguration.CPU.Count = *cpuCount
	} else {
		data.Set(resourceKeyServerCPUCount, *cpuCount)
	}

	// Network
	var (
		primaryVLANID      *string
		primaryIPv4Address *string
	)
	if len(primaryVLAN) > 0 {
		primaryVLANID = &primaryVLAN
	}
	if len(primaryIPv4) > 0 {
		primaryIPv4Address = &primaryIPv4
	}
	deploymentConfiguration.Network = compute.VirtualMachineNetwork{
		NetworkDomainID: networkDomainID,
		PrimaryAdapter: compute.VirtualMachineNetworkAdapter{
			VLANID:             primaryVLANID,
			PrivateIPv4Address: primaryIPv4Address,
		},
	}
	deploymentConfiguration.PrimaryDNS = primaryDNS
	deploymentConfiguration.SecondaryDNS = secondaryDNS

	log.Printf("Server deployment configuration: %+v", deploymentConfiguration)
	log.Printf("Server CPU deployment configuration: %+v", deploymentConfiguration.CPU)

	serverID, err := providerClient.DeployServer(deploymentConfiguration)
	if err != nil {
		return err
	}

	data.SetId(serverID)

	log.Printf("Server '%s' is being provisioned...", name)

	resource, err := providerClient.WaitForDeploy(compute.ResourceTypeServer, serverID, resourceCreateTimeoutServer)
	if err != nil {
		return err
	}

	// Capture additional properties that are only available after deployment.
	server := resource.(*compute.Server)
	data.Set(resourceKeyServerPrimaryIPv6, server.Network.PrimaryAdapter.PrivateIPv6Address)

	// Configure connection info so that we can use a provisioner if required.
	switch server.OperatingSystem.Family {
	case "UNIX":
		data.SetConnInfo(map[string]string{
			"type":     "ssh",
			"host":     *server.Network.PrimaryAdapter.PrivateIPv6Address,
			"user":     "root",
			"password": adminPassword,
		})

	case "WINDOWS":
		data.SetConnInfo(map[string]string{
			"type":     "winrm",
			"host":     *server.Network.PrimaryAdapter.PrivateIPv6Address,
			"user":     "Administrator",
			"password": adminPassword,
		})
	}

	return nil
}

// Read a server resource.
func resourceServerRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Read server '%s' (Id = '%s') in network domain '%s' (description = '%s').", name, id, networkDomainID, description)

	providerClient := provider.(*compute.Client)
	server, err := providerClient.GetServer(id)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("Server '%s' has been deleted.", id)

		// Mark as deleted.
		data.SetId("")

		return nil
	}

	data.Set(resourceKeyServerName, server.Name)
	data.Set(resourceKeyServerDescription, server.Description)
	data.Set(resourceKeyServerMemoryGB, server.MemoryGB)
	data.Set(resourceKeyServerCPUCount, server.CPU.Count)
	data.Set(resourceKeyServerPrimaryVLAN, server.Network.PrimaryAdapter.VLANID)
	data.Set(resourceKeyServerPrimaryIPv4, server.Network.PrimaryAdapter.PrivateIPv4Address)
	data.Set(resourceKeyServerPrimaryIPv6, server.Network.PrimaryAdapter.PrivateIPv6Address)
	data.Set(resourceKeyServerNetworkDomainID, server.Network.NetworkDomainID)

	return nil
}

// Update a server resource.
func resourceServerUpdate(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Id()

	if data.HasChange(resourceKeyServerName) {
		return fmt.Errorf("Changing the 'name' property of a 'ddcloud_server' resource type is not yet implemented.")
	}

	if data.HasChange(resourceKeyServerDescription) {
		return fmt.Errorf("Changing the 'description' property of a 'ddcloud_server' resource type is not yet implemented.")
	}

	log.Printf("Update server '%s'.", serverID)

	providerClient := provider.(*compute.Client)
	server, err := providerClient.GetServer(serverID)
	if err != nil {
		return err
	}

	var memoryGB, cpuCount *int
	if data.HasChange(resourceKeyServerMemoryGB) {
		memoryGB = data.Get(resourceKeyServerMemoryGB).(*int)
	}
	if data.HasChange(resourceKeyServerCPUCount) {
		cpuCount = data.Get(resourceKeyServerCPUCount).(*int)
	}

	if memoryGB != nil || cpuCount != nil {
		err = updateServerConfiguration(providerClient, server, memoryGB, cpuCount)
		if err != nil {
			return err
		}
	}

	var primaryIPv4, primaryIPv6 *string
	if data.HasChange(resourceKeyServerPrimaryIPv4) {
		primaryIPv4 = data.Get(resourceKeyServerPrimaryIPv4).(*string)
	}
	if data.HasChange(resourceKeyServerPrimaryIPv6) {
		primaryIPv4 = data.Get(resourceKeyServerPrimaryIPv6).(*string)
	}

	if primaryIPv4 != nil || primaryIPv6 != nil {
		err = updateServerIPAddress(providerClient, server, primaryIPv4, primaryIPv6)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete a server resource.
func resourceServerDelete(data *schema.ResourceData, provider interface{}) error {
	var id, name, networkDomainID string

	id = data.Id()
	name = data.Get(resourceKeyServerName).(string)
	networkDomainID = data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Delete server '%s' ('%s') in network domain '%s'.", id, name, networkDomainID)

	providerClient := provider.(*compute.Client)
	err := providerClient.DeleteServer(id)
	if err != nil {
		return err
	}

	log.Printf("Server '%s' is being deleted...", id)

	return providerClient.WaitForDelete(compute.ResourceTypeServer, id, resourceDeleteTimeoutServer)
}

func updateServerIPAddress(providerClient *compute.Client, server *compute.Server, primaryIPv4 *string, primaryIPv6 *string) error {
	log.Printf("Update primary IP address(es) for server '%s'...", server.ID)

	primaryNetworkAdapterID := *server.Network.PrimaryAdapter.ID
	err := providerClient.NotifyServerIPAddressChange(primaryNetworkAdapterID, primaryIPv4, primaryIPv6)
	if err != nil {
		return err
	}

	compositeNetworkAdapterID := fmt.Sprintf("%s/%s", server.ID, primaryNetworkAdapterID)
	_, err = providerClient.WaitForChange(compute.ResourceTypeNetworkAdapter, compositeNetworkAdapterID, "Update adapter IP address", resourceUpdateTimeoutServer)

	return err
}

func updateServerConfiguration(providerClient *compute.Client, server *compute.Server, memoryGB *int, cpuCount *int) error {
	log.Printf("Update configuration for server '%s'...", server.ID)

	return fmt.Errorf("Server reconfiguration (for 'ddcloud_server' resource type) is not yet implemented.")
}
