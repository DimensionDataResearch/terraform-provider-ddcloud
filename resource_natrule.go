package main

import (
	"compute-api/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

const (
	resourceKeyNATNetworkDomainID = "networkdomain"
	resourceKeyNATPrivateAddress  = "private_ipv4"
	resourceKeyNATPublicAddress   = "public_ipv4"
	resourceCreateTimeoutNAT      = 30 * time.Minute
	resourceUpdateTimeoutNAT      = 10 * time.Minute
	resourceDeleteTimeoutNAT      = 15 * time.Minute
)

const computedPropertyDescription = "<computed>"

func resourceNAT() *schema.Resource {
	return &schema.Resource{
		Create: resourceNATCreate,
		Read:   resourceNATRead,
		Update: resourceNATUpdate,
		Delete: resourceNATDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyNATNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The network domain that the NAT rule applies to.",
			},
			resourceKeyNATPrivateAddress: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The private (internal) IPv4 address.",
			},
			resourceKeyNATPublicAddress: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The public (external) IPv4 address.",
			},
		},
	}
}

// Create a NAT resource.
func resourceNATCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	networkDomainID := data.Get(resourceKeyNATNetworkDomainID).(string)
	privateIP := data.Get(resourceKeyNATPrivateAddress).(string)
	publicIP := propertyHelper.GetOptionalString(resourceKeyNATPublicAddress, false)

	publicIPDescription := computedPropertyDescription
	if publicIP != nil {
		publicIPDescription = *publicIP
	}
	log.Printf("Create NAT rule (from public IP '%s' to private IP '%s') in network domain '%s'.", publicIPDescription, privateIP, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	data.Set(resourceKeyNATPublicAddress, publicIPDescription)

	return nil
}

// Read a NAT resource.
func resourceNATRead(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	id := data.Id()
	networkDomainID := data.Get(resourceKeyNATNetworkDomainID).(string)
	privateIP := data.Get(resourceKeyNATPrivateAddress).(string)
	publicIP := propertyHelper.GetOptionalString(resourceKeyNATPublicAddress, false)

	publicIPDescription := computedPropertyDescription
	if publicIP != nil {
		publicIPDescription = *publicIP
	}

	log.Printf("Read NAT '%s' (private IP = '%s', public IP = '%s') in network domain '%s'.", id, privateIP, publicIPDescription, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}

// Update a NAT resource.
func resourceNATUpdate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	id := data.Id()
	networkDomainID := data.Get(resourceKeyNATNetworkDomainID).(string)
	privateIP := data.Get(resourceKeyNATPrivateAddress).(string)
	publicIP := propertyHelper.GetOptionalString(resourceKeyNATPublicAddress, false)

	publicIPDescription := computedPropertyDescription
	if publicIP != nil {
		publicIPDescription = *publicIP
	}

	log.Printf("Update NAT '%s' (private IP = '%s', public IP = '%s') in network domain '%s'.", id, privateIP, publicIPDescription, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}

// Delete a NAT resource.
func resourceNATDelete(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	id := data.Id()
	networkDomainID := data.Get(resourceKeyNATNetworkDomainID).(string)
	privateIP := data.Get(resourceKeyNATPrivateAddress).(string)
	publicIP := propertyHelper.GetOptionalString(resourceKeyNATPublicAddress, false)

	publicIPDescription := computedPropertyDescription
	if publicIP != nil {
		publicIPDescription = *publicIP
	}

	log.Printf("Delete NAT '%s' (private IP = '%s', public IP = '%s') in network domain '%s'.", id, privateIP, publicIPDescription, networkDomainID)

	providerClient := provider.(*compute.Client)
	providerClient.Reset() // TODO: Replace call to Reset with appropriate API call(s).

	return nil
}
