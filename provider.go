package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	// TODO: Define schema and resources.

	return &schema.Provider{
		Schema:        map[string]*schema.Schema{},
		ResourcesMap:  map[string]*schema.Resource{},
		ConfigureFunc: configure,
	}
}

func configure(d *schema.ResourceData) (interface{}, error) {
	// TODO: Create provider client.

	return nil, fmt.Errorf("Not implemented yet.")
}
