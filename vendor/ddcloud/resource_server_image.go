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

	var (
		osImage       *compute.OSImage
		customerImage *compute.CustomerImage
	)
	switch imageType {
	case serverImageTypeOS:
		if isID {
			osImage, err = lookupOSImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}
			if osImage != nil {
				resolvedImage = osImage

				return
			}

			err = fmt.Errorf("Cannot find an OS image with Id '%s' in datacenter '%s'",
				imageNameOrID,
				dataCenterID,
			)
		} else {
			osImage, err = lookupOSImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}
			if osImage != nil {
				resolvedImage = osImage

				return
			}

			err = fmt.Errorf("Cannot find an OS image named '%s' in datacenter '%s'",
				imageNameOrID,
				dataCenterID,
			)
		}
	case serverImageTypeCustomer:
		if isID {
			customerImage, err = lookupCustomerImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				return
			}

			err = fmt.Errorf("Cannot find a customer image with Id '%s' in datacenter '%s'",
				imageNameOrID,
				dataCenterID,
			)
		} else {
			customerImage, err = lookupCustomerImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				return
			}

			err = fmt.Errorf("Cannot find a customer image named '%s' in datacenter '%s'",
				imageNameOrID,
				dataCenterID,
			)
		}
	case serverImageTypeAuto:
		if isID {
			osImage, err = lookupOSImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}
			if osImage != nil {
				resolvedImage = osImage

				return
			}

			// Fall back to customer image, if required.
			customerImage, err = lookupCustomerImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				return
			}

			err = fmt.Errorf("Cannot find an OS or customer image with Id '%s' in datacenter '%s'",
				imageNameOrID,
				dataCenterID,
			)
		} else {
			osImage, err = lookupOSImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}
			if osImage != nil {
				resolvedImage = osImage

				return
			}

			// Fall back to customer image, if required.
			customerImage, err = lookupCustomerImageByName(imageNameOrID, dataCenterID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				return
			}

			err = fmt.Errorf("Cannot find an OS or customer image named '%s' in datacenter '%s'",
				imageNameOrID,
				dataCenterID,
			)
		}
	default:
		err = fmt.Errorf("Invalid image type '%s'", imageType)

		return
	}

	return
}

func lookupOSImageByID(imageID string, apiClient *compute.Client) (*compute.OSImage, error) {
	log.Printf("Looking up OS image '%s' by Id...", imageID)

	return apiClient.GetOSImage(imageID)
}

func lookupOSImageByName(imageName string, dataCenterID string, apiClient *compute.Client) (*compute.OSImage, error) {
	log.Printf("Looking up OS image '%s' by name in datacenter '%s'...", imageName, dataCenterID)

	return apiClient.FindOSImage(imageName, dataCenterID)
}

func lookupCustomerImageByID(imageID string, apiClient *compute.Client) (*compute.CustomerImage, error) {
	log.Printf("Looking up customer image '%s' by Id...", imageID)

	return apiClient.GetCustomerImage(imageID)
}

func lookupCustomerImageByName(imageName string, dataCenterID string, apiClient *compute.Client) (*compute.CustomerImage, error) {
	log.Printf("Looking up customer image '%s' by name in datacenter '%s'...", imageName, dataCenterID)

	return apiClient.FindCustomerImage(imageName, dataCenterID)
}
