package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	resourceKeySSLOffloadProfileNetworkDomainID = "networkdomain"
	resourceKeySSLOffloadProfileName            = "name"
	resourceKeySSLOffloadProfileDescription     = "description"
	resourceKeySSLOffloadProfileCiphers         = "ciphers"
	resourceKeySSLOffloadProfileCertificateID   = "certificate"
	resourceKeySSLOffloadProfileChainID         = "chain"
)

func resourceSSLOffloadProfile() *schema.Resource {
	return &schema.Resource{
		Exists: resourceSSLOffloadProfileExists,
		Create: resourceSSLOffloadProfileCreate,
		Read:   resourceSSLOffloadProfileRead,
		Update: resourceSSLOffloadProfileUpdate,
		Delete: resourceSSLOffloadProfileDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSSLOffloadProfileImport,
		},

		Schema: map[string]*schema.Schema{
			resourceKeySSLOffloadProfileNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the SSL-offload profile will be create.",
			},
			resourceKeySSLOffloadProfileName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the SSL-offload profile.",
			},
			resourceKeySSLOffloadProfileDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "A description of the SSL-offload profile.",
			},
			resourceKeySSLOffloadProfileCertificateID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Id of the SSL domain certificate to use.",
			},
			resourceKeySSLOffloadProfileChainID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Id of the SSL certificate chain to use.",
			},
			resourceKeySSLOffloadProfileCiphers: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The SSL ciphers use.",
			},
		},
	}
}

// Check if a ddcloud_ssl_domain_certificate resource exists.
func resourceSSLOffloadProfileExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()
	log.Printf("Check if SSL-offload profile '%s' exists.", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	sslOffloadProfile, err := apiClient.GetSSLOffloadProfile(id)
	if err != nil {
		return false, err
	}

	exists := sslOffloadProfile != nil

	log.Printf("SSL-offload profile '%s' exists: %t.", id, exists)

	return exists, nil
}

// Create a ddcloud_ssl_domain_certificate resource.
func resourceSSLOffloadProfileCreate(data *schema.ResourceData, provider interface{}) error {
	var err error

	propertyHelper := propertyHelper(data)

	networkDomainID := data.Get(resourceKeySSLOffloadProfileNetworkDomainID).(string)
	name := data.Get(resourceKeySSLOffloadProfileName).(string)
	description := data.Get(resourceKeySSLOffloadProfileDescription).(string)
	certificateID := data.Get(resourceKeySSLOffloadProfileCertificateID).(string)
	chainID := data.Get(resourceKeySSLOffloadProfileChainID).(string)
	ciphers := propertyHelper.GetOptionalString(resourceKeySSLOffloadProfileCiphers, false)

	log.Printf("Create SSL-offload profile '%s' in network domain '%s'.", name, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	var (
		sslOffloadProfileID string
		createError         error
	)

	operationDescription := fmt.Sprintf("Create SSL-offload profile '%s' in network domain '%s'.", name, networkDomainID)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		sslOffloadProfileID, createError = apiClient.CreateSSLOffloadProfile(networkDomainID, name, description, ciphers, certificateID, chainID)
		if createError != nil {
			if compute.IsResourceBusyError(createError) {
				context.Retry()
			} else {
				context.Fail(createError)
			}
		}
	})
	if err != nil {
		return err
	}

	data.SetId(sslOffloadProfileID)
	log.Printf("Successfully created SSL-offload profile '%s'.", sslOffloadProfileID)

	sslOffloadProfile, err := apiClient.GetSSLOffloadProfile(sslOffloadProfileID)
	if err != nil {
		return err
	}

	if sslOffloadProfile == nil {
		return fmt.Errorf("cannot find newly-added SSL-offload profile '%s'", sslOffloadProfileID)
	}

	data.Set(resourceKeySSLOffloadProfileCiphers, sslOffloadProfile.Ciphers)

	return nil
}

// Read a ddcloud_ssl_domain_certificate resource.
func resourceSSLOffloadProfileRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLOffloadProfileNetworkDomainID).(string)
	name := data.Get(resourceKeySSLOffloadProfileName).(string)

	log.Printf("Read SSL-offload profile '%s' ('%s') in network domain '%s'.", name, id, networkDomainID)

	apiClient := provider.(*providerState).Client()

	sslOffloadProfile, err := apiClient.GetSSLOffloadProfile(id)
	if err != nil {
		return err
	}
	if sslOffloadProfile == nil {
		data.SetId("") // SSL-offload profile has been deleted

		return nil
	}

	data.Set(resourceKeySSLOffloadProfileCiphers, sslOffloadProfile.Ciphers)

	return nil
}

// Update a ddcloud_ssl_domain_certificate resource.
func resourceSSLOffloadProfileUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLOffloadProfileNetworkDomainID).(string)
	name := data.Get(resourceKeySSLOffloadProfileName).(string)

	log.Printf("Update SSL-offload profile '%s' ('%s') in network domain '%s'.", name, id, networkDomainID)

	hasChange := false

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	sslOffloadProfile, err := apiClient.GetSSLOffloadProfile(id)
	if err != nil {
		return err
	}
	if sslOffloadProfile == nil {
		data.SetId("") // SSL-offload profile has been deleted

		return nil
	}

	if data.HasChange(resourceKeySSLOffloadProfileName) {
		sslOffloadProfile.Name = data.Get(resourceKeySSLOffloadProfileName).(string)

		hasChange = true
	}

	if data.HasChange(resourceKeySSLOffloadProfileDescription) {
		sslOffloadProfile.Description = data.Get(resourceKeySSLOffloadProfileDescription).(string)

		hasChange = true
	}

	if data.HasChange(resourceKeySSLOffloadProfileCertificateID) {
		sslOffloadProfile.SSLDomainCertificate.ID = data.Get(resourceKeySSLOffloadProfileCertificateID).(string)

		hasChange = true
	}

	if data.HasChange(resourceKeySSLOffloadProfileChainID) {
		sslOffloadProfile.SSLCertificateChain.ID = data.Get(resourceKeySSLOffloadProfileChainID).(string)

		hasChange = true
	}

	if data.HasChange(resourceKeySSLOffloadProfileCiphers) {
		sslOffloadProfile.Ciphers = data.Get(resourceKeySSLOffloadProfileCiphers).(string)

		hasChange = true
	}

	if !hasChange {
		return nil
	}

	var editError error

	operationDescription := fmt.Sprintf("Create SSL-offload profile '%s' in network domain '%s'.", name, networkDomainID)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		editError = apiClient.EditSSLOffloadProfile(*sslOffloadProfile)
		if editError != nil {
			if compute.IsResourceBusyError(editError) {
				context.Retry()
			} else {
				context.Fail(editError)
			}
		}
	})
	if err != nil {
		return err
	}

	return nil
}

// Delete a ddcloud_ssl_domain_certificate resource.
func resourceSSLOffloadProfileDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLOffloadProfileNetworkDomainID).(string)
	name := data.Get(resourceKeySSLOffloadProfileName).(string)

	log.Printf("Delete SSL-offload profile '%s' ('%s') in network domain '%s' (nothing to do).", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Delete SSL-offload profile '%s", id)

	return providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		err := apiClient.DeleteSSLOffloadProfile(id)
		if err != nil {
			if compute.IsResourceBusyError(err) {
				context.Retry()
			} else {
				context.Fail(err)
			}
		}
	})
}

// Import data for an existing SSL-offload profile.
func resourceSSLOffloadProfileImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	id := data.Id()
	log.Printf("Import SSL-offload profile '%s'.", id)

	var sslOffloadProfile *compute.SSLOffloadProfile
	sslOffloadProfile, err = apiClient.GetSSLOffloadProfile(id)
	if err != nil {
		return
	}
	if sslOffloadProfile == nil {
		err = fmt.Errorf("SSL-offload profile '%s' not found", id)

		return
	}

	data.Set(resourceKeySSLOffloadProfileNetworkDomainID, sslOffloadProfile.NetworkDomainID)
	data.Set(resourceKeySSLOffloadProfileName, sslOffloadProfile.Name)
	data.Set(resourceKeySSLOffloadProfileDescription, sslOffloadProfile.Description)
	data.Set(resourceKeySSLOffloadProfileCiphers, sslOffloadProfile.Ciphers)
	data.Set(resourceKeySSLOffloadProfileCertificateID, sslOffloadProfile.SSLDomainCertificate.ID)
	data.Set(resourceKeySSLOffloadProfileChainID, sslOffloadProfile.SSLCertificateChain.ID)

	importedData = []*schema.ResourceData{data}

	return
}
