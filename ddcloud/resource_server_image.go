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

func resolveServerImage(imageNameOrID string, imageType string, datacenterID string, apiClient *compute.Client) (resolvedImage compute.Image, err error) {
	log.Printf("Resolve server image '%s' (%s) in datacenter '%s'.",
		imageNameOrID,
		imageType,
		datacenterID,
	)

	isID, err := isUUID(imageNameOrID)
	if err != nil {
		return
	}

	if isID {
		log.Printf("'%s' appears to be a UUID; treating it as an image Id.", imageNameOrID)
	} else {
		log.Printf("'%s' does not appear to be a UUID; treating it as an image name.", imageNameOrID)
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

				log.Printf("Resolved OS image with Id '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			err = fmt.Errorf("cannot find an OS image with Id '%s' in datacenter '%s'",
				imageNameOrID,
				datacenterID,
			)
		} else {
			osImage, err = lookupOSImageByName(imageNameOrID, datacenterID, apiClient)
			if err != nil {
				return
			}
			if osImage != nil {
				resolvedImage = osImage

				log.Printf("Resolved OS image named '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			err = fmt.Errorf("cannot find an OS image named '%s' in datacenter '%s'",
				imageNameOrID,
				datacenterID,
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

				log.Printf("Resolved customer image with Id '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			err = fmt.Errorf("cannot find a customer image with Id '%s' in datacenter '%s'",
				imageNameOrID,
				datacenterID,
			)
		} else {
			customerImage, err = lookupCustomerImageByName(imageNameOrID, datacenterID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				log.Printf("Resolved customer image named '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			err = fmt.Errorf("cannot find a customer image named '%s' in datacenter '%s'",
				imageNameOrID,
				datacenterID,
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

				log.Printf("Resolved OS image with Id '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			// Fall back to customer image, if required.
			customerImage, err = lookupCustomerImageByID(imageNameOrID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				log.Printf("Resolved customer image with Id '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			err = fmt.Errorf("cannot find an OS or customer image with Id '%s' in datacenter '%s'",
				imageNameOrID,
				datacenterID,
			)
		} else {
			osImage, err = lookupOSImageByName(imageNameOrID, datacenterID, apiClient)
			if err != nil {
				return
			}
			if osImage != nil {
				resolvedImage = osImage

				log.Printf("Resolved OS image named '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			// Fall back to customer image, if required.
			customerImage, err = lookupCustomerImageByName(imageNameOrID, datacenterID, apiClient)
			if err != nil {
				return
			}
			if customerImage != nil {
				resolvedImage = customerImage

				log.Printf("Resolved customer image named '%s' in datacenter '%s'.", imageNameOrID, datacenterID)

				return
			}

			err = fmt.Errorf("cannot find an OS or customer image named '%s' in datacenter '%s'",
				imageNameOrID,
				datacenterID,
			)
		}
	default:
		err = fmt.Errorf("invalid image type '%s'", imageType)

		return
	}

	return
}

func lookupOSImageByID(imageID string, apiClient *compute.Client) (*compute.OSImage, error) {
	log.Printf("Looking up OS image '%s' by Id...", imageID)

	return apiClient.GetOSImage(imageID)
}

func lookupOSImageByName(imageName string, datacenterID string, apiClient *compute.Client) (*compute.OSImage, error) {
	log.Printf("Looking up OS image '%s' by name in datacenter '%s'...", imageName, datacenterID)

	return apiClient.FindOSImage(imageName, datacenterID)
}

func lookupCustomerImageByID(imageID string, apiClient *compute.Client) (*compute.CustomerImage, error) {
	log.Printf("Looking up customer image '%s' by Id...", imageID)

	return apiClient.GetCustomerImage(imageID)
}

func lookupCustomerImageByName(imageName string, datacenterID string, apiClient *compute.Client) (*compute.CustomerImage, error) {
	log.Printf("Looking up customer image '%s' by name in datacenter '%s'...", imageName, datacenterID)

	return apiClient.FindCustomerImage(imageName, datacenterID)
}
