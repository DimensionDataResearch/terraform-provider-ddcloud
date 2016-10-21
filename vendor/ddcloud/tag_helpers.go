package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

// Create a set containing the names of all tag keys defined in CloudControl for the current user's organisation.
func getDefinedTagKeys(apiClient *compute.Client) (definedTagKeys *schema.Set, err error) {
	definedTagKeys = &schema.Set{F: schema.HashString}

	page := compute.DefaultPaging()
	page.PageSize = 20

	for {
		tagKeys, err := apiClient.ListTagKeys(page)
		if err != nil {
			return nil, err
		}

		if tagKeys.IsEmpty() {
			break
		}

		for _, tagKey := range tagKeys.Items {
			definedTagKeys.Add(tagKey.Name)
		}

		page.Next()
	}

	return definedTagKeys, nil
}

// Given a set of tags, ensure that the corresponding tag keys are defined.
//
// The current user must have permissions to manage tags for their organisation.
func ensureTagKeysAreDefined(apiClient *compute.Client, configuredTags []compute.Tag) error {
	definedTagKeys, err := getDefinedTagKeys(apiClient)
	if err != nil {
		return err
	}

	for _, configuredTag := range configuredTags {
		if !definedTagKeys.Contains(configuredTag.Name) {
			log.Printf("No tag key named '%s' is defined. Since the provider is configured to auto-create missing tag keys, this tag key will now be created.",
				configuredTag.Name,
			)

			var tagKeyID string
			tagKeyID, err = apiClient.CreateTagKey(configuredTag.Name,
				fmt.Sprintf("Tag '%s'", configuredTag.Name),
				false, // isValueRequired
				false, // displayOnReports
			)
			if err != nil {
				return err
			}

			log.Printf("Created new tag key '%s' with ID '%s'.", configuredTag.Name, tagKeyID)
		}
	}

	return nil
}
