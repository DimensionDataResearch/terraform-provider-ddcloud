package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// resourcePropertyHelper provides commonly-used functionality for working with Terraform's schema.ResourceData.
type resourcePropertyHelper struct {
	data *schema.ResourceData
}

func propertyHelper(data *schema.ResourceData) resourcePropertyHelper {
	return resourcePropertyHelper{data}
}

func (helper resourcePropertyHelper) GetOptionalString(key string) *string {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case string:
		return &typedValue
	default:
		return nil
	}
}

func (helper resourcePropertyHelper) GetOptionalInt(key string) *int {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case int:
		return &typedValue
	default:
		return nil
	}
}

func (helper resourcePropertyHelper) GetOptionalBool(key string) *bool {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case bool:
		return &typedValue
	default:
		return nil
	}
}
