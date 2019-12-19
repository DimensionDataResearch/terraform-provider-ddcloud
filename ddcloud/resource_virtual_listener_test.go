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

		resource "ddcloud_virtual_listener" "acc_test_listener" {
			name                 	= "%s"
			protocol             	= "HTTP"
			optimization_profile 	= "TCP"
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
				Config: testAccDDCloudVirtualListenerBasic(
					"AccTestListener",
					"192.168.18.10",
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVirtualListenerExists("acc_test_listener", true),
					testCheckDDCloudVirtualListenerMatches("acc_test_listener", compute.VirtualListener{
						Name:                "AccTestListener",
						Protocol:            compute.VirtualListenerStandardProtocolHTTP,
						ListenerIPAddress:   "192.168.18.10",
						Enabled:             true,
						OptimizationProfile: "TCP",
					}),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_virtual_listener (disabling causes in-place update):
//
// Create a virtual listener, then disable it, and verify that it gets updated in-place to Disabled.
func TestAccVirtualListenerBasicUpdateDisable(t *testing.T) {
	testAccResourceUpdateInPlace(t, testAccResourceUpdate{
		ResourceName: "ddcloud_virtual_listener.acc_test_listener",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudVirtualListenerBasic(
			"AccTestListener",
			"192.168.18.10",
			true,
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerExists("acc_test_listener", true),
			testCheckDDCloudVirtualListenerMatches("acc_test_listener", compute.VirtualListener{
				Name:                "AccTestListener",
				Protocol:            compute.VirtualListenerStandardProtocolHTTP,
				ListenerIPAddress:   "192.168.18.10",
				Enabled:             true,
				OptimizationProfile: "TCP",
			}),
		),

		// Update
		UpdateConfig: testAccDDCloudVirtualListenerBasic(
			"AccTestListener",
			"192.168.18.10",
			false,
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerExists("acc_test_listener", true),
			testCheckDDCloudVirtualListenerMatches("acc_test_listener", compute.VirtualListener{
				Name:                "AccTestListener",
				Protocol:            compute.VirtualListenerStandardProtocolHTTP,
				ListenerIPAddress:   "192.168.18.10",
				Enabled:             false,
				OptimizationProfile: "TCP",
			}),
		),
	})
}

// Acceptance test for ddcloud_virtual_listener (changing name causes destroy-and-recreate):
//
// Create a virtual listener, then change its name, and verify that it gets destroyed and recreated with the new name.
func TestAccVirtualListenerBasicUpdateName(t *testing.T) {
	testAccResourceUpdateReplace(t, testAccResourceUpdate{
		ResourceName: "ddcloud_virtual_listener.acc_test_listener",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudVirtualListenerBasic(
			"AccTestListener",
			"192.168.18.10",
			true,
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerExists("acc_test_listener", true),
			testCheckDDCloudVirtualListenerMatches("acc_test_listener", compute.VirtualListener{
				Name:                "AccTestListener",
				Protocol:            compute.VirtualListenerStandardProtocolHTTP,
				ListenerIPAddress:   "192.168.18.10",
				Enabled:             true,
				OptimizationProfile: "TCP",
			}),
		),

		// Update
		UpdateConfig: testAccDDCloudVirtualListenerBasic(
			"AccTestListener1",
			"192.168.18.10",
			true,
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVirtualListenerExists("acc_test_listener", true),
			testCheckDDCloudVirtualListenerMatches("acc_test_listener", compute.VirtualListener{
				Name:                "AccTestListener1",
				Protocol:            compute.VirtualListenerStandardProtocolHTTP,
				ListenerIPAddress:   "192.168.18.10",
				Enabled:             true,
				OptimizationProfile: "TCP",
			}),
		),
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_virtual_listener:
//
// Check if the virtual listener exists.
func testCheckDDCloudVirtualListenerExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_virtual_listener")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		virtualListenerID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		virtualListener, err := client.GetVirtualListener(virtualListenerID)
		if err != nil {
			return fmt.Errorf("bad: Get VirtualListener: %s", err)
		}
		if exists && virtualListener == nil {
			return fmt.Errorf("bad: virtual listener not found with Id '%s'", virtualListenerID)
		} else if !exists && virtualListener != nil {
			return fmt.Errorf("bad: virtual listener still exists with Id '%s'", virtualListenerID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_virtual_listener:
//
// Check if the VirtualListener's configuration matches the expected configuration.
func testCheckDDCloudVirtualListenerMatches(name string, expected compute.VirtualListener) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_virtual_listener")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		virtualListenerID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		virtualListener, err := client.GetVirtualListener(virtualListenerID)
		if err != nil {
			return fmt.Errorf("bad: Get VirtualListener: %s", err)
		}
		if virtualListener == nil {
			return fmt.Errorf("bad: virtual listener not found with Id '%s'", virtualListenerID)
		}

		if virtualListener.Name != expected.Name {
			return fmt.Errorf("bad: virtual listener '%s' has name '%s' (expected '%s')", virtualListenerID, virtualListener.Name, expected.Name)
		}

		if virtualListener.Description != expected.Description {
			return fmt.Errorf("bad: virtual listener '%s' has name '%s' (expected '%s')", virtualListenerID, virtualListener.Description, expected.Description)
		}

		if virtualListener.ListenerIPAddress != expected.ListenerIPAddress {
			return fmt.Errorf("bad: virtual listener '%s' has IPv4 address '%s' (expected '%s')", virtualListenerID, virtualListener.ListenerIPAddress, expected.ListenerIPAddress)
		}

		if virtualListener.Enabled != expected.Enabled {
			return fmt.Errorf("bad: virtual listener '%s' has enablement status '%t' (expected '%t')", virtualListenerID, virtualListener.Enabled, expected.Enabled)
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
