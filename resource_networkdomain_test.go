package main

import (
	"compute-api/compute"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccNetworkDomainCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testCheckDDComputeNetworkDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudNetworkDomainBasic,
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudNetworkDomainExists("ddcloud_networkdomain.acc_test_domain"),
				),
			},
		},
	})
}

func testCheckDDCloudNetworkDomainExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		networkDomainID := res.Primary.ID

		client := testAccProvider.Meta().(*compute.Client)
		networkDomain, err := client.GetNetworkDomain(networkDomainID)
		if err != nil {
			return fmt.Errorf("Bad: Get network domain: %s", err)
		}
		if networkDomain == nil {
			return fmt.Errorf("Bad: Network domain not found with Id '%s'.", networkDomainID)
		}

		return nil
	}
}

func testCheckDDComputeNetworkDomainDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_networkdomain" {
			continue
		}

		networkDomainID := res.Primary.ID

		client := testAccProvider.Meta().(*compute.Client)
		networkDomain, err := client.GetNetworkDomain(networkDomainID)
		if err != nil {
			return nil
		}
		if networkDomain != nil {
			return fmt.Errorf("Network domain '%s' still exists.", networkDomainID)
		}
	}

	return nil
}

const testAccDDCloudNetworkDomainBasic = `
resource "ddcloud_networkdomain" "acc_test_domain" {
	name		= "acc-test-domain"
	description	= "Network domain for Terraform acceptance test."
	datacenter	= "AU9"
}
`
