package main

import (
	"compute-api/compute"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyServerName                 = "name"
	resourceKeyServerDescription          = "description"
	resourceKeyServerAdminPassword        = "admin_password"
	resourceKeyServerNetworkDomainID      = "networkdomain"
	resourceKeyServerMemoryGB             = "memory_gb"
	resourceKeyServerCPUCount             = "cpu_count"
	resourceKeyServerAdditionalDisk       = "additional_disk"
	resourceKeyServerAdditionalDiskID     = "disk_id"
	resourceKeyServerAdditionalDiskSizeGB = "size_gb"
	resourceKeyServerAdditionalDiskUnitID = "scsi_unit_id"
	resourceKeyServerAdditionalDiskSpeed  = "speed"
	resourceKeyServerOSImageID            = "osimage_id"
	resourceKeyServerOSImageName          = "osimage_name"
	resourceKeyServerPrimaryVLAN          = "primary_adapter_vlan"
	resourceKeyServerPrimaryIPv4          = "primary_adapter_ipv4"
	resourceKeyServerPrimaryIPv6          = "primary_adapter_ipv6"
	resourceKeyServerPrimaryDNS           = "dns_primary"
	resourceKeyServerSecondaryDNS         = "dns_secondary"
	resourceKeyServerAutoStart            = "auto_start"
	resourceCreateTimeoutServer           = 30 * time.Minute
	resourceUpdateTimeoutServer           = 10 * time.Minute
	resourceDeleteTimeoutServer           = 15 * time.Minute
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

				// TODO: Mark this property as sensitive when we upgrade to a version of Terraform that supports it.
			},
			resourceKeyServerMemoryGB: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Default:  nil,
			},
			resourceKeyServerCPUCount: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Default:  nil,
			},
			resourceKeyServerAdditionalDisk: &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						resourceKeyServerAdditionalDiskID: &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						resourceKeyServerAdditionalDiskSizeGB: &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						resourceKeyServerAdditionalDiskUnitID: &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						resourceKeyServerAdditionalDiskSpeed: &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "STANDARD",
						},
					},
				},
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
				Computed: true,
				Default:  nil,
			},
			resourceKeyServerPrimaryIPv4: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
				Default:  nil,
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
				Computed: true,
				Default:  nil,
			},
			resourceKeyServerOSImageName: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
				Default:  nil,
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
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)
	primaryDNS := data.Get(resourceKeyServerPrimaryDNS).(string)
	secondaryDNS := data.Get(resourceKeyServerSecondaryDNS).(string)
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

	propertyHelper := propertyHelper(data)

	// Retrieve image details.
	osImageID := propertyHelper.GetOptionalString(resourceKeyServerOSImageID, false)
	osImageName := propertyHelper.GetOptionalString(resourceKeyServerOSImageName, false)

	var osImage *compute.OSImage
	if osImageID != nil {
		// TODO: Look up OS image by Id (first, implement in compute API client).

		return fmt.Errorf("Specifying osimage_id is not supported yet.")
	} else if osImageName != nil {
		log.Printf("Looking up OS image '%s' by name...", *osImageName)

		osImage, err = providerClient.FindOSImage(*osImageName, dataCenterID)
		if err != nil {
			return err
		}

		if osImage == nil {
			log.Printf("Warning - unable to find an OS image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", *osImageName, dataCenterID, networkDomainID)

			return fmt.Errorf("Unable to find an OS image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", *osImageName, dataCenterID, networkDomainID)
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

	// Memory and CPU
	memoryGB := propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	if memoryGB != nil {
		deploymentConfiguration.MemoryGB = *memoryGB
	} else {
		data.Set(resourceKeyServerMemoryGB, deploymentConfiguration.MemoryGB)
	}

	cpuCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	if cpuCount != nil {
		deploymentConfiguration.CPU.Count = *cpuCount
	} else {
		data.Set(resourceKeyServerCPUCount, deploymentConfiguration.CPU.Count)
	}

	// Network
	primaryVLANID := propertyHelper.GetOptionalString(resourceKeyServerPrimaryVLAN, false)
	primaryIPv4Address := propertyHelper.GetOptionalString(resourceKeyServerPrimaryIPv4, false)

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
	serverIPv6Address := *server.Network.PrimaryAdapter.PrivateIPv6Address
	data.Set(resourceKeyServerPrimaryIPv6, serverIPv6Address)

	disks := server.Disks
	diskProperties := make([]map[string]interface{}, len(disks))
	for index, disk := range disks {
		diskProperties[index] = map[string]interface{}{
			resourceKeyServerAdditionalDiskID:     *disk.ID,
			resourceKeyServerAdditionalDiskSizeGB: disk.SizeGB,
			resourceKeyServerAdditionalDiskUnitID: disk.SCSIUnitID,
			resourceKeyServerAdditionalDiskSpeed:  disk.Speed,
		}
	}
	data.Set(resourceKeyServerAdditionalDisk, diskProperties)

	// TODO: Configure additional disks for the server, if specified in the configuration.
	additionalDisks := propertyHelper.GetVirtualMachineDisks(resourceKeyServerAdditionalDisk)
	if len(additionalDisks) > 0 {
		log.Printf("We'll need to configure %d additional disks now the server is deployed: '%#v'", len(additionalDisks), additionalDisks)
	}

	deploymentConfiguration.Disks = mergeDisks(deploymentConfiguration.Disks, additionalDisks)

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
	data.Set(resourceKeyServerOSImageID, server.SourceImageID)
	data.Set(resourceKeyServerMemoryGB, server.MemoryGB)
	data.Set(resourceKeyServerCPUCount, server.CPU.Count)
	data.Set(resourceKeyServerPrimaryVLAN, *server.Network.PrimaryAdapter.VLANID)
	data.Set(resourceKeyServerPrimaryIPv4, *server.Network.PrimaryAdapter.PrivateIPv4Address)
	data.Set(resourceKeyServerPrimaryIPv6, *server.Network.PrimaryAdapter.PrivateIPv6Address)
	data.Set(resourceKeyServerNetworkDomainID, server.Network.NetworkDomainID)

	disks := server.Disks
	diskProperties := make([]map[string]interface{}, len(disks))
	for index, disk := range disks {
		diskProperties[index] = buildDiskProperties(disk)
	}
	data.Set(resourceKeyServerAdditionalDisk, diskProperties)

	return nil
}

// Update a server resource.
func resourceServerUpdate(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Id()

	// These changes can only be made through the V1 API (we're mostly using V2).
	// Later, we can come back and implement the required functionality in the compute API client.
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

	data.Partial(true)

	propertyHelper := propertyHelper(data)

	var memoryGB, cpuCount *int
	if data.HasChange(resourceKeyServerMemoryGB) {
		memoryGB = propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	}
	if data.HasChange(resourceKeyServerCPUCount) {
		cpuCount = propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	}

	if memoryGB != nil || cpuCount != nil {
		log.Printf("Server CPU / memory configuration change detected.")

		err = updateServerConfiguration(providerClient, server, memoryGB, cpuCount)
		if err != nil {
			return err
		}

		if data.HasChange(resourceKeyServerMemoryGB) {
			data.SetPartial(resourceKeyServerMemoryGB)
		}

		if data.HasChange(resourceKeyServerCPUCount) {
			data.SetPartial(resourceKeyServerCPUCount)
		}
	}

	var primaryIPv4, primaryIPv6 *string
	if data.HasChange(resourceKeyServerPrimaryIPv4) {
		primaryIPv4 = propertyHelper.GetOptionalString(resourceKeyServerPrimaryIPv4, false)
	}
	if data.HasChange(resourceKeyServerPrimaryIPv6) {
		primaryIPv6 = propertyHelper.GetOptionalString(resourceKeyServerPrimaryIPv6, false)
	}

	if primaryIPv4 != nil || primaryIPv6 != nil {
		log.Printf("Server network configuration change detected.")

		err = updateServerIPAddress(providerClient, server, primaryIPv4, primaryIPv6)
		if err != nil {
			return err
		}

		if data.HasChange(resourceKeyServerPrimaryIPv4) {
			data.SetPartial(resourceKeyServerPrimaryIPv4)
		}

		if data.HasChange(resourceKeyServerPrimaryIPv6) {
			data.SetPartial(resourceKeyServerPrimaryIPv6)
		}
	}

	data.Partial(false)

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

// updateServerConfiguration reconfigures a server, changing the allocated RAM and / or CPU count.
func updateServerConfiguration(providerClient *compute.Client, server *compute.Server, memoryGB *int, cpuCount *int) error {
	memoryDescription := "no change"
	if memoryGB != nil {
		memoryDescription = fmt.Sprintf("will change to %dGB", *memoryGB)
	}

	cpuCountDescription := "no change"
	if memoryGB != nil {
		memoryDescription = fmt.Sprintf("will change to %d", *cpuCount)
	}

	log.Printf("Update configuration for server '%s' (memory: %s, CPU: %s)...", server.ID, memoryDescription, cpuCountDescription)

	err := providerClient.ReconfigureServer(server.ID, memoryGB, cpuCount)
	if err != nil {
		return err
	}

	_, err = providerClient.WaitForChange(compute.ResourceTypeServer, server.ID, "Reconfigure server", resourceUpdateTimeoutServer)

	return err
}

// updateServerIPAddress notifies the compute infrastructure that a server's IP address has changed.
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

// Parse and append additional disks to those specified by the image being deployed.
func mergeDisks(imageDisks []compute.VirtualMachineDisk, additionalDisks []compute.VirtualMachineDisk) []compute.VirtualMachineDisk {
	diskSet := schema.NewSet(
		func(item interface{}) int {
			disk := item.(compute.VirtualMachineDisk)

			return disk.SCSIUnitID
		},
		[]interface{}{},
	)

	for _, disk := range imageDisks {
		log.Printf("Merge image disk with SCSI unit Id %d.", disk.SCSIUnitID)
		diskSet.Add(disk)
	}

	for _, disk := range additionalDisks {
		log.Printf("Merge additional disk with SCSI unit Id %d.", disk.SCSIUnitID)
		diskSet.Add(disk)
	}

	mergedDisks := make([]compute.VirtualMachineDisk, diskSet.Len())
	for index, disk := range diskSet.List() {
		mergedDisks[index] = disk.(compute.VirtualMachineDisk)
	}

	return mergedDisks
}

func buildDiskProperties(disk compute.VirtualMachineDisk) map[string]interface{} {
	diskProperties := map[string]interface{}{
		resourceKeyServerAdditionalDiskSizeGB: disk.SizeGB,
		resourceKeyServerAdditionalDiskUnitID: disk.SCSIUnitID,
		resourceKeyServerAdditionalDiskSpeed:  disk.Speed,
	}

	if disk.ID != nil {
		diskProperties[resourceKeyServerAdditionalDiskID] = *disk.ID
	}

	return diskProperties
}
