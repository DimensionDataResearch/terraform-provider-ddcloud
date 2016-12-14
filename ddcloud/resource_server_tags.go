package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerTag      = "tag"
	resourceKeyServerTagName  = "name"
	resourceKeyServerTagValue = "value"
)

func schemaServerTag() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Default:     nil,
		Description: "A set of tags to apply to the server",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerTagName: &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The tag name",
				},
				resourceKeyServerTagValue: &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The tag value",
				},
			},
		},
		Set: hashServerTag,
	}
}

// Apply configured tags to a server.
func applyServerTags(data *schema.ResourceData, apiClient *compute.Client, providerSettings ProviderSettings) error {
	var (
		response *compute.APIResponseV2
		err      error
	)

	serverID := data.Id()

	log.Printf("Configuring tags for server '%s'...", serverID)

	propertyHelper := propertyHelper(data)
	configuredTags := propertyHelper.GetTags(resourceKeyServerTag)

	serverTags, err := getServerTags(apiClient, serverID)
	if err != nil {
		return err
	}

	// Capture any tags that are no-longer needed.
	unusedTags := &schema.Set{
		F: schema.HashString,
	}
	for _, tag := range serverTags {
		unusedTags.Add(tag.Name)
	}
	for _, tag := range configuredTags {
		unusedTags.Remove(tag.Name)
	}

	if len(configuredTags) > 0 {
		log.Printf("Applying %d tags to server '%s'...", len(configuredTags), serverID)

		response, err = apiClient.ApplyAssetTags(serverID, compute.AssetTypeServer, configuredTags...)
		if err != nil {
			return err
		}

		if response.ResponseCode != compute.ResponseCodeOK {
			return response.ToError("Failed to apply %d tags to server '%s' (response code '%s'): %s", len(configuredTags), serverID, response.ResponseCode, response.Message)
		}
	} else {
		log.Printf("No tags need to be added to server '%s'.", serverID)
	}

	// Trim unused tags (currently-configured tags will overwrite any existing values).
	if unusedTags.Len() > 0 {
		unusedTagNames := make([]string, unusedTags.Len())
		for index, unusedTagName := range unusedTags.List() {
			unusedTagNames[index] = unusedTagName.(string)
		}

		log.Printf("Removing %d unused tags from server '%s'...", len(unusedTagNames), serverID)

		response, err = apiClient.RemoveAssetTags(serverID, compute.AssetTypeServer, unusedTagNames...)
		if err != nil {
			return err
		}

		if response.ResponseCode != compute.ResponseCodeOK {
			return response.ToError("Failed to remove %d tags from server '%s' (response code '%s'): %s", len(configuredTags), serverID, response.ResponseCode, response.Message)
		}
	}

	return nil
}

// Read tags from a server and update resource data accordingly.
func readServerTags(data *schema.ResourceData, apiClient *compute.Client) error {
	propertyHelper := propertyHelper(data)

	serverID := data.Id()

	log.Printf("Reading tags for server '%s'...", serverID)

	serverTags, err := getServerTags(apiClient, serverID)
	if err != nil {
		return err
	}

	log.Printf("Read %d tags for server '%s'.", len(serverTags), serverID)

	propertyHelper.SetTags(resourceKeyServerTag, serverTags)

	return nil
}

func hashServerTagName(item interface{}) int {
	tagData := item.(map[string]interface{})

	return schema.HashString(
		tagData[resourceKeyServerTagName].(string),
	)
}

func hashServerTag(item interface{}) int {
	tagData := item.(map[string]interface{})

	return schema.HashString(fmt.Sprintf(
		"%s=%s",
		tagData[resourceKeyServerTagName].(string),
		tagData[resourceKeyServerTagValue].(string),
	))
}

func getServerTags(apiClient *compute.Client, serverID string) (serverTags []compute.Tag, err error) {
	page := compute.DefaultPaging()
	page.PageSize = 20

	var tagDetails *compute.TagDetails
	for {
		tagDetails, err = apiClient.GetAssetTags(serverID, compute.AssetTypeServer, page)
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
			serverTags = append(serverTags,
				tagDetail.ToTag(),
			)
		}

		page.Next()
	}

	return
}
