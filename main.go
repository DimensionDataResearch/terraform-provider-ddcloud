package main

import (
	"fmt"
	"github.com/hashicorp/terraform/plugin"
	"os"
	"path"
)

// The main program entry-point.
func main() {
	if len(os.Args) == 1 {
		fmt.Printf("%s %s\n\n", path.Base(os.Args[0]), ProviderVersion)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}
