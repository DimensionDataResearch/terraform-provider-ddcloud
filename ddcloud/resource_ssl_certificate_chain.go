package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeySSLCertificateChainNetworkDomainID = "networkdomain"
	resourceKeySSLCertificateChainName            = "name"
	resourceKeySSLCertificateChainDescription     = "description"
	resourceKeySSLCertificateChainChain           = "chain"
)

func resourceSSLCertificateChain() *schema.Resource {
	return &schema.Resource{
		Exists: resourceSSLCertificateChainExists,
		Create: resourceSSLCertificateChainCreate,
		Read:   resourceSSLCertificateChainRead,
		Update: resourceSSLCertificateChainUpdate,
		Delete: resourceSSLCertificateChainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSSLCertificateChainImport,
		},

		Schema: map[string]*schema.Schema{
			resourceKeySSLCertificateChainNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the SSL certificate chain will be used for SSL offload.",
			},
			resourceKeySSLCertificateChainName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The name of the SSL certificate chain.",
			},
			resourceKeySSLCertificateChainDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "A description of the SSL certificate chain.",
			},
			resourceKeySSLCertificateChainChain: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The certificate chain (in PEM format).",
			},
		},
	}
}

// Check if a ddcloud_ssl_domain_certificate resource exists.
func resourceSSLCertificateChainExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()
	log.Printf("Check if SSL certificate chain '%s' exists.", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	SSLCertificateChain, err := apiClient.GetSSLCertificateChain(id)
	if err != nil {
		return false, err
	}

	exists := SSLCertificateChain != nil

	log.Printf("SSL certificate chain '%s' exists: %t.", id, exists)

	return exists, nil
}

// Create a ddcloud_ssl_domain_certificate resource.
func resourceSSLCertificateChainCreate(data *schema.ResourceData, provider interface{}) error {
	var err error

	networkDomainID := data.Get(resourceKeySSLCertificateChainNetworkDomainID).(string)
	name := data.Get(resourceKeySSLCertificateChainName).(string)
	description := data.Get(resourceKeySSLCertificateChainDescription).(string)
	chainPEM := data.Get(resourceKeySSLCertificateChainChain).(string)

	log.Printf("Create SSL certificate chain '%s' in network domain '%s'.", name, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	var (
		certificateChainID string
		createError        error
	)

	operationDescription := fmt.Sprintf("Create SSL certificate chain '%s' in network domain '%s'.", name, networkDomainID)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		certificateChainID, createError = apiClient.ImportSSLCertificateChain(networkDomainID, name, description, chainPEM)
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

	data.SetId(certificateChainID)
	log.Printf("Successfully created SSL certificate chain '%s'.", certificateChainID)

	certificateChain, err := apiClient.GetSSLCertificateChain(certificateChainID)
	if err != nil {
		return err
	}

	if certificateChain == nil {
		return fmt.Errorf("cannot find newly-added SSL certificate chain '%s'", certificateChainID)
	}

	return nil
}

// Read a ddcloud_ssl_domain_certificate resource.
func resourceSSLCertificateChainRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLCertificateChainNetworkDomainID).(string)
	name := data.Get(resourceKeySSLCertificateChainName).(string)

	log.Printf("Read SSL certificate chain '%s' ('%s') in network domain '%s'.", name, id, networkDomainID)

	apiClient := provider.(*providerState).Client()

	certificateChain, err := apiClient.GetSSLCertificateChain(id)
	if err != nil {
		return err
	}
	if certificateChain == nil {
		data.SetId("") // SSL certificate chain has been deleted

		return nil
	}

	return nil
}

// Update a ddcloud_ssl_domain_certificate resource.
func resourceSSLCertificateChainUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLCertificateChainNetworkDomainID).(string)
	name := data.Get(resourceKeySSLCertificateChainName).(string)

	log.Printf("Update SSL certificate chain '%s' ('%s') in network domain '%s' (nothing to do).", name, id, networkDomainID)

	return nil
}

// Delete a ddcloud_ssl_domain_certificate resource.
func resourceSSLCertificateChainDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLCertificateChainNetworkDomainID).(string)
	name := data.Get(resourceKeySSLCertificateChainName).(string)

	log.Printf("Delete SSL certificate chain '%s' ('%s') in network domain '%s' (nothing to do).", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Delete SSL certificate chain '%s", id)

	return providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		err := apiClient.DeleteSSLCertificateChain(id)
		if err != nil {
			if compute.IsResourceBusyError(err) {
				context.Retry()
			} else {
				context.Fail(err)
			}
		}
	})
}

// Import data for an existing SSL certificate chain.
func resourceSSLCertificateChainImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	id := data.Id()
	log.Printf("Import SSL certificate chain '%s'.", id)

	var certificateChain *compute.SSLCertificateChain
	certificateChain, err = apiClient.GetSSLCertificateChain(id)
	if err != nil {
		return
	}
	if certificateChain == nil {
		err = fmt.Errorf("SSL certificate chain '%s' not found", id)

		return
	}

	data.Set(resourceKeySSLCertificateChainNetworkDomainID, certificateChain.NetworkDomainID)
	data.Set(resourceKeySSLCertificateChainName, certificateChain.Name)
	data.Set(resourceKeySSLCertificateChainDescription, certificateChain.Description)

	importedData = []*schema.ResourceData{data}

	return
}
