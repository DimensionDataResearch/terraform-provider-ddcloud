package main

import (
	"compute-api/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyServerName            = "name"
	resourceKeyServerDescription     = "description"
	resourceKeyServerNetworkDomainID = "networkdomain"
	resourceKeyServerOSImageID       = "osimage_id"
	resourceKeyServerOSImageName     = "osimage_name"
	resourceDeleteTimeoutServer      = 2 * time.Minute
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
			resourceKeyServerNetworkDomainID: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			resourceKeyServerOSImageID: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyServerOSImageName,
				},
			},
			resourceKeyServerOSImageName: &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				ConflictsWith: []string{
					resourceKeyServerOSImageID,
				},
			},
		},
	}
}

// Create a server resource.
func resourceServerCreate(data *schema.ResourceData, provider interface{}) error {
	var name, description, networkDomainID string

	name = data.Get(resourceKeyServerName).(string)
	description = data.Get(resourceKeyServerDescription).(string)
	networkDomainID = data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Create server '%s' in network domain '%s' (description = '%s').", name, networkDomainID, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call.

	return nil
}

// Read a server resource.
func resourceServerRead(data *schema.ResourceData, provider interface{}) error {
	var id, name, description, networkDomainID string

	id = data.Id()
	name = data.Get(resourceKeyServerName).(string)
	description = data.Get(resourceKeyServerDescription).(string)

	log.Printf("Read server '%s' (Id = '%s') in network domain '%s' (description = '%s').", name, id, networkDomainID, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call.

	return nil
}

// Update a server resource.
func resourceServerUpdate(data *schema.ResourceData, provider interface{}) error {
	var id, name, description string

	id = data.Id()

	if data.HasChange(resourceKeyServerName) {
		name = data.Get(resourceKeyServerName).(string)
	}

	if data.HasChange(resourceKeyServerDescription) {
		description = data.Get(resourceKeyServerDescription).(string)
	}

	log.Printf("Update server '%s' (Name = '%s', Description = '%s').", id, name, description)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call.

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
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call.

	return nil
}
