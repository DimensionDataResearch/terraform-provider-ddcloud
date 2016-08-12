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

// A basic virtual listener (and the network domain that contains it).
func testAccDDCloudVirtualListenerBasic(name string, listenerIPAddress string, enabled bool) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"

			plan		= "ADVANCED"
		}

		resource "ddcloud_virtual_listener" "acc_test_virtual_listener" {
			name                 	= "%s"
			protocol             	= "HTTP"
			optimization_profiles 	= ["TCP"]
			ipv4                	= "%s"
			enabled                 = "%t"

			networkdomain 		 	= "${ddcloud_networkdomain.acc_test_domain.id}"
		}
	`, name, listenerIPAddress, enabled)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_virtual_listener (basic):
//
// Create a virtual listener and verify that it gets created with the correct configuration.
func TestAccVirtualListenerBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudVirtualListenerBasic("acc-test-virtual-listener", "192.168.18.10", true),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVirtualListenerExists("ddcloud_virtual_listener.acc_test_virtual_listener", true),
					testCheckDDCloudVirtualListenerMatches("ddcloud_virtual_listener.acc_test_virtual_listener", compute.VirtualListener{
						Name:              "acc-test-virtual-listener",
						Protocol:          compute.VirtualListenerStandardProtocolHTTP,
						ListenerIPAddress: "192.168.18.10",
						Enabled:           true,
						OptimizationProfiles: []string{
							"TCP",
						},
					}),
				),
			},
		},
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_virtual_listener:
//
// Check if the virtual listener exists.
func testCheckDDCloudVirtualListenerExists(name string, exists bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		virtualListenerID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		virtualListener, err := client.GetVirtualListener(virtualListenerID)
		if err != nil {
			return fmt.Errorf("Bad: Get VirtualListener: %s", err)
		}
		if exists && virtualListener == nil {
			return fmt.Errorf("Bad: virtual listener not found with Id '%s'.", virtualListenerID)
		} else if !exists && virtualListener != nil {
			return fmt.Errorf("Bad: virtual listener still exists with Id '%s'.", virtualListenerID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_virtual_listener:
//
// Check if the VirtualListener's configuration matches the expected configuration.
func testCheckDDCloudVirtualListenerMatches(name string, expected compute.VirtualListener) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		virtualListenerID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		virtualListener, err := client.GetVirtualListener(virtualListenerID)
		if err != nil {
			return fmt.Errorf("Bad: Get VirtualListener: %s", err)
		}
		if virtualListener == nil {
			return fmt.Errorf("Bad: virtual listener not found with Id '%s'.", virtualListenerID)
		}

		if virtualListener.Name != expected.Name {
			return fmt.Errorf("Bad: virtual listener '%s' has name '%s' (expected '%s').", virtualListenerID, virtualListener.Name, expected.Name)
		}

		if virtualListener.Description != expected.Description {
			return fmt.Errorf("Bad: virtual listener '%s' has name '%s' (expected '%s').", virtualListenerID, virtualListener.Description, expected.Description)
		}

		if virtualListener.ListenerIPAddress != expected.ListenerIPAddress {
			return fmt.Errorf("Bad: virtual listener '%s' has IPv4 address '%s' (expected '%s').", virtualListenerID, virtualListener.ListenerIPAddress, expected.ListenerIPAddress)
		}

		if virtualListener.Enabled != expected.Enabled {
			return fmt.Errorf("Bad: virtual listener '%s' has enablement status '%t' (expected '%t').", virtualListenerID, virtualListener.Enabled, expected.Enabled)
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_virtual_listener:
//
// Check all VirtualListeners specified in the configuration have been destroyed.
func testCheckDDCloudVirtualListenerDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_virtual_listener" {
			continue
		}

		virtualListenerID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		virtualListener, err := client.GetVirtualListener(virtualListenerID)
		if err != nil {
			return nil
		}
		if virtualListener != nil {
			return fmt.Errorf("Virtual listener '%s' still exists", virtualListenerID)
		}
	}

	return nil
}
