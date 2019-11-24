package ddcloud

import (
	"fmt"
	"testing"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

/*
 * Acceptance-test configurations.
 */

// A basic VLAN (and the network domain that contains it).
func testAccDDCloudVLANBasic(name string, description string) string {

	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"
	   		plan = "ENTERPRISE"
		}

		resource "ddcloud_vlan" "acc_test_vlan" {
			name				= "%s"
			description 		= "%s"
			networkdomain 		= "${ddcloud_networkdomain.acc_test_domain.id}"
			ipv4_base_address	= "192.168.17.0"
			ipv4_prefix_size	= 24
	   		attached_vlan_gateway_addressing = "HIGH"
		}

        resource "ddcloud_vlan" "acc_test_vlan_detached" {
            name                    = "terraform-test-detached-vlan"
            description             = "This is my Terraform test VLAN detached."
            networkdomain           = "${ddcloud_networkdomain.acc_test_domain.id}"
            ipv4_base_address       = "10.1.1.0"
            ipv4_prefix_size        = 28
            
            ## attached_vlan_gateway_addressing and detached_vlan_gateway_address are mutually exclusive
            detached_vlan_gateway_address = "10.0.0.1"
        }
	`, name, description)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_vlan (basic):
//
// Create a VLAN and verify that it gets created with the correct configuration.
func TestAccVLANBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudVLANBasic("acc-test-vlan", "VLAN for Terraform acceptance test."),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVLANExists("ddcloud_vlan.acc_test_vlan", true),
					testCheckDDCloudVLANMatches("ddcloud_vlan.acc_test_vlan", compute.VLAN{
						Name:        "acc-test-vlan",
						Description: "VLAN for Terraform acceptance test.",
						IPv4Range: compute.IPv4Range{
							BaseAddress: "192.168.17.0",
							PrefixSize:  24,
						},
						NetworkDomain: compute.EntityReference{
							Name: "acc-test-networkdomain",
						},
					}),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_vlan (basic):
//
// Update a VLAN and verify that it gets updated with the correct configuration.
func TestAccVLANBasicUpdate(test *testing.T) {
	testAccResourceUpdateInPlace(test, testAccResourceUpdate{
		ResourceName: "ddcloud_vlan.acc_test_vlan",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudVLANBasic(
			"acc-test-vlan",
			"VLAN for Terraform acceptance test.",
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVLANExists("ddcloud_vlan.acc_test_vlan", true),
			testCheckDDCloudVLANMatches("ddcloud_vlan.acc_test_vlan", compute.VLAN{
				Name:        "acc-test-vlan",
				Description: "VLAN for Terraform acceptance test.",
				IPv4Range: compute.IPv4Range{
					BaseAddress: "192.168.17.0",
					PrefixSize:  24,
				},
				NetworkDomain: compute.EntityReference{
					Name: "acc-test-networkdomain",
				},
			}),
		),

		// Update
		UpdateConfig: testAccDDCloudVLANBasic(
			"acc-test-vlan-updated",
			"Updated VLAN for Terraform acceptance test.",
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVLANExists("ddcloud_vlan.acc_test_vlan", true),
			testCheckDDCloudVLANMatches("ddcloud_vlan.acc_test_vlan", compute.VLAN{
				Name:        "acc-test-vlan-updated",
				Description: "Updated VLAN for Terraform acceptance test.",
				IPv4Range: compute.IPv4Range{
					BaseAddress: "192.168.17.0",
					PrefixSize:  24,
				},
				NetworkDomain: compute.EntityReference{
					Name: "acc-test-networkdomain",
				},
			}),
		),
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_vlan:
//
// Check if the VLAN exists.
func testCheckDDCloudVLANExists(name string, exists bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vlanID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vlan, err := client.GetVLAN(vlanID)
		if err != nil {
			return fmt.Errorf("bad: Get VLAN: %s", err)
		}
		if exists && vlan == nil {
			return fmt.Errorf("bad: VLAN not found with Id '%s'", vlanID)
		} else if !exists && vlan != nil {
			return fmt.Errorf("bad: VLAN still exists with Id '%s'", vlanID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_vlan:
//
// Check if the VLAN's configuration matches the expected configuration.
func testCheckDDCloudVLANMatches(name string, expected compute.VLAN) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vlanID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vlan, err := client.GetVLAN(vlanID)
		if err != nil {
			return fmt.Errorf("bad: Get VLAN: %s", err)
		}
		if vlan == nil {
			return fmt.Errorf("bad: VLAN not found with Id '%s'", vlanID)
		}

		if vlan.Name != expected.Name {
			return fmt.Errorf("bad: VLAN '%s' has name '%s' (expected '%s')", vlanID, vlan.Name, expected.Name)
		}

		if vlan.Description != expected.Description {
			return fmt.Errorf("bad: VLAN '%s' has name '%s' (expected '%s')", vlanID, vlan.Description, expected.Description)
		}

		if vlan.IPv4Range.BaseAddress != expected.IPv4Range.BaseAddress {
			return fmt.Errorf("bad: VLAN '%s' has IPv4 base address '%s' (expected '%s')", vlanID, vlan.IPv4Range.BaseAddress, expected.IPv4Range.BaseAddress)
		}

		if vlan.IPv4Range.PrefixSize != expected.IPv4Range.PrefixSize {
			return fmt.Errorf("bad: VLAN '%s' has IPv4 prefix size '%d' (expected '%d')", vlanID, vlan.IPv4Range.PrefixSize, expected.IPv4Range.PrefixSize)
		}

		if vlan.NetworkDomain.Name != expected.NetworkDomain.Name {
			return fmt.Errorf("bad: VLAN '%s' has network domain named '%s' (expected '%s')", vlanID, vlan.NetworkDomain.Name, expected.NetworkDomain.Name)
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_vlan:
//
// Check all VLANs specified in the configuration have been destroyed.
func testCheckDDCloudVLANDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_vlan" {
			continue
		}

		vlanID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vlan, err := client.GetVLAN(vlanID)
		if err != nil {
			return nil
		}
		if vlan != nil {
			return fmt.Errorf("VLAN '%s' still exists", vlanID)
		}
	}

	return nil
}
