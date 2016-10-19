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

// Acceptance test configuration - ddcloud_port_list with ports and port ranges.
func testAccDDCloudPortListBasic(resourceName string, portListName string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Port list for Terraform acceptance test."
			datacenter			= "AU9"

			plan				= "ADVANCED"
		}

		resource "ddcloud_port_list" "%s" {
			name					= "%s"
			description 			= "Adam's Terraform test port list (do not delete)."

			port {
				begin			= 80
			}
			port {
				begin			= 443
			}

			port {
				begin			= 8000
				end				= 9100
			}

			networkdomain 			= "${ddcloud_networkdomain.acc_test_domain.id}"
		}
	`, resourceName, portListName)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_port_list:
//
// Create a port list and verify that it gets created with the correct configuration.
func TestAccPortListBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudPortListDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudPortListBasic("acc_test_list", "af_terraform_list"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudPortListExists("acc_test_list", true),
					testCheckDDCloudPortListMatches("acc_test_list", compute.PortList{
						Name:        "af_terraform_list",
						Description: "Adam's Terraform test port list (do not delete).",
						Ports: []compute.PortListEntry{
							compute.PortListEntry{
								Begin: 80,
							},
							compute.PortListEntry{
								Begin: 443,
							},
							compute.PortListEntry{
								Begin: 8000,
								End:   intToPtr(9100),
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

// Acceptance test check for ddcloud_port_list:
//
// Check if the port list exists.
func testCheckDDCloudPortListExists(name string, exists bool) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_port_list")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		portListID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		portList, err := client.GetPortList(portListID)
		if err != nil {
			return fmt.Errorf("Bad: Get port list: %s", err)
		}
		if exists && portList == nil {
			return fmt.Errorf("Bad: port list not found with Id '%s'.", portListID)
		} else if !exists && portList != nil {
			return fmt.Errorf("Bad: port list still exists with Id '%s'.", portListID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_port_list:
//
// Check if the port list's configuration matches the expected configuration.
func testCheckDDCloudPortListMatches(name string, expected compute.PortList) resource.TestCheckFunc {
	name = ensureResourceTypePrefix(name, "ddcloud_port_list")

	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		portListID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		portList, err := client.GetPortList(portListID)
		if err != nil {
			return fmt.Errorf("Bad: Get port list: %s", err)
		}
		if portList == nil {
			return fmt.Errorf("Bad: port list not found with Id '%s'", portListID)
		}

		if portList.Name != expected.Name {
			return fmt.Errorf("Bad: port list '%s' has name '%s' (expected '%s')", portListID, portList.Name, expected.Name)
		}

		if portList.Description != expected.Description {
			return fmt.Errorf("Bad: port list '%s' has description '%s' (expected '%s')", portListID, portList.Description, expected.Description)
		}

		if len(portList.Ports) != len(expected.Ports) {
			return fmt.Errorf("Bad: port list '%s' has %d ports or port ranges (expected '%d')", portListID, len(portList.Ports), len(expected.Ports))
		}

		err = comparePortListEntries(expected, *portList)
		if err != nil {
			return err
		}

		if len(portList.ChildLists) != len(expected.ChildLists) {
			return fmt.Errorf("Bad: port list '%s' has %d child lists (expected '%d')", portListID, len(portList.ChildLists), len(expected.ChildLists))
		}

		for index := range portList.ChildLists {
			expectedChildListID := expected.ChildLists[index].ID
			actualChildListID := portList.ChildLists[index].ID

			if actualChildListID != expectedChildListID {
				return fmt.Errorf("Bad: port list '%s' has child list at index %d with Id %s (expected '%s')",
					portListID, index, actualChildListID, expectedChildListID,
				)
			}
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_port_list:
//
// Check all port lists specified in the configuration have been destroyed.
func testCheckDDCloudPortListDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_port_list" {
			continue
		}

		portListID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		portList, err := client.GetPortList(portListID)
		if err != nil {
			return nil
		}
		if portList != nil {
			return fmt.Errorf("port list '%s' still exists", portListID)
		}
	}

	return nil
}

func comparePortListEntries(expectedPortList compute.PortList, actualPortList compute.PortList) error {
	portListID := actualPortList.ID

	for index := range expectedPortList.Ports {
		expectedPort := expectedPortList.Ports[index]
		actualPort := actualPortList.Ports[index]

		if expectedPort.Begin != actualPort.Begin {
			return fmt.Errorf("Bad: port list '%s' has entry at index %d with begin port %s (expected '%s')",
				portListID, index, formatPort(actualPort.Begin), formatPort(expectedPort.Begin),
			)
		}

		expectedPortEnd := formatPort(expectedPort.End)
		actualPortEnd := formatPort(actualPort.End)
		if expectedPortEnd != actualPortEnd {
			return fmt.Errorf("Bad: port list '%s' has entry at index %d with end port %s (expected %s)",
				portListID, index, actualPortEnd, expectedPortEnd,
			)
		}
	}

	return nil
}

func formatPort(port interface{}) string {
	switch typedPort := port.(type) {
	case int:
		return fmt.Sprintf("%d", typedPort)
	case *int:
		if typedPort == nil {
			return "nil"
		}

		return fmt.Sprintf("%d", *typedPort)
	}

	return "unknown"
}
