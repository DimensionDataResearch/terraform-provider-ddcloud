package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	resourceKeyTag      = "tag"
	resourceKeyTagName  = "name"
	resourceKeyTagValue = "value"
)

func schemaTag() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Default:     nil,
		Description: "A set of tags to apply",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyTagName: &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The tag name",
				},
				resourceKeyTagValue: &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The tag value",
				},
			},
		},
		Set: hashTag,
	}
}

// Apply configured tags to a resource.
func applyTags(data *schema.ResourceData, apiClient *compute.Client, assetType string, providerSettings ProviderSettings) error {
	var (
		response *compute.APIResponseV2
		err      error
	)

	resourceID := data.Id()

	log.Printf("Configuring tags for resource '%s'...", resourceID)

	propertyHelper := propertyHelper(data)
	configuredTags := propertyHelper.GetTags(resourceKeyTag)

	tags, err := getTags(apiClient, resourceID, assetType)
	if err != nil {
		return err
	}

	// Capture any tags that are no-longer needed.
	unusedTags := &schema.Set{
		F: schema.HashString,
	}
	for _, tag := range tags {
		unusedTags.Add(tag.Name)
	}
	for _, tag := range configuredTags {
		unusedTags.Remove(tag.Name)
	}

	if len(configuredTags) > 0 {
		log.Printf("Applying %d tags to resource '%s'...", len(configuredTags), resourceID)

		response, err = apiClient.ApplyAssetTags(resourceID, assetType, configuredTags...)
		if err != nil {
			return err
		}

		if response.ResponseCode != compute.ResponseCodeOK {
			return response.ToError("Failed to apply %d tags to resource:'%s' asset_type:'%s' (response code '%s'): %s", len(configuredTags), resourceID, assetType, response.ResponseCode, response.Message)
		}
	} else {
		log.Printf("No tags need to be added to resource '%s'.", resourceID)
	}

	// Trim unused tags (currently-configured tags will overwrite any existing values).
	if unusedTags.Len() > 0 {
		unusedTagNames := make([]string, unusedTags.Len())
		for index, unusedTagName := range unusedTags.List() {
			unusedTagNames[index] = unusedTagName.(string)
		}

		log.Printf("Removing %d unused tags from resource '%s'...", len(unusedTagNames), resourceID)

		response, err = apiClient.RemoveAssetTags(resourceID, assetType, unusedTagNames...)
		if err != nil {
			return err
		}

		if response.ResponseCode != compute.ResponseCodeOK {
			return response.ToError("Failed to remove %d tags from resource '%s' (response code '%s'): %s", len(configuredTags), resourceID, response.ResponseCode, response.Message)
		}
	}

	return nil
}

// Read tags from a resource and update resource data accordingly.
func readTags(data *schema.ResourceData, apiClient *compute.Client, assetType string) error {
	propertyHelper := propertyHelper(data)

	resourceID := data.Id()

	log.Printf("Reading tags for resource '%s'...", resourceID)

	tags, err := getTags(apiClient, resourceID, assetType)
	if err != nil {
		return err
	}

	log.Printf("Read %d tags for resource '%s'.", len(tags), resourceID)

	propertyHelper.SetTags(resourceID, tags)

	return nil
}

func hashTagName(item interface{}) int {
	tagData := item.(map[string]interface{})

	return schema.HashString(
		tagData[resourceKeyTagName].(string),
	)
}

func hashTag(item interface{}) int {
	tagData := item.(map[string]interface{})

	return schema.HashString(fmt.Sprintf(
		"%s=%s",
		tagData[resourceKeyTagName].(string),
		tagData[resourceKeyTagValue].(string),
	))
}

func getTags(apiClient *compute.Client, resourceID string, assetType string) (tags []compute.Tag, err error) {
	page := compute.DefaultPaging()
	page.PageSize = 20

	var tagDetails *compute.TagDetails
	for {
		tagDetails, err = apiClient.GetAssetTags(resourceID, assetType, page)
		if err != nil {
			apiError, ok := err.(*compute.APIError)
			if ok && apiError.Response.GetResponseCode() == compute.ResponseCodeUnexpectedError {
				// This is due to a bug in the CloudControl API (asking for a non-existent page results in UNKNOWN_ERROR).
				err = nil

				break
			}

			return
		}

		if tagDetails.IsEmpty() {
			break
		}

		for _, tagDetail := range tagDetails.Items {
			tags = append(tags,
				tagDetail.ToTag(),
			)
		}

		page.Next()
	}

	return
}
