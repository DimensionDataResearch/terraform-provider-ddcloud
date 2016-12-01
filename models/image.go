package models

import (
	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/maps"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
)

// Image represents the Terraform configuration for a ddcloud_server image.
type Image struct {
	ID   string
	Name string
	Type string
}

// ReadMap populates the Image with values from the specified map.
func (image *Image) ReadMap(imageProperties map[string]interface{}) {
	reader := maps.NewReader(imageProperties)

	image.ID = reader.GetString("id")
	image.Name = reader.GetString("name")
	image.Type = reader.GetString("type")
}

// ReadImage populates the Image with values from the specified compute.Image.
func (image *Image) ReadImage(computeImage compute.Image) {
	image.ID = computeImage.GetID()
	image.Name = computeImage.GetName()
}

// ToMap creates a new map using the values from the Image.
func (image *Image) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	image.UpdateMap(data)

	return data
}

// UpdateMap updates a map using values from the Image.
func (image *Image) UpdateMap(imageProperties map[string]interface{}) {
	writer := maps.NewWriter(imageProperties)

	writer.SetString("id", image.ID)
	writer.SetString("name", image.ID)
	writer.SetString("type", image.Type)
}

// NewImageFromMap creates a Image from the values in the specified map.
func NewImageFromMap(imageProperties map[string]interface{}) Image {
	image := Image{}
	image.ReadMap(imageProperties)

	return image
}
