package ddcloud

import (
	"fmt"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

func addPublicIPBlock(networkDomainID string, apiClient *compute.Client) (*compute.PublicIPBlock, error) {
	blockID, err := apiClient.AddPublicIPBlock(networkDomainID)
	if err != nil {
		return nil, err
	}

	publicIPBlock, err := apiClient.GetPublicIPBlock(blockID)
	if err != nil {
		return nil, err
	}
	if publicIPBlock == nil {
		return nil, fmt.Errorf("cannot find newly-added public IPv4 address block '%s'", blockID)
	}

	return publicIPBlock, nil
}
