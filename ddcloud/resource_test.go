package ddcloud

import (
	"fmt"
	"strings"
	"testing"

	"math/rand"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

/*
 * Common resource definitions for acceptance tests.
 */

const (
	testAccTargetRegion              = "AU"
	testAccResourceRandomName        = "acc_test"
	testAccResourceNetworkDomainName = "acc_test_domain"
	testAccResourceVLANName          = "acc_test_vlan"

	testAccResourceRandomByteLength        = 2
	testAccResourceNetworkDomainDatacenter = "AU9"
	testAccResourceVLANNetworkAddress      = "10.0.0.1/24"
)

// The default provider configuration for acceptance test runs.
func testAccProviderConfig() string {
	return testAccProviderConfigCustom(testAccTargetRegion)
}

// Custom provider configuration for acceptance test runs.
func testAccProviderConfigCustom(region string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region = "%s"
		}
	`, region)
}

// Resource to provide the default random number for acceptance test runs.
func testAccResourceRandom() string {
	return testAccResourceRandomCustom(
		testAccResourceRandomByteLength,
		testAccResourceRandomName,
	)
}

// Resource to provide a custom random number for acceptance test runs.
func testAccResourceRandomCustom(byteLength int, resourceName string) string {
	return fmt.Sprintf(`
		resource "random" "%s" {
			byte_length = %d
		}
	`, resourceName, byteLength)
}

// Resource representing the default network domain for acceptance tests.
func testAccResourceNetworkDomain() string {
	return testAccResourceNetworkDomainCustom(
		testAccResourceNetworkDomainDatacenter,
		testAccResourceRandomName,
		testAccResourceNetworkDomainName,
	)
}

// Resource representing a custom network domain for acceptance tests.
func testAccResourceNetworkDomainCustom(datacenter string, randomResourceName string, resourceName string) string {
	return fmt.Sprintf(`
		resource "ddcloud_networkdomain" "%s" {
			name        = "tf_acc_test_domain_${random.%s.b64}"
			description = "Network domain for Terraform acceptance test."
			datacenter  = "%s"
		}
	`, resourceName, randomResourceName, datacenter)
}

// Resource representing the default VLAN for acceptance tests.
func testAccResourceVLAN() string {
	return testAccResourceVLANCustom(
		testAccResourceVLANNetworkAddress,
		testAccResourceVLANName,
		testAccResourceRandomName,
		testAccResourceNetworkDomainName,
	)
}

// Resource representing a custom VLAN for acceptance tests.
func testAccResourceVLANCustom(ipv4Network string, networkDomainResourceName string, randomResourceName string, resourceName string) string {
	ipv4NetworkComponents := strings.Split(ipv4Network, "/")

	return fmt.Sprintf(`
		resource "ddcloud_vlan" "%s" {
			name				= "tf_acc_test_vlan_${random.%s.b64}"
			description			= "VLAN for Terraform acceptance test."

			ipv4_base_address	= "%s"
			ipv4_prefix_size	= %s
			
			networkdomain		= "%s"
		}
	`, resourceName, randomResourceName, networkDomainResourceName, ipv4NetworkComponents[0], ipv4NetworkComponents[1])
}

/*
 * Random numbers for acceptance tests.
 */

// Make a name unique by appending "-x", where x is a random number.
func testAccMakeUniqueName(name string) string {
	return fmt.Sprintf("%s-%d", name, rand.Int())
}

/*
 * Aggregate test helpers for resources.
 */

// Data structure that holds resource data for use between test steps.
type testAccResourceData struct {
	// Map from terraform resource names to provider resource Ids.
	NamesToResourceIDs map[string]string
}

func newTestAccResourceData() testAccResourceData {
	return testAccResourceData{
		NamesToResourceIDs: map[string]string{},
	}
}

// The configuration for a resource-update acceptance test.
type testAccResourceUpdate struct {
	ResourceName  string
	CheckDestroy  resource.TestCheckFunc
	InitialConfig string
	InitialCheck  resource.TestCheckFunc
	UpdateConfig  string
	UpdateCheck   resource.TestCheckFunc
}

// Aggregate test - update resource in-place (resource is updated, not destroyed and re-created).
func testAccResourceUpdateInPlace(test *testing.T, testDefinition testAccResourceUpdate) {
	resourceData := newTestAccResourceData()

	resource.Test(test, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testDefinition.CheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testDefinition.InitialConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckCaptureID(testDefinition.ResourceName, &resourceData),
					testDefinition.InitialCheck,
				),
			},
			resource.TestStep{
				Config: testDefinition.UpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceUpdatedInPlace(testDefinition.ResourceName, &resourceData),
					testDefinition.UpdateCheck,
				),
			},
		},
	})
}

// Aggregate test - update resource by replacing it (resource is destroyed and re-created).
func testAccResourceUpdateReplace(test *testing.T, testDefinition testAccResourceUpdate) {
	resourceData := newTestAccResourceData()

	resource.Test(test, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testDefinition.CheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testDefinition.InitialConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckCaptureID(testDefinition.ResourceName, &resourceData),
					testDefinition.InitialCheck,
				),
			},
			resource.TestStep{
				Config: testDefinition.UpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceReplaced(testDefinition.ResourceName, &resourceData),
					testDefinition.UpdateCheck,
				),
			},
		},
	})
}

// Acceptance test check helper:
//
// Capture the resource's Id.
func testCheckCaptureID(name string, testData *testAccResourceData) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		testData.NamesToResourceIDs[name] = res.Primary.ID

		return nil
	}
}

// Acceptance test check helper:
//
// Check if the resource was updated in-place (its Id has not changed).
func testCheckResourceUpdatedInPlace(name string, testData *testAccResourceData) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceType, err := getResourceTypeFromName(name)
		if err != nil {
			return err
		}

		capturedResourceID, ok := testData.NamesToResourceIDs[name]
		if !ok {
			return fmt.Errorf("No Id has been captured for resource '%s'.", name)
		}

		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		currentResourceID := res.Primary.ID
		if currentResourceID != capturedResourceID {
			return fmt.Errorf("Bad: The update was expected to be performed in-place but the Id for %s has changed (was: %s, now: %s) which indicates that the resource was destroyed and re-created", resourceType, capturedResourceID, currentResourceID)
		}

		return nil
	}
}

// Acceptance test check helper:
//
// Check if the resource was updated by destroying and re-creating it (its Id has changed).
func testCheckResourceReplaced(name string, testData *testAccResourceData) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceType, err := getResourceTypeFromName(name)
		if err != nil {
			return err
		}

		capturedResourceID, ok := testData.NamesToResourceIDs[name]
		if !ok {
			return fmt.Errorf("No Id has been captured for resource '%s'.", name)
		}

		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		currentResourceID := res.Primary.ID
		if currentResourceID == capturedResourceID {
			return fmt.Errorf("Bad: The update was expected to be performed by destroying and re-creating %s but its Id has not changed (still %s) which indicates that the resource was performed in-place", resourceType, currentResourceID)
		}

		return nil
	}
}

func getResourceTypeFromName(name string) (string, error) {
	resourceNameComponents := strings.SplitN(name, ".", 2)
	if len(resourceNameComponents) != 2 {
		return "", fmt.Errorf("Invalid resource name: '%s' (should be 'resource_type.resource_name')", name)
	}

	return resourceNameComponents[0], nil
}

// Ensure that a resource name starts with the specified resource-type prefix.
func ensureResourceTypePrefix(resourceName string, resourceType string) string {
	prefix := resourceType + "."

	if strings.HasPrefix(resourceName, prefix) {
		return resourceName
	}

	return prefix + resourceName
}

// Ensure that a data-source name starts with the specified data-source-type prefix.
func ensureDataSourceTypePrefix(dataSourceName string, dataSourceType string) string {
	prefix := "data." + dataSourceType + "."

	if strings.HasPrefix(dataSourceType, prefix) {
		return dataSourceName
	}

	return prefix + dataSourceName
}
