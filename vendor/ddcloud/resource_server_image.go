package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/models"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerImage     = "image"
	resourceKeyServerImageID   = "id"
	resourceKeyServerImageName = "name"
	resourceKeyServerImageType = "type"

	serverImageTypeOS       = "os"
	serverImageTypeCustomer = "customer"
	serverImageTypeAuto     = "auto"
)

func schemaServerImage() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList, // Unfortunate limitation of Terraform schema - have to model this as a list with exactly 1 item.
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				resourceKeyServerImageID: &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Default:     nil,
					Description: "The Id of the image from which the server is created",
				},
				resourceKeyServerImageName: &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Default:     nil,
					Description: "The Id of the image from which the server is created",
				},
				resourceKeyServerImageType: &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Default:     serverImageTypeAuto,
					Description: "The type of image from which the server is created (default is auto-detect)",
					ValidateFunc: func(value interface{}, key string) (warnings []string, errors []error) {
						imageType := value.(string)

						switch imageType {
						case serverImageTypeOS:
						case serverImageTypeCustomer:
						case serverImageTypeAuto:
							return
						default:
							errors = append(errors,
								fmt.Errorf("Invalid image type '%s'", imageType),
							)
						}

						return
					},
				},
			},
		},
	}
}

func resolveServerImage(configuredImage *models.Image, dataCenterID string, apiClient *compute.Client) (resolvedImage compute.Image, err error) {
	switch configuredImage.Type {
	case serverImageTypeOS:
		if configuredImage.ID != "" {
			resolvedImage, err = lookupOSImageByID(configuredImage.ID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS image with Id '%s' in datacenter '%s'",
					configuredImage.ID,
					dataCenterID,
				)
			}
		} else if configuredImage.Name != "" {
			resolvedImage, err = lookupOSImageByName(configuredImage.Name, dataCenterID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS image named '%s' in datacenter '%s'",
					configuredImage.Name,
					dataCenterID,
				)
			}
		}
	case serverImageTypeCustomer:
		if configuredImage.ID != "" {
			resolvedImage, err = lookupCustomerImageByID(configuredImage.ID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find a customer image with Id '%s' in datacenter '%s'",
					configuredImage.ID,
					dataCenterID,
				)
			}
		} else if configuredImage.Name != "" {
			resolvedImage, err = lookupCustomerImageByName(configuredImage.Name, dataCenterID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find a customer image named '%s' in datacenter '%s'",
					configuredImage.Name,
					dataCenterID,
				)
			}
		}
	case serverImageTypeAuto:
		if configuredImage.ID != "" {
			resolvedImage, err = lookupOSImageByID(configuredImage.ID, apiClient)
			if err != nil {
				return
			}

			// Fall back to customer image, if required.
			if resolvedImage == nil {
				resolvedImage, err = lookupCustomerImageByID(configuredImage.ID, apiClient)
				if err != nil {
					return
				}
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS or customer image with Id '%s' in datacenter '%s'",
					configuredImage.ID,
					dataCenterID,
				)
			}
		} else if configuredImage.Name != "" {
			resolvedImage, err = lookupOSImageByName(configuredImage.Name, dataCenterID, apiClient)
			if err != nil {
				return
			}

			// Fall back to customer image, if required.
			if resolvedImage == nil {
				resolvedImage, err = lookupCustomerImageByName(configuredImage.Name, dataCenterID, apiClient)
				if err != nil {
					return
				}
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS or customer image named '%s' in datacenter '%s'",
					configuredImage.Name,
					dataCenterID,
				)
			}
		}
	default:
		err = fmt.Errorf("Invalid image type '%s'", configuredImage.Type)

		return
	}

	return
}

func lookupOSImageByID(imageID string, apiClient *compute.Client) (compute.Image, error) {
	log.Printf("Looking up OS image '%s' by Id...", imageID)

	return apiClient.GetOSImage(imageID)
}

func lookupOSImageByName(imageName string, dataCenterID string, apiClient *compute.Client) (compute.Image, error) {
	log.Printf("Looking up OS image '%s' by name in datacenter '%s'...", imageName, dataCenterID)

	return apiClient.FindOSImage(imageName, dataCenterID)
}

func lookupCustomerImageByID(imageID string, apiClient *compute.Client) (compute.Image, error) {
	log.Printf("Looking up customer image '%s' by Id...", imageID)

	return apiClient.GetCustomerImage(imageID)
}

func lookupCustomerImageByName(imageName string, dataCenterID string, apiClient *compute.Client) (compute.Image, error) {
	log.Printf("Looking up customer image '%s' by name in datacenter '%s'...", imageName, dataCenterID)

	return apiClient.FindOSImage(imageName, dataCenterID)
}
