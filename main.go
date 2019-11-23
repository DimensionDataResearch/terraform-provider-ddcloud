package main

import (
	"fmt"
	"os"
	"path"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/ddcloud"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// The main program entry-point.
func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf("%s %s\n\n", path.Base(os.Args[0]), ddcloud.ProviderVersion)

		return
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ddcloud.Provider,
	})
}
