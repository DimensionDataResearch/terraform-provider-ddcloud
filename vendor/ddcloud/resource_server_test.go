package ddcloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

/*
 * Acceptance-test configurations.
 */

// A basic Server (and its accompanying network domain and VLAN).
func testAccDDCloudServerBasic(name string, description string, primaryIPv4Address string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"
		}

		resource "ddcloud_vlan" "acc_test_vlan" {
			name				= "acc-test-vlan"
			description 		= "VLAN for Terraform acceptance test."

			networkdomain 		= "${ddcloud_networkdomain.acc_test_domain.id}"

			ipv4_base_address	= "192.168.17.0"
			ipv4_prefix_size	= 24
		}

		resource "ddcloud_server" "acc_test_server" {
			name				 = "%s"
			description 		 = "%s"
			admin_password		 = "snausages!"

			memory_gb			 = 8

			networkdomain 		 = "${ddcloud_networkdomain.acc_test_domain.id}"
			primary_adapter_vlan = "${ddcloud_vlan.acc_test_vlan.id}"
			primary_adapter_ipv4 = "%s"

			dns_primary			 = "8.8.8.8"
			dns_secondary		 = "8.8.4.4"

			os_image_name		 = "CentOS 7 64-bit 2 CPU"

			auto_start			 = false

			# Image disk
			disk {
				scsi_unit_id     = 0
				size_gb          = 10
				speed            = "STANDARD"
			}
		}
	`, name, description, primaryIPv4Address)
}

// A Server (and its accompanying network domain and VLAN) with a single image disk.
func testAccDDCloudServerImageDisk1(sizeGB int, speed string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"
		}

		resource "ddcloud_vlan" "acc_test_vlan" {
			name				= "acc-test-vlan"
			description 		= "VLAN for Terraform acceptance test."

			networkdomain 		= "${ddcloud_networkdomain.acc_test_domain.id}"

			ipv4_base_address	= "192.168.17.0"
			ipv4_prefix_size	= 24
		}

		resource "ddcloud_server" "acc_test_server" {
			name				 = "acc-test-server-1-image-disk"
			description 		 = "Server for Terraform acceptance test (single image disk)."
			admin_password		 = "snausages!"

			memory_gb			 = 8

			networkdomain 		 = "${ddcloud_networkdomain.acc_test_domain.id}"
			primary_adapter_vlan = "${ddcloud_vlan.acc_test_vlan.id}"
			primary_adapter_ipv4 = "192.168.17.6"

			dns_primary			 = "8.8.8.8"
			dns_secondary		 = "8.8.4.4"

			os_image_name		 = "CentOS 7 64-bit 2 CPU"

			auto_start			 = false

			# Image disk
			disk {
				scsi_unit_id     = 0
				size_gb          = %d
				speed            = "%s"
			}
		}
	`, sizeGB, speed)
}

// A Server (and its accompanying network domain and VLAN) with a single additional disk.
func testAccDDCloudServerAdditionalDisk1(scsiUnitID int, sizeGB int, speed string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"
		}

		resource "ddcloud_vlan" "acc_test_vlan" {
			name				= "acc-test-vlan"
			description 		= "VLAN for Terraform acceptance test."

			networkdomain 		= "${ddcloud_networkdomain.acc_test_domain.id}"

			ipv4_base_address	= "192.168.17.0"
			ipv4_prefix_size	= 24
		}

		resource "ddcloud_server" "acc_test_server" {
			name				 = "acc-test-server-1-additional-disk"
			description 		 = "Server for Terraform acceptance test (single additional disk)."
			admin_password		 = "snausages!"

			memory_gb			 = 8

			networkdomain 		 = "${ddcloud_networkdomain.acc_test_domain.id}"
			primary_adapter_vlan = "${ddcloud_vlan.acc_test_vlan.id}"
			primary_adapter_ipv4 = "192.168.17.6"

			dns_primary			 = "8.8.8.8"
			dns_secondary		 = "8.8.4.4"

			os_image_name		 = "CentOS 7 64-bit 2 CPU"

			auto_start			 = false

			# Image disk
			disk {
				scsi_unit_id     = 0
				size_gb          = 10
				speed            = "STANDARD"
			}

			# Additional disk
			disk {
				scsi_unit_id     = %d
				size_gb          = %d
				speed            = "%s"
			}
		}
	`, scsiUnitID, sizeGB, speed)
}

// A Server with tags (and its accompanying network domain and VLAN).
func testAccDDCloudServerTag(tags map[string]string) string {
	tagConfiguration := ""
	for tagName := range tags {
		tagValue := tags[tagName]
		tagConfiguration += fmt.Sprintf(`
			tag {
				name  = "%s"
				value = "%s"
			}
		`, tagName, tagValue)
	}

	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name		= "acc-test-networkdomain"
			description	= "Network domain for Terraform acceptance test."
			datacenter	= "AU9"
		}

		resource "ddcloud_vlan" "acc_test_vlan" {
			name				= "acc-test-vlan"
			description 		= "VLAN for Terraform acceptance test."

			networkdomain 		= "${ddcloud_networkdomain.acc_test_domain.id}"

			ipv4_base_address	= "192.168.17.0"
			ipv4_prefix_size	= 24
		}

		resource "ddcloud_server" "acc_test_server" {
			name				 = "acc-test-server-tags"
			description 		 = "Server for Terraform acceptance test (tags)."
			admin_password		 = "snausages!"

			memory_gb			 = 8

			networkdomain 		 = "${ddcloud_networkdomain.acc_test_domain.id}"
			primary_adapter_vlan = "${ddcloud_vlan.acc_test_vlan.id}"
			primary_adapter_ipv4 = "192.168.17.6"

			dns_primary			 = "8.8.8.8"
			dns_secondary		 = "8.8.4.4"

			os_image_name		 = "CentOS 7 64-bit 2 CPU"

			# Image disk
			disk {
				scsi_unit_id     = 0
				size_gb          = 10
				speed            = "STANDARD"
			}

			# Tags
			%s
		}
	`, tagConfiguration)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_server (basic):
//
// Create a server and verify that it gets created with the correct configuration.
func TestAccServerBasicCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudServerDestroy,
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerBasic("acc-test-server",
					"Server for Terraform acceptance test.",
					"192.168.17.6",
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerMatches("ddcloud_server.acc_test_server",
						"ddcloud_networkdomain.acc_test_domain",
						compute.Server{
							Name:        "acc-test-server",
							Description: "Server for Terraform acceptance test.",
							MemoryGB:    8,
							Network: compute.VirtualMachineNetwork{
								PrimaryAdapter: compute.VirtualMachineNetworkAdapter{
									PrivateIPv4Address: stringToPtr("192.168.17.6"),
								},
							},
						},
					),
					testCheckDDCloudServerDiskMatches("ddcloud_server.acc_test_server",
						testImageDiskCentOS7(10, "STANDARD"),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_server (1 additional disk):
//
// Create a server with a single image disk and verify that the image disk is resized once the server has been deployed.
func TestAccServerImageDisk1ResizeCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudServerDestroy,
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerImageDisk1(15, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerDiskMatches("ddcloud_server.acc_test_server",
						testImageDiskCentOS7(15, "STANDARD"),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_server (1 additional disk):
//
// Create a server with a single image disk, then update it and verify that the image disk is resized.
func TestAccServerImageDisk1ResizeUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudServerDestroy,
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerImageDisk1(10, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerDiskMatches("ddcloud_server.acc_test_server",
						testImageDiskCentOS7(10, "STANDARD"),
					),
				),
			},
			resource.TestStep{
				Config: testAccDDCloudServerImageDisk1(15, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerDiskMatches("ddcloud_server.acc_test_server",
						testImageDiskCentOS7(15, "STANDARD"),
					),
				),
			},
		},
	})
}

// TODO: TestAccServerAdditionalDisk1RemoveUpdate

// Acceptance test for ddcloud_server (1 additional disk):
//
// Create a server with a single additional disk and verify that it gets created with the correct configuration.
func TestAccServerAdditionalDisk1Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudServerDestroy,
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerAdditionalDisk1(1, 15, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerDiskMatches("ddcloud_server.acc_test_server",
						testImageDiskCentOS7(10, "STANDARD"),
						compute.VirtualMachineDisk{
							SCSIUnitID: 1,
							SizeGB:     15,
							Speed:      "STANDARD",
						},
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_server (tags):
//
// Create a server with 2 tags and verify that it gets created with the correct tags.
func TestAccServerTagCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudServerDestroy,
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerTag(map[string]string{
					"role":      "hello world",
					"consul_dc": "goodbye moon",
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerTagMatches("ddcloud_server.acc_test_server", map[string]string{
						"role":      "hello world",
						"consul_dc": "goodbye moon",
					}),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_server (tags):
//
// Create a server with 2 tags and verify that it gets created with the correct tags.
func TestAccServerTagUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudServerDestroy,
			testCheckDDCloudVLANDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudServerTag(map[string]string{
					"role":      "hello world",
					"consul_dc": "goodbye moon",
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerTagMatches("ddcloud_server.acc_test_server", map[string]string{
						"role":      "hello world",
						"consul_dc": "goodbye moon",
					}),
				),
			},
			resource.TestStep{
				Config: testAccDDCloudServerTag(map[string]string{
					"role":      "greetings, earth",
					"consul_dc": "farewell, luna",
				}),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudServerExists("ddcloud_server.acc_test_server", true),
					testCheckDDCloudServerTagMatches("ddcloud_server.acc_test_server", map[string]string{
						"role":      "greetings, earth",
						"consul_dc": "farewell, luna",
					}),
				),
			},
		},
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_server:
//
// Check if the server exists.
func testCheckDDCloudServerExists(name string, exists bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		serverID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		server, err := client.GetServer(serverID)
		if err != nil {
			return fmt.Errorf("Bad: Get server: %s", err)
		}
		if exists && server == nil {
			return fmt.Errorf("Bad: Server not found with Id '%s'.", serverID)
		} else if !exists && server != nil {
			return fmt.Errorf("Bad: Server still exists with Id '%s'.", serverID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_server:
//
// Check if the server's configuration matches the expected configuration.
func testCheckDDCloudServerMatches(name string, networkDomainName string, expected compute.Server) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		serverResource, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		serverID := serverResource.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		server, err := client.GetServer(serverID)
		if err != nil {
			return fmt.Errorf("Bad: Get server: %s", err)
		}
		if server == nil {
			return fmt.Errorf("Bad: Server not found with Id '%s'", serverID)
		}

		if server.Name != expected.Name {
			return fmt.Errorf("Bad: Server '%s' has name '%s' (expected '%s')", serverID, server.Name, expected.Name)
		}

		if server.Description != expected.Description {
			return fmt.Errorf("Bad: Server '%s' has name '%s' (expected '%s')", serverID, server.Description, expected.Description)
		}

		if server.MemoryGB != expected.MemoryGB {
			return fmt.Errorf("Bad: Server '%s' has been allocated %dGB of memory (expected %dGB)", serverID, server.MemoryGB, expected.MemoryGB)
		}

		expectedPrimaryIPv4 := *expected.Network.PrimaryAdapter.PrivateIPv4Address
		actualPrimaryIPv4 := *server.Network.PrimaryAdapter.PrivateIPv4Address
		if actualPrimaryIPv4 != expectedPrimaryIPv4 {
			return fmt.Errorf("Bad: Primary network adapter for server '%s' has IPv4 address '%s' (expected '%s')", serverID, actualPrimaryIPv4, expectedPrimaryIPv4)
		}

		expectedPrimaryIPv6, ok := serverResource.Primary.Attributes[resourceKeyServerPrimaryAdapterIPv6]
		if !ok {
			return fmt.Errorf("Bad: %s.%s is missing '%s' attribute.", serverResource.Type, name, resourceKeyServerPrimaryAdapterIPv6)
		}

		actualPrimaryIPv6 := *server.Network.PrimaryAdapter.PrivateIPv6Address
		if actualPrimaryIPv6 != expectedPrimaryIPv6 {
			return fmt.Errorf("Bad: Primary network adapter for server '%s' has IPv6 address '%s' (expected '%s')", serverID, actualPrimaryIPv6, expectedPrimaryIPv6)
		}

		networkDomainResource := state.RootModule().Resources[networkDomainName]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		expectedNetworkDomainID := networkDomainResource.Primary.ID
		if server.Network.NetworkDomainID != expectedNetworkDomainID {
			return fmt.Errorf("Bad: Server '%s' is part of network domain '%s' (expected '%s')", serverID, server.Network.NetworkDomainID, expectedNetworkDomainID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_server:
//
// Check if the server's disk configuration matches the expected configuration.
func testCheckDDCloudServerDiskMatches(name string, expected ...compute.VirtualMachineDisk) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		serverResource, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		serverID := serverResource.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		server, err := client.GetServer(serverID)
		if err != nil {
			return fmt.Errorf("Bad: Get server: %s", err)
		}
		if server == nil {
			return fmt.Errorf("Bad: Server not found with Id '%s'", serverID)
		}

		var validationMessages []string
		expectedDisksByUnitID := getDisksByUnitID(expected)
		for _, actualDisk := range server.Disks {
			expectedDisk, ok := expectedDisksByUnitID[actualDisk.SCSIUnitID]
			if !ok {
				validationMessages = append(validationMessages, fmt.Sprintf(
					"found unexpected server disk '%s' with SCSI unit ID %d.",
					*actualDisk.ID,
					actualDisk.SCSIUnitID,
				))

				continue
			}
			delete(expectedDisksByUnitID, actualDisk.SCSIUnitID)

			if actualDisk.SizeGB != expectedDisk.SizeGB {
				validationMessages = append(validationMessages, fmt.Sprintf(
					"server disk '%s' with SCSI unit ID %d has size %dGB (expected %dGB).",
					*actualDisk.ID,
					actualDisk.SCSIUnitID,
					actualDisk.SizeGB,
					expectedDisk.SizeGB,
				))
			}

			if actualDisk.Speed != expectedDisk.Speed {
				validationMessages = append(validationMessages, fmt.Sprintf(
					"server disk '%s' with SCSI unit ID %d has speed '%s' (expected '%s').",
					*actualDisk.ID,
					actualDisk.SCSIUnitID,
					actualDisk.Speed,
					expectedDisk.Speed,
				))
			}
		}

		for expectedUnitID := range expectedDisksByUnitID {
			expectedDisk := expectedDisksByUnitID[expectedUnitID]

			validationMessages = append(validationMessages, fmt.Sprintf(
				"no server disk was found with SCSI unit ID %d.",
				expectedDisk.SCSIUnitID,
			))
		}

		if len(validationMessages) > 0 {
			return fmt.Errorf("Bad: %s", strings.Join(validationMessages, ", "))
		}

		return nil
	}
}

// Acceptance test check for ddcloud_server:
//
// Check if the server's tags match the expected tags.
func testCheckDDCloudServerTagMatches(name string, expected map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		serverResource, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		serverID := serverResource.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		tags, err := client.GetAssetTags(serverID, compute.AssetTypeServer, nil)
		if err != nil {
			return fmt.Errorf("Bad: Get server: %s", err)
		}

		expectedTags := make(map[string]string)
		for tagName := range expected {
			expectedTags[tagName] = expected[tagName]
		}

		var validationMessages []string

		for _, actualTag := range tags.Items {
			expectedTagValue, ok := expectedTags[actualTag.Name]
			if !ok {
				validationMessages = append(validationMessages, fmt.Sprintf(
					"found unexpected tag '%s' on server '%s'.",
					actualTag.Name,
					serverID,
				))

				continue
			}
			delete(expectedTags, actualTag.Name)

			if actualTag.Value != expectedTagValue {
				validationMessages = append(validationMessages, fmt.Sprintf(
					"tag '%s' on server '%s' has value '%s' (expected '%s').",
					actualTag.Name,
					serverID,
					actualTag.Value,
					expectedTagValue,
				))
			}
		}

		for expectedTagName := range expectedTags {
			expectedTagValue := expectedTags[expectedTagName]

			validationMessages = append(validationMessages, fmt.Sprintf(
				"no tag was found named '%s' (with value '%s') on server '%s'.",
				expectedTagName,
				expectedTagValue,
				serverID,
			))
		}

		if len(validationMessages) > 0 {
			return fmt.Errorf("Bad: %s", strings.Join(validationMessages, ", "))
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_server:
//
// Check all servers specified in the configuration have been destroyed.
func testCheckDDCloudServerDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_server" {
			continue
		}

		serverID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		server, err := client.GetServer(serverID)
		if err != nil {
			return nil
		}
		if server != nil {
			return fmt.Errorf("Server '%s' still exists", serverID)
		}
	}

	return nil
}

/*
 * Test disk definitions.
 */

// The image disk definition for CentOS 7.
func testImageDiskCentOS7(sizeGB int, speed string) compute.VirtualMachineDisk {
	return compute.VirtualMachineDisk{
		SCSIUnitID: 0,
		SizeGB:     sizeGB,
		Speed:      speed,
	}
}
