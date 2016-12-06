package ddcloud

import (
	"fmt"
	"testing"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

/*
 * Acceptance-test configurations.
 */

// Acceptance test configuration - ddcloud_address_list with simple addresses.
func testAccDDCloudAddressListSimple(resourceName string, addressListName string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Address list for Terraform acceptance test."
			datacenter			= "AU9"

			plan				= "ADVANCED"
		}

		resource "ddcloud_address_list" "%s" {
			name					= "%s"
			description 			= "Adam's Terraform test address list (do not delete)."

			ip_version				= "IPv4"

			addresses				= ["192.168.1.10", "192.168.1.20"]

			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}
	`, resourceName, addressListName)
}

// Acceptance test configuration - ddcloud_address_list with addresses and address-ranges.
func testAccDDCloudAddressListComplex(resourceName string, addressListName string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Address list for Terraform acceptance test."
			datacenter			= "AU9"

			plan				= "ADVANCED"
		}

		resource "ddcloud_address_list" "%s" {
			name					= "%s"
			description 			= "Adam's Terraform test address list (do not delete)."

			ip_version				= "IPv4"

			address {
				begin			= "192.168.1.10"
			}
			address {
				begin			= "192.168.1.20"
			}

			address {
				begin			= "192.168.2.10"
				end				= "192.168.2.12"
			}

			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}
	`, resourceName, addressListName)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_address_list:
//
// Create a address list with simple addresses, and verify that it gets created with the correct configuration.
func TestAccAddressListSimpleCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudAddressListDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudAddressListSimple("acc_test_list", "af_terraform_list"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudAddressListExists("acc_test_list", true),
					testCheckDDCloudAddressListMatches("acc_test_list", compute.IPAddressList{
						Name:        "af_terraform_list",
						Description: "Adam's Terraform test address list (do not delete).",
						Addresses: []compute.IPAddressListEntry{
							compute.IPAddressListEntry{
								Begin: "192.168.1.10",
							},
							compute.IPAddressListEntry{
								Begin: "192.168.1.20",
							},
						},
					}),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_address_list:
//
// Create a address list with addresses and address-ranges, and verify that it gets created with the correct configuration.
func TestAccAddressListComplexCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudAddressListDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudAddressListComplex("acc_test_list", "af_terraform_list"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudAddressListExists("acc_test_list", true),
					testCheckDDCloudAddressListMatches("acc_test_list", compute.IPAddressList{
						Name:        "af_terraform_list",
						Description: "Adam's Terraform test address list (do not delete).",
						Addresses: []compute.IPAddressListEntry{
							compute.IPAddressListEntry{
								Begin: "192.168.1.10",
							},
							compute.IPAddressListEntry{
								Begin: "192.168.1.20",
							},
							compute.IPAddressListEntry{
								Begin: "192.168.2.10",
								End:   stringToPtr("192.168.2.12"),
							},
						},
					}),
				),
			},
		},
	})
}

/*
 * Acceptance-test checks
 */

// Acceptance test check for ddcloud_address_list:
//
// Check if the address list exists.
func testCheckDDCloudAddressListExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_address_list")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		addressListID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		addressList, err := client.GetIPAddressList(addressListID)
		if err != nil {
			return fmt.Errorf("Bad: Get address list: %s", err)
		}
		if exists && addressList == nil {
			return fmt.Errorf("Bad: address list not found with Id '%s'.", addressListID)
		} else if !exists && addressList != nil {
			return fmt.Errorf("Bad: address list still exists with Id '%s'.", addressListID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_address_list:
//
// Check if the address list's configuration matches the expected configuration.
func testCheckDDCloudAddressListMatches(name string, expected compute.IPAddressList) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_address_list")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		addressListID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		addressList, err := client.GetIPAddressList(addressListID)
		if err != nil {
			return fmt.Errorf("Bad: Get address list: %s", err)
		}
		if addressList == nil {
			return fmt.Errorf("Bad: address list not found with Id '%s'", addressListID)
		}

		if addressList.Name != expected.Name {
			return fmt.Errorf("Bad: address list '%s' has name '%s' (expected '%s')", addressListID, addressList.Name, expected.Name)
		}

		if addressList.Description != expected.Description {
			return fmt.Errorf("Bad: address list '%s' has description '%s' (expected '%s')", addressListID, addressList.Description, expected.Description)
		}

		if len(addressList.Addresses) != len(expected.Addresses) {
			return fmt.Errorf("Bad: address list '%s' has %d addresses or address-ranges (expected '%d')", addressListID, len(addressList.Addresses), len(expected.Addresses))
		}

		err = compareAddressListEntries(expected, *addressList)
		if err != nil {
			return err
		}

		if len(addressList.ChildLists) != len(expected.ChildLists) {
			return fmt.Errorf("Bad: address list '%s' has %d child lists (expected '%d')", addressListID, len(addressList.ChildLists), len(expected.ChildLists))
		}

		for index := range addressList.ChildLists {
			expectedChildListID := expected.ChildLists[index].ID
			actualChildListID := addressList.ChildLists[index].ID

			if actualChildListID != expectedChildListID {
				return fmt.Errorf("Bad: address list '%s' has child list at index %d with Id %s (expected '%s')",
					addressListID, index, actualChildListID, expectedChildListID,
				)
			}
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_address_list:
//
// Check all address lists specified in the configuration have been destroyed.
func testCheckDDCloudAddressListDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_address_list" {
			continue
		}

		addressListID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		addressList, err := client.GetIPAddressList(addressListID)
		if err != nil {
			return nil
		}
		if addressList != nil {
			return fmt.Errorf("address list '%s' still exists", addressListID)
		}
	}

	return nil
}

func compareAddressListEntries(expectedAddressList compute.IPAddressList, actualAddressList compute.IPAddressList) error {
	addressListID := actualAddressList.ID

	for index := range expectedAddressList.Addresses {
		expectedAddress := expectedAddressList.Addresses[index]
		actualAddress := actualAddressList.Addresses[index]

		if expectedAddress.Begin != actualAddress.Begin {
			return fmt.Errorf("Bad: address list '%s' has entry at index %d with begin address %s (expected '%s')",
				addressListID, index, formatAddress(actualAddress.Begin), formatAddress(expectedAddress.Begin),
			)
		}

		expectedAddressEnd := formatAddress(expectedAddress.End)
		actualAddressEnd := formatAddress(actualAddress.End)
		if expectedAddressEnd != actualAddressEnd {
			return fmt.Errorf("Bad: address list '%s' has entry at index %d with end address %s (expected %s)",
				addressListID, index, actualAddressEnd, expectedAddressEnd,
			)
		}
	}

	return nil
}

func formatAddress(address interface{}) string {
	switch typedAddress := address.(type) {
	case string:
		return typedAddress
	case *string:
		if typedAddress == nil {
			return "nil"
		}

		return *typedAddress
	}

	return "unknown"
}
