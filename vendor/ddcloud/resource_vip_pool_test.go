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

// Acceptance test configuration - ddcloud_vip_pool with basic properties
func testAccDDCloudVIPPoolBasic(poolName string, loadBalanceMethod string, serviceDownAction string, slowRampTime int) string {
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
			name					= "%s"
			description 			= "VIP pool for Terraform acceptance test."
			load_balance_method		= "%s"
			service_down_action		= "%s",
			slow_ramp_time			= %d,
      health_monitors = ["CCDEFAULT.Udp","CCDEFAULT.Tcp"]
			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}
	`, poolName, loadBalanceMethod, serviceDownAction, slowRampTime)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_vip_pool:
//
// Create a VIP pool and verify that it gets created with the correct configuration.
func TestAccVIPPoolBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudVIPPoolBasic(
					"AccTestPool",
					compute.LoadBalanceMethodRoundRobin,
					compute.ServiceDownActionDrop,
					5,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudVIPPoolExists("acc_test_pool", true),
					testCheckDDCloudVIPPoolMatches("acc_test_pool", compute.VIPPool{
						Name:              "AccTestPool",
						LoadBalanceMethod: compute.LoadBalanceMethodRoundRobin,
						ServiceDownAction: compute.ServiceDownActionDrop,
						SlowRampTime:      5,
						HealthMonitors: []compute.EntityReference{
							compute.EntityReference{
								Name: "CCDEFAULT.Tcp",
							},
							compute.EntityReference{
								Name: "CCDEFAULT.Udp",
							},
						},
					}),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_vip_pool (changing updatable properties causes in-place update):
//
// Create a VIP pool, then change some of its updatable properties, and verify that it gets updated in-place with the correct status.
func TestAccVIPPoolBasicUpdateInline(t *testing.T) {
	testAccResourceUpdateInPlace(t, testAccResourceUpdate{
		ResourceName: "ddcloud_vip_pool.acc_test_pool",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudVIPPoolBasic(
			"AccTestPool",
			compute.LoadBalanceMethodRoundRobin,
			compute.ServiceDownActionDrop,
			5,
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolExists("acc_test_pool", true),
			testCheckDDCloudVIPPoolMatches("acc_test_pool", compute.VIPPool{
				Name:              "AccTestPool",
				LoadBalanceMethod: compute.LoadBalanceMethodRoundRobin,
				ServiceDownAction: compute.ServiceDownActionDrop,
				SlowRampTime:      5,
				HealthMonitors: []compute.EntityReference{
					compute.EntityReference{
						Name: "CCDEFAULT.Tcp",
					},
					compute.EntityReference{
						Name: "CCDEFAULT.Udp",
					},
				},
			}),
		),

		// Update
		UpdateConfig: testAccDDCloudVIPPoolBasic(
			"AccTestPool",
			compute.LoadBalanceMethodLeastConnectionsNode,
			compute.ServiceDownActionNone,
			10,
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolExists("acc_test_pool", true),
			testCheckDDCloudVIPPoolMatches("acc_test_pool", compute.VIPPool{
				Name:              "AccTestPool",
				LoadBalanceMethod: compute.LoadBalanceMethodLeastConnectionsNode,
				ServiceDownAction: compute.ServiceDownActionNone,
				SlowRampTime:      10,
				HealthMonitors: []compute.EntityReference{
					compute.EntityReference{
						Name: "CCDEFAULT.Tcp",
					},
					compute.EntityReference{
						Name: "CCDEFAULT.Udp",
					},
				},
			}),
		),
	})
}

// Acceptance test for ddcloud_vip_pool (changing name causes destroy-and-recreate):
//
// Create a VIP pool, then change its name, and verify that it gets destroyed and recreated with the new name.
func TestAccVIPPoolBasicUpdateName(t *testing.T) {
	testAccResourceUpdateReplace(t, testAccResourceUpdate{
		ResourceName: "ddcloud_vip_pool.acc_test_pool",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudVIPPoolBasic(
			"AccTestPool",
			compute.LoadBalanceMethodRoundRobin,
			compute.ServiceDownActionDrop,
			5,
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolExists("acc_test_pool", true),
			testCheckDDCloudVIPPoolMatches("acc_test_pool", compute.VIPPool{
				Name:              "AccTestPool",
				LoadBalanceMethod: compute.LoadBalanceMethodRoundRobin,
				ServiceDownAction: compute.ServiceDownActionDrop,
				SlowRampTime:      5,
				HealthMonitors: []compute.EntityReference{
					compute.EntityReference{
						Name: "CCDEFAULT.Tcp",
					},
					compute.EntityReference{
						Name: "CCDEFAULT.Udp",
					},
				},
			}),
		),

		// Update
		UpdateConfig: testAccDDCloudVIPPoolBasic(
			"AccTestPool1",
			compute.LoadBalanceMethodRoundRobin,
			compute.ServiceDownActionDrop,
			5,
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudVIPPoolExists("acc_test_pool", true),
			testCheckDDCloudVIPPoolMatches("acc_test_pool", compute.VIPPool{
				Name:              "AccTestPool1",
				LoadBalanceMethod: compute.LoadBalanceMethodRoundRobin,
				ServiceDownAction: compute.ServiceDownActionDrop,
				SlowRampTime:      5,
				HealthMonitors: []compute.EntityReference{
					compute.EntityReference{
						Name: "CCDEFAULT.Tcp",
					},
					compute.EntityReference{
						Name: "CCDEFAULT.Udp",
					},
				},
			}),
		),
	})
}

/*
 * Acceptance-test checks
 */

// Acceptance test check for ddcloud_vip_pool:
//
// Check if the VIP pool exists.
func testCheckDDCloudVIPPoolExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_vip_pool")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vipPoolID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipPool, err := client.GetVIPPool(vipPoolID)
		if err != nil {
			return fmt.Errorf("Bad: Get VIP pool: %s", err)
		}
		if exists && vipPool == nil {
			return fmt.Errorf("Bad: VIP pool not found with Id '%s'.", vipPoolID)
		} else if !exists && vipPool != nil {
			return fmt.Errorf("Bad: VIP pool still exists with Id '%s'.", vipPoolID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_vip_pool:
//
// Check if the VIP pool's configuration matches the expected configuration.
func testCheckDDCloudVIPPoolMatches(name string, expected compute.VIPPool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_vip_pool")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		vipPoolID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipPool, err := client.GetVIPPool(vipPoolID)
		if err != nil {
			return fmt.Errorf("Bad: Get VIP pool: %s", err)
		}
		if vipPool == nil {
			return fmt.Errorf("Bad: VIP pool not found with Id '%s'", vipPoolID)
		}

		if vipPool.Name != expected.Name {
			return fmt.Errorf("Bad: VIP pool '%s' has name '%s' (expected '%s')", vipPoolID, vipPool.Name, expected.Name)
		}

		if vipPool.LoadBalanceMethod != expected.LoadBalanceMethod {
			return fmt.Errorf("Bad: VIP pool '%s' has load-balancing method '%s' (expected '%s')", vipPoolID, vipPool.LoadBalanceMethod, expected.LoadBalanceMethod)
		}

		if vipPool.ServiceDownAction != expected.ServiceDownAction {
			return fmt.Errorf("Bad: VIP pool '%s' has service-down action '%s' (expected '%s')", vipPoolID, vipPool.ServiceDownAction, expected.ServiceDownAction)
		}

		if vipPool.SlowRampTime != expected.SlowRampTime {
			return fmt.Errorf("Bad: VIP pool '%s' has slow-ramp time '%d' (expected '%d')", vipPoolID, vipPool.SlowRampTime, expected.SlowRampTime)
		}

		if len(vipPool.HealthMonitors) != len(expected.HealthMonitors) {
			return fmt.Errorf("Bad: VIP pool '%s' has health mointors count '%d' (expected '%d')", vipPoolID, len(vipPool.HealthMonitors), len(expected.HealthMonitors))
		}

		// TODO: Verify other properties.

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_vip_pool:
//
// Check all VIP pools specified in the configuration have been destroyed.
func testCheckDDCloudVIPPoolDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_vip_pool" {
			continue
		}

		vipPoolID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		vipPool, err := client.GetVIPPool(vipPoolID)
		if err != nil {
			return nil
		}
		if vipPool != nil {
			return fmt.Errorf("VIP pool '%s' still exists", vipPoolID)
		}
	}

	return nil
}
