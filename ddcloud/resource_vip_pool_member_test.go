package ddcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

/*
 * Acceptance-test configurations.
 */

// Acceptance test configuration - ddcloud_vip_pool_member with basic properties
func testAccDDCloudVIPPoolMemberBasic(port *int) string {
	var memberPort = ""
	if port != nil {
		memberPort = fmt.Sprintf("port = %d", *port)
	}

	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc_test_domain"
			description			= "VIP pool for Terraform acceptance test."
			datacenter			= "AU9"

			plan				= "ADVANCED"
		}

		resource "ddcloud_vip_pool" "acc_test_pool" {
			name					= "acc_test_tool"
			description 			= "VIP pool for Terraform acceptance test."
			load_balance_method		= "ROUND_ROBIN"
			service_down_action		= "NONE",
			slow_ramp_time			= 10,

			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}

		resource "ddcloud_vip_node" "acc_test_node" {
			name					= "acc_test_node"
			description 			= "VIP node for Terraform acceptance test."
			ipv4_address			= "192.168.17.10"
			status					= "ENABLED"

			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}

		resource "ddcloud_vip_pool_member" "acc_test_pool_member" {
			pool					= "${ddcloud_vip_pool.acc_test_pool.id}"
			node 					= "${ddcloud_vip_node.acc_test_node.id}"

			%s
		}
	`, memberPort)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_vip_pool_member:
//
// Create a VIP pool, a VIP node, and a membership (without port constraint) between them, and verify that the membership gets created with the correct configuration.
func TestAccVIPPoolMemberBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolMemberDestroy,
			testCheckDDCloudVIPNodeDestroy,
			testCheckDDCloudVIPPoolDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudVIPPoolMemberBasic(
					nil, /* no port */
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVIPPoolMemberExists("acc_test_pool_member", true),
					testCheckDDCloudVIPPoolMemberMatches("acc_test_pool_member",
						nil, /* no port */
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_vip_pool_member:
//
// Create a VIP pool, a VIP node, and a membership (with port constraint) between them, and verify that the membership gets created with the correct configuration.
func TestAccVIPPoolMemberBasicWithPortCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolMemberDestroy,
			testCheckDDCloudVIPNodeDestroy,
			testCheckDDCloudVIPPoolDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudVIPPoolMemberBasic(
					intToPtr(8080), // port 8080
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVIPPoolMemberExists("acc_test_pool_member", true),
					testCheckDDCloudVIPPoolMemberMatches("acc_test_pool_member",
						intToPtr(8080), // port 8080
					),
				),
			},
		},
	})
}

/*
 * Acceptance-test checks
 */

// Acceptance test check for ddcloud_vip_pool:
//
// Check if the VIP pool exists.
func testCheckDDCloudVIPPoolMemberExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_vip_pool_member")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vipPoolID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipPool, err := client.GetVIPPoolMember(vipPoolID)
		if err != nil {
			return fmt.Errorf("bad: Get VIP pool member: %s", err)
		}
		if exists && vipPool == nil {
			return fmt.Errorf("bad: VIP pool member not found with Id '%s'", vipPoolID)
		} else if !exists && vipPool != nil {
			return fmt.Errorf("bad: VIP pool member still exists with Id '%s'", vipPoolID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_vip_pool_member:
//
// Check if the VIP pool's configuration matches the expected configuration.
func testCheckDDCloudVIPPoolMemberMatches(name string, expectedPort *int) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_vip_pool_member")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vipPoolMemberID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipPoolMember, err := client.GetVIPPoolMember(vipPoolMemberID)
		if err != nil {
			return fmt.Errorf("bad: Get VIP pool member: %s", err)
		}
		if vipPoolMember == nil {
			return fmt.Errorf("bad: VIP pool member not found with Id '%s'", vipPoolMemberID)
		}

		if vipPoolMember.Port != nil {
			if expectedPort == nil {
				return fmt.Errorf("bad: VIP pool member '%s' is constrained to port %d (expected no port constraint)", vipPoolMemberID, *vipPoolMember.Port)
			}
		} else if expectedPort != nil {
			if vipPoolMember.Port == nil {
				return fmt.Errorf("bad: VIP pool member '%s' has no port constraint (expected it to be constrained to port %d)", vipPoolMemberID, *expectedPort)
			}
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_vip_pool_member:
//
// Check all VIP pools specified in the configuration have been destroyed.
func testCheckDDCloudVIPPoolMemberDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_vip_pool_member" {
			continue
		}

		vipPoolMemberID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipPoolMember, err := client.GetVIPPoolMember(vipPoolMemberID)
		if err != nil {
			return nil
		}
		if vipPoolMember != nil {
			return fmt.Errorf("VIP pool member '%s' still exists", vipPoolMemberID)
		}
	}

	return nil
}
