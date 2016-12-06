package ddcloud

import (
	"fmt"
	"log"
	"regexp"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

const (
	serverImageTypeOS       = "os"
	serverImageTypeCustomer = "customer"
	serverImageTypeAuto     = "auto"
)

func isUUID(str string) (bool, error) {
	return regexp.Match(`[A-Fa-f0-9]{8}(-[A-Fa-f0-9]{4}){3}-[A-Fa-f0-9]{12}`, []byte(str))
}

func resolveServerImage(imageNameOrID string, imageType string, dataCenterID string, apiClient *compute.Client) (resolvedImage compute.Image, err error) {
	isID, err := isUUID(imageNameOrID)
	if err != nil {
		return
	}

	switch imageType {
	case serverImageTypeOS:
		if isID {
			resolvedImage, err = lookupOSImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS image with Id '%s' in datacenter '%s'",
					imageNameOrID,
					dataCenterID,
				)
			}
		} else {
			resolvedImage, err = lookupOSImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS image named '%s' in datacenter '%s'",
					imageNameOrID,
					dataCenterID,
				)
			}
		}
	case serverImageTypeCustomer:
		if isID {
			resolvedImage, err = lookupCustomerImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find a customer image with Id '%s' in datacenter '%s'",
					imageNameOrID,
					dataCenterID,
				)
			}
		} else {
			resolvedImage, err = lookupCustomerImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find a customer image named '%s' in datacenter '%s'",
					imageNameOrID,
					dataCenterID,
				)
			}
		}
	case serverImageTypeAuto:
		if isID {
			resolvedImage, err = lookupOSImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}

			// Fall back to customer image, if required.
			if resolvedImage == nil {
				resolvedImage, err = lookupCustomerImageByID(imageNameOrID, apiClient)
				if err != nil {
					return
				}
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS or customer image with Id '%s' in datacenter '%s'",
					imageNameOrID,
					dataCenterID,
				)
			}
		} else {
			resolvedImage, err = lookupOSImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}

			// Fall back to customer image, if required.
			if resolvedImage == nil {
				resolvedImage, err = lookupCustomerImageByName(imageNameOrID, dataCenterID, apiClient)
				if err != nil {
					return
				}
			}

			if resolvedImage == nil {
				err = fmt.Errorf("Cannot find an OS or customer image named '%s' in datacenter '%s'",
					imageNameOrID,
					dataCenterID,
				)
			}
		}
	default:
		err = fmt.Errorf("Invalid image type '%s'", imageType)

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

	return apiClient.FindCustomerImage(imageName, dataCenterID)
}
