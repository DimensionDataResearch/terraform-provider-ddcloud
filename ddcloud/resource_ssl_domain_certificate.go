package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	resourceKeySSLDomainCertificateNetworkDomainID = "networkdomain"
	resourceKeySSLDomainCertificateName            = "name"
	resourceKeySSLDomainCertificateDescription     = "description"
	resourceKeySSLDomainCertificateCertificate     = "certificate"
	resourceKeySSLDomainCertificatePrivateKey      = "private_key"
)

func resourceSSLDomainCertificate() *schema.Resource {
	return &schema.Resource{
		Exists: resourceSSLDomainCertificateExists,
		Create: resourceSSLDomainCertificateCreate,
		Read:   resourceSSLDomainCertificateRead,
		Update: resourceSSLDomainCertificateUpdate,
		Delete: resourceSSLDomainCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSSLDomainCertificateImport,
		},

		Schema: map[string]*schema.Schema{
			resourceKeySSLDomainCertificateNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the SSL domain certificate will be used for SSL offload.",
			},
			resourceKeySSLDomainCertificateName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The name of the SSL domain certificate.",
			},
			resourceKeySSLDomainCertificateDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "A description of the SSL domain certificate.",
			},
			resourceKeySSLDomainCertificateCertificate: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The certificate (in PEM format).",
			},
			resourceKeySSLDomainCertificatePrivateKey: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Sensitive:   true,
				Description: "The certificate's private key (in PEM format).",
			},
		},
	}
}

// Check if a ddcloud_ssl_domain_certificate resource exists.
func resourceSSLDomainCertificateExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()
	log.Printf("Check if SSL domain certificate '%s' exists.", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	SSLDomainCertificate, err := apiClient.GetSSLDomainCertificate(id)
	if err != nil {
		return false, err
	}

	exists := SSLDomainCertificate != nil

	log.Printf("SSL domain certificate '%s' exists: %t.", id, exists)

	return exists, nil
}

// Create a ddcloud_ssl_domain_certificate resource.
func resourceSSLDomainCertificateCreate(data *schema.ResourceData, provider interface{}) error {
	var err error

	networkDomainID := data.Get(resourceKeySSLDomainCertificateNetworkDomainID).(string)
	name := data.Get(resourceKeySSLDomainCertificateName).(string)
	description := data.Get(resourceKeySSLDomainCertificateDescription).(string)
	certificatePEM := data.Get(resourceKeySSLDomainCertificateCertificate).(string)
	privateKeyPEM := data.Get(resourceKeySSLDomainCertificatePrivateKey).(string)

	log.Printf("Create SSL domain certificate '%s' in network domain '%s'.", name, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	var (
		domainCertificateID string
		createError         error
	)

	operationDescription := fmt.Sprintf("Create SSL domain certificate '%s' in network domain '%s'.", name, networkDomainID)
	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		domainCertificateID, createError = apiClient.ImportSSLDomainCertificate(networkDomainID, name, description, certificatePEM, privateKeyPEM)
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

	data.SetId(domainCertificateID)
	log.Printf("Successfully created SSL domain certificate '%s'.", domainCertificateID)

	domainCertificate, err := apiClient.GetSSLDomainCertificate(domainCertificateID)
	if err != nil {
		return err
	}

	if domainCertificate == nil {
		return fmt.Errorf("cannot find newly-added SSL domain certificate '%s'", domainCertificateID)
	}

	return nil
}

// Read a ddcloud_ssl_domain_certificate resource.
func resourceSSLDomainCertificateRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLDomainCertificateNetworkDomainID).(string)
	name := data.Get(resourceKeySSLDomainCertificateName).(string)

	log.Printf("Read SSL domain certificate '%s' ('%s') in network domain '%s'.", name, id, networkDomainID)

	apiClient := provider.(*providerState).Client()

	SSLDomainCertificate, err := apiClient.GetSSLDomainCertificate(id)
	if err != nil {
		return err
	}
	if SSLDomainCertificate == nil {
		data.SetId("") // SSL domain certificate has been deleted

		return nil
	}

	return nil
}

// Update a ddcloud_ssl_domain_certificate resource.
func resourceSSLDomainCertificateUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLDomainCertificateNetworkDomainID).(string)
	name := data.Get(resourceKeySSLDomainCertificateName).(string)

	log.Printf("Update SSL domain certificate '%s' ('%s') in network domain '%s' (nothing to do).", name, id, networkDomainID)

	return nil
}

// Delete a ddcloud_ssl_domain_certificate resource.
func resourceSSLDomainCertificateDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	networkDomainID := data.Get(resourceKeySSLDomainCertificateNetworkDomainID).(string)
	name := data.Get(resourceKeySSLDomainCertificateName).(string)

	log.Printf("Delete SSL domain certificate '%s' ('%s') in network domain '%s' (nothing to do).", name, id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Delete SSL domain certificate '%s", id)

	return providerState.RetryAction(operationDescription, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		err := apiClient.DeleteSSLDomainCertificate(id)
		if err != nil {
			if compute.IsResourceBusyError(err) {
				context.Retry()
			} else {
				context.Fail(err)
			}
		}
	})
}

// Import data for an existing SSL domain certificate.
func resourceSSLDomainCertificateImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	id := data.Id()
	log.Printf("Import SSL domain certificate '%s'.", id)

	var domainCertificate *compute.SSLDomainCertificate
	domainCertificate, err = apiClient.GetSSLDomainCertificate(id)
	if err != nil {
		return
	}
	if domainCertificate == nil {
		err = fmt.Errorf("SSL domain certificate '%s' not found", id)

		return
	}

	data.Set(resourceKeySSLDomainCertificateNetworkDomainID, domainCertificate.NetworkDomainID)
	data.Set(resourceKeySSLDomainCertificateName, domainCertificate.Name)
	data.Set(resourceKeySSLDomainCertificateDescription, domainCertificate.Description)

	importedData = []*schema.ResourceData{data}

	return
}
