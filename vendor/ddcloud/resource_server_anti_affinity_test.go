package ddcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

/*
 * Acceptance-test configurations.
 */

func testAccDDCloudServerAntiAffinityRuleBasic() string {
	return `
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"

			plan		= "ADVANCED"
		}

		resource "ddcloud_vlan" "acc_test_vlan" {
			name				= "acc-test-vlan"
			description 		= "VLAN for Terraform acceptance test."

			networkdomain 		= "${ddcloud_networkdomain.acc_test_domain.id}"

			ipv4_base_address	= "192.168.17.0"
			ipv4_prefix_size	= 24
		}

		resource "ddcloud_server" "acc_test_server" {
			count					= 2
			name					= "acc_test_server-${format("%d", count.index + 1)}"
			description 			= "Server ${format("%d", count.index + 1)} for Terraform anti-affinity acceptance test."
			admin_password			= "snausages!"

			memory_gb				= 8

			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
			primary_adapter_vlan	= "${ddcloud_vlan.acc_test_vlan.id}"
			primary_adapter_ipv4	= "192.168.17.${count.index + 6}"

			dns_primary				= "8.8.8.8"
			dns_secondary			= "8.8.4.4"

			os_image_name			= "CentOS 7 64-bit 2 CPU"

			auto_start				= false

			# Image disk
			disk {
				scsi_unit_id     = 0
				size_gb          = 10
				speed            = "STANDARD"
			}
		}

		resource "ddcloud_server_anti_affinity" "acc_test_anti_affinity_rule" {
			server1 = "${element(ddcloud_server.acc_test_server.*.id, 0)}"
			server2 = "${element(ddcloud_server.acc_test_server.*.id, 1)}"
		}
	`
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_server_anti_affinity (basic):
//
// Create a server anti-affinity rule and verify that it gets created with the correct configuration.
func TestAccServerAntiAffinityRuleBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testCheckDDCloudServerAntiAffinityRuleDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerAntiAffinityRuleBasic(),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerAntiAffinityRuleExists("acc_test_anti_affinity_rule", true),
					// TODO: Validate rule targets correct servers.
				),
			},
		},
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_server_anti_affinity:
//
// Check if the server anti-affinity rule exists.
func testCheckDDCloudServerAntiAffinityRuleExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_server_anti_affinity")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		ruleID := res.Primary.ID
		networkDomainID := res.Primary.Attributes[resourceKeyServerAntiAffinityRuleNetworkDomainID]

		client := testAccProvider.Meta().(*providerState).Client()
		networkDomain, err := client.GetServerAntiAffinityRule(ruleID, networkDomainID)
		if err != nil {
			return fmt.Errorf("Bad: Get server anti-affinity rule: %s", err.Error())
		}
		if exists && networkDomain == nil {
			return fmt.Errorf("Bad: Server anti-affinity rule not found with Id '%s' in network domain '%s'", ruleID, networkDomainID)
		} else if !exists && networkDomain != nil {
			return fmt.Errorf("Bad: Server anti-affinity rule still exists with Id '%s' in network domain '%s'", ruleID, networkDomainID)
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_server_anti_affinity:
//
// Check all server anti-affinity rules specified in the configuration have been destroyed.
func testCheckDDCloudServerAntiAffinityRuleDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_server_anti_affinity" {
			continue
		}

		ruleID := res.Primary.ID
		networkDomainID := res.Primary.Attributes[resourceKeyServerAntiAffinityRuleNetworkDomainID]

		client := testAccProvider.Meta().(*providerState).Client()
		networkDomain, err := client.GetServerAntiAffinityRule(ruleID, networkDomainID)
		if err != nil {
			return nil
		}
		if networkDomain != nil {
			return fmt.Errorf("Server anti-affinity rule '%s' still exists in network domain '%s'", ruleID, networkDomainID)
		}
	}

	return nil
}
