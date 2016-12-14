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

// Acceptance test configuration - ddcloud_vip_node with status
func testAccDDCloudVIPNodeBasic(resourceName string, nodeName string, status string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "VIP node for Terraform acceptance test."
			datacenter			= "AU9"

			plan				= "ADVANCED"
		}

		resource "ddcloud_vip_node" "%s" {
			name					= "%s"
			description 			= "Adam's Terraform test VIP node (do not delete)."
			ipv4_address			= "192.168.17.10"
			status					= "%s"
      health_monitor  = "CCDEFAULT.Icmp"
			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}
	`, resourceName, nodeName, status)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_vip_node:
//
// Create a VIP node and verify that it gets created with the correct configuration.
func TestAccVIPNodeBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPNodeDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudVIPNodeBasic("acc_test_node", "af_terraform_node", compute.VIPNodeStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVIPNodeExists("acc_test_node", true),
					testCheckDDCloudVIPNodeMatches("acc_test_node", compute.VIPNode{
						Name:        "af_terraform_node",
						IPv4Address: "192.168.17.10",
						Status:      compute.VIPNodeStatusEnabled,
						HealthMonitor: compute.VIPNodeHealthMonitor{
							Name: "CCDEFAULT.Icmp",
						},
					}),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_vip_node (changing status causes in-place update):
//
// Create a VIP node, then change its status, and verify that it gets updated in-place with the correct status.
func TestAccVIPNodeBasicUpdateStatus(t *testing.T) {
	testAccResourceUpdateInPlace(t, testAccResourceUpdate{
		ResourceName: "ddcloud_vip_node.acc_test_node",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPNodeDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudVIPNodeBasic(
			"acc_test_node",
			"af_terraform_node",
			compute.VIPNodeStatusEnabled,
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPNodeExists("acc_test_node", true),
			testCheckDDCloudVIPNodeMatches("acc_test_node", compute.VIPNode{
				Name:        "af_terraform_node",
				IPv4Address: "192.168.17.10",
				Status:      compute.VIPNodeStatusEnabled,
				HealthMonitor: compute.VIPNodeHealthMonitor{
					Name: "CCDEFAULT.Icmp",
				},
			}),
		),

		// Update
		UpdateConfig: testAccDDCloudVIPNodeBasic(
			"acc_test_node",
			"af_terraform_node",
			compute.VIPNodeStatusDisabled,
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPNodeExists("acc_test_node", true),
			testCheckDDCloudVIPNodeMatches("acc_test_node", compute.VIPNode{
				Name:        "af_terraform_node",
				IPv4Address: "192.168.17.10",
				Status:      compute.VIPNodeStatusDisabled,
				HealthMonitor: compute.VIPNodeHealthMonitor{
					Name: "CCDEFAULT.Icmp",
				},
			}),
		),
	})
}

/*
 * Acceptance-test checks
 */

// Acceptance test check for ddcloud_vip_node:
//
// Check if the vip node exists.
func testCheckDDCloudVIPNodeExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_vip_node")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vipNodeID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipNode, err := client.GetVIPNode(vipNodeID)
		if err != nil {
			return fmt.Errorf("Bad: Get vip node: %s", err)
		}
		if exists && vipNode == nil {
			return fmt.Errorf("Bad: VIP node not found with Id '%s'.", vipNodeID)
		} else if !exists && vipNode != nil {
			return fmt.Errorf("Bad: VIP node still exists with Id '%s'.", vipNodeID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_vip_node:
//
// Check if the vip node's configuration matches the expected configuration.
func testCheckDDCloudVIPNodeMatches(name string, expected compute.VIPNode) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_vip_node")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vipNodeID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipNode, err := client.GetVIPNode(vipNodeID)
		if err != nil {
			return fmt.Errorf("Bad: Get vip node: %s", err)
		}
		if vipNode == nil {
			return fmt.Errorf("Bad: VIP node not found with Id '%s'", vipNodeID)
		}

		if vipNode.Name != expected.Name {
			return fmt.Errorf("Bad: VIP node '%s' has name '%s' (expected '%s')", vipNodeID, vipNode.Name, expected.Name)
		}

		if vipNode.IPv4Address != expected.IPv4Address {
			return fmt.Errorf("Bad: VIP node '%s' has IPv4 address '%s' (expected '%s')", vipNodeID, vipNode.IPv4Address, expected.IPv4Address)
		}

		if vipNode.IPv6Address != expected.IPv6Address {
			return fmt.Errorf("Bad: VIP node '%s' has IPv6 address '%s' (expected '%s')", vipNodeID, vipNode.IPv6Address, expected.IPv6Address)
		}
		if vipNode.HealthMonitor.Name != expected.HealthMonitor.Name {
			return fmt.Errorf("Bad: VIP node '%s' has health monitor '%s' (expected '%s')", vipNodeID, vipNode.HealthMonitor.Name, expected.HealthMonitor.Name)
		}

		if vipNode.Status != expected.Status {
			return fmt.Errorf("Bad: VIP node '%s' has status '%s' (expected '%s')", vipNodeID, vipNode.Status, expected.Status)
		}

		// TODO: Verify other properties.

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_vip_node:
//
// Check all vip nodes specified in the configuration have been destroyed.
func testCheckDDCloudVIPNodeDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_vip_node" {
			continue
		}

		vipNodeID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipNode, err := client.GetVIPNode(vipNodeID)
		if err != nil {
			return nil
		}
		if vipNode != nil {
			return fmt.Errorf("VIP node '%s' still exists", vipNodeID)
		}
	}

	return nil
}
