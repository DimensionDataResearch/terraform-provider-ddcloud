package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

/*
 * Acceptance-test configurations.
 */

func testAccDDCloudNetworkDomainDSBasic(name string, description string, datacenterID string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "%s"
			description	= "%s"
			datacenter	= "%s"
		}

		data "ddcloud_networkdomain" "acc_ds_test_domain" {
			name		= "${ddcloud_networkdomain.acc_test_domain.name}"
			datacenter	= "${ddcloud_networkdomain.acc_test_domain.datacenter}"
		}`,
		name, description, datacenterID,
	)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_networkdomain data-source (basic):
//
// Create a network domain and data-source referencing it, and verify that it gets created with the correct configuration.
func TestAccNetworkDomainDSBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testCheckDDCloudNetworkDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudNetworkDomainDSBasic(
					"acc-test-domain",
					"Network domain for Terraform acceptance test.",
					"AU9",
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudNetworkDomainExists("acc_test_domain", true),
					testCheckDDCloudNetworkDomainMatches("acc_test_domain", compute.NetworkDomain{
						Name:         "acc-test-domain",
						Description:  "Network domain for Terraform acceptance test.",
						DatacenterID: "AU9",
					}),
					testCheckDDCloudNetworkDomainMatchesDataSource("acc_test_domain", "acc_ds_test_domain"),
				),
			},
		},
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_networkdomain:
//
// Check if a network domain resource's configuration matches a network domain data-source's configuration.
func testCheckDDCloudNetworkDomainMatchesDataSource(resourceName string, dataSourceName string) resource.TestCheckFunc {
	resourceName = ensureResourceTypePrefix(resourceName, "ddcloud_networkdomain")
	dataSourceName = ensureDataSourceTypePrefix(dataSourceName, "ddcloud_networkdomain")

	return func(state *terraform.State) error {
		for key, value := range state.RootModule().Resources {
			fmt.Printf("Resource: '%s' (%s/%s)\n", key, value.Provider, value.Type)
		}

		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		ds, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Data-source not found: %s", dataSourceName)
		}

		if res.Primary.ID != ds.Primary.ID {
			return fmt.Errorf("Bad: Resource '%s' has Id '%s', but data-source '%s' has Id '%s' (expected the IDs to match)",
				resourceName,
				res.Primary.ID,
				dataSourceName,
				ds.Primary.ID,
			)
		}

		resAttributes := res.Primary.Attributes
		dsAttributes := ds.Primary.Attributes

		if resAttributes[resourceKeyNetworkDomainDescription] != dsAttributes[resourceKeyNetworkDomainDescription] {
			return fmt.Errorf("Bad: Resource '%s' has description '%s', but data-source '%s' has description '%s' (expected the plans to match)",
				resourceName,
				resAttributes[resourceKeyNetworkDomainDescription],
				dataSourceName,
				dsAttributes[resourceKeyNetworkDomainDescription],
			)
		}

		if resAttributes[resourceKeyNetworkDomainPlan] != dsAttributes[resourceKeyNetworkDomainPlan] {
			return fmt.Errorf("Bad: Resource '%s' has plan '%s', but data-source '%s' has plan '%s' (expected the plans to match)",
				resourceName,
				resAttributes[resourceKeyNetworkDomainPlan],
				dataSourceName,
				dsAttributes[resourceKeyNetworkDomainPlan],
			)
		}

		if res.Primary.Attributes[resourceKeyNetworkDomainNatIPv4Address] != ds.Primary.Attributes[resourceKeyNetworkDomainNatIPv4Address] {
			return fmt.Errorf("Bad: Resource '%s' has S/NAT address '%s', but data-source '%s' has S/NAT address '%s' (expected the addresses to match)",
				resourceName,
				resAttributes[resourceKeyNetworkDomainNatIPv4Address],
				dataSourceName,
				dsAttributes[resourceKeyNetworkDomainNatIPv4Address],
			)
		}

		return nil
	}
}
