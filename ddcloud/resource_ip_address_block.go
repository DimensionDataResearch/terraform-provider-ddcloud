package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

const (
	resourceKeyIPAddressBlockID          = "ipblock_id"
	resourceKeyIPAddressBlockDomainID    = "domain_id"
	resourceKeyIPAddressBlockBaseIp      = "base_ip"
	resourceKeyIPAddressBlockDescription = "description"
)

func resourceIPAddressBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPAddressBlockCreate,
		Read:   resourceIPAddressBlockRead,
		Update: resourceIPAddressBlockUpdate,
		Delete: resourceIPAddressBlockDelete,
		Importer: &schema.ResourceImporter{
			State: resourceIPAddressBlockImport,
		},

		Schema: map[string]*schema.Schema{
			resourceKeyIPAddressBlockID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Id of the IP Address block.",
			},
			resourceKeyIPAddressBlockDomainID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The network domain where IP address block resides.",
			},
			resourceKeyIPAddressBlockBaseIp: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    false,
				Description: "An IPv4 address in dot notation, corresponding to the lowest IPv address in the block.",
			},
			resourceKeyIPAddressBlockDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the IP Address Block",
			},
			resourceKeyTag: schemaTag(),
		},
	}
}

func resourceIPAddressBlockCreate(data *schema.ResourceData, provider interface{}) (err error) {
	domainID := data.Get(resourceKeyIPAddressBlockDomainID).(string)
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	log.Printf("Creating new IP address block in network domain '%s'",
		domainID,
	)

	ipBlockID, err := apiClient.AddPublicIPBlock(domainID)
	data.SetId(ipBlockID)

	if err != nil {
		return err
	}

	// Tags
	err = applyTags(data, apiClient, compute.AssetTypePublicIPBlock, providerState.Settings())
	if err != nil {
		return err
	}
	data.SetPartial(resourceKeyTag)
	data.Partial(false)

	return nil
}

func resourceIPAddressBlockRead(data *schema.ResourceData, provider interface{}) (err error) {

	id := data.Id()
	domainID := data.Get(resourceKeyIPAddressBlockDomainID).(string)
	ipBlockID := data.Get(resourceKeyIPAddressBlockID).(string)
	baseIp := data.Get(resourceKeyIPAddressBlockBaseIp).(string)

	log.Printf("Read IP Address Block Id = '%s' in network domain '%s', IP Base Address '%s').", ipBlockID, domainID, baseIp)

	apiClient := provider.(*providerState).Client()

	ipBlock, err := apiClient.GetPublicIPBlock(id)
	if err != nil {
		return err
	}

	if ipBlock != nil {
		data.Set(resourceKeyIPAddressBlockDomainID, ipBlock.ID)
		data.Set(resourceKeyIPAddressBlockDomainID, ipBlock.NetworkDomainID)
		data.Set(resourceKeyIPAddressBlockBaseIp, ipBlock.BaseIP)

		err = readTags(data, apiClient, compute.AssetTypePublicIPBlock)
		if err != nil {
			return err
		}

	} else {
		data.SetId("") // Mark resource as deleted.
	}

	return nil
}

func resourceIPAddressBlockUpdate(data *schema.ResourceData, provider interface{}) (err error) {

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	if data.HasChange(resourceKeyTag) {
		err := applyTags(data, apiClient, compute.AssetTypePublicIPBlock, providerState.Settings())
		if err != nil {
			return err
		}

		data.SetPartial(resourceKeyTag)
	}

	return
}

func resourceIPAddressBlockDelete(data *schema.ResourceData, provider interface{}) (err error) {

	id := data.Id()
	networkDomainID := data.Get(resourceKeyIPAddressBlockDomainID).(string)

	log.Printf("Delete IP Address Block '%s' in network domain '%s'.", id, networkDomainID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Delete IP address block '%s'", id)
	err = providerState.Retry().Action(operationDescription, deleteTimeoutVLAN, func(context retry.Context) {
		// CloudControl has issues if more than one asynchronous operation is initated at a time (returns UNEXPECTED_ERROR).
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released once the current attempt is complete.

		deleteError := apiClient.RemovePublicIPBlock(id)
		if deleteError != nil {
			if compute.IsResourceBusyError(deleteError) {
				context.Retry()
			} else {
				context.Fail(deleteError)
			}
		}

		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	log.Printf("IP Address Block '%s' is being deleted...", id)

	return apiClient.WaitForDelete(compute.ResourceTypePublicIPBlock, id, 2)
}

func resourceIPAddressBlockImport(data *schema.ResourceData, provider interface{}) (importedData []*schema.ResourceData, err error) {
	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	ipAddressBlockID := data.Id()
	log.Printf("Import IP Address Block by ID '%s'.", ipAddressBlockID)
	ipAddressBlock, err := apiClient.GetPublicIPBlock(ipAddressBlockID)

	if err != nil {
		return
	}

	if ipAddressBlock == nil {
		err = fmt.Errorf("IP Address block '%s' not found", ipAddressBlockID)

		return
	}

	data.SetId(ipAddressBlock.ID)

	data.Set(resourceKeyIPAddressBlockID, ipAddressBlock.ID)
	data.Set(resourceKeyIPAddressBlockDomainID, ipAddressBlock.NetworkDomainID)
	data.Set(resourceKeyIPAddressBlockBaseIp, ipAddressBlock.BaseIP)

	err = readTags(data, apiClient, compute.AssetTypePublicIPBlock)

	importedData = []*schema.ResourceData{data}

	return
}
