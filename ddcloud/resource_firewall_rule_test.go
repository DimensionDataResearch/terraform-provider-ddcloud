package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strings"
	"testing"
)

// ipv6.internode.on.net (CloudControl does not allow "ANY" as a source address for IPv6 rules)
const testIPv6Address = "2001:44b8:8020:f501:250:56ff:feb3:6633"

/*
 * Acceptance-test configurations.
 */

// Acceptance test configuration - ddcloud_firewall_rule (IP, from source address to destination address)
func testAccDDCloudFirewallRuleIPFromHostToHost(name string, ipVersion string, sourceHost string, destinationHost string, enabled bool) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Firewall rule for Terraform acceptance test."
			datacenter			= "AU9"
		}

		resource "ddcloud_firewall_rule" "acc_test_rule" {
			name				= "%s"
			ip_version			= "%s"
			protocol			= "IP"

			source_address		= "%s"
			destination_address	= "%s"

			action				= "ACCEPT_DECISIVELY"
			placement			= "FIRST"

			enabled				= %t

			networkdomain		= "${ddcloud_networkdomain.acc_test_domain.id}"
		}`,
		name, ipVersion, sourceHost, destinationHost, enabled,
	)
}

// Acceptance test configuration - ddcloud_firewall_rule (IP, from source network to destination address)
func testAccDDCloudFirewallRuleIPFromNetworkToHost(name string, ipVersion string, sourceNetwork string, destinationHost string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Firewall rule for Terraform acceptance test."
			datacenter			= "AU9"
		}

		resource "ddcloud_firewall_rule" "acc_test_rule" {
			name				= "%s"
			ip_version			= "%s"
			protocol			= "IP"

			source_network		= "%s"
			destination_address	= "%s"

			action				= "ACCEPT_DECISIVELY"
			placement			= "FIRST"

			enabled				= true

			networkdomain		= "${ddcloud_networkdomain.acc_test_domain.id}"
		}`,
		name, ipVersion, sourceNetwork, destinationHost,
	)
}

// Acceptance test configuration - ddcloud_firewall_rule (TCP, from source address to destination address and destination port)
func testAccDDCloudFirewallRuleTCPFromHostToHost(name string, ipVersion string, sourceHost string, destinationHost string, destinationPort int) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Firewall rule for Terraform acceptance test."
			datacenter			= "AU9"
		}

		resource "ddcloud_firewall_rule" "acc_test_rule" {
			name				= "%s"
			ip_version			= "%s"
			protocol			= "TCP"

			source_address		= "%s"
			destination_address	= "%s"
			destination_port	= %d

			action				= "ACCEPT_DECISIVELY"
			placement			= "FIRST"

			enabled				= true

			networkdomain		= "${ddcloud_networkdomain.acc_test_domain.id}"
		}`,
		name, ipVersion, sourceHost, destinationHost, destinationPort,
	)
}

// Acceptance test configuration - ddcloud_firewall_rule (ICMP, from source address to destination address)
func testAccDDCloudFirewallRuleICMPFromHostToHost(name string, ipVersion string, sourceHost string, destinationHost string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Firewall rule for Terraform acceptance test."
			datacenter			= "AU9"
		}

		resource "ddcloud_firewall_rule" "acc_test_rule" {
			name				= "%s"
			ip_version			= "%s"
			protocol			= "ICMP"

			source_address		= "%s"
			destination_address	= "%s"

			action				= "ACCEPT_DECISIVELY"
			placement			= "FIRST"

			enabled				= true

			networkdomain		= "${ddcloud_networkdomain.acc_test_domain.id}"
		}`,
		name, ipVersion, sourceHost, destinationHost,
	)
}

// Acceptance test configuration - ddcloud_firewall_rule (ICMP, from source address to destination address)
func testAccDDCloudFirewallRuleICMPFromNetworkToHost(name string, ipVersion string, sourceNetwork string, destinationHost string) string {
	return fmt.Sprintf(`
		provider "ddcloud" {
			region		= "AU"
		}

		resource "ddcloud_networkdomain" "acc_test_domain" {
			name				= "acc-test-domain"
			description			= "Firewall rule for Terraform acceptance test."
			datacenter			= "AU9"
		}

		resource "ddcloud_firewall_rule" "acc_test_rule" {
			name				= "%s"
			ip_version			= "%s"
			protocol			= "ICMP"

			source_network		= "%s"
			destination_address	= "%s"

			action				= "ACCEPT_DECISIVELY"
			placement			= "FIRST"

			enabled				= true

			networkdomain		= "${ddcloud_networkdomain.acc_test_domain.id}"
		}`,
		name, ipVersion, sourceNetwork, destinationHost,
	)
}

/*
 * Acceptance tests.
 */

// Acceptance test for ddcloud_firewall_rule (IPv4 from any address to any address):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleIPv4FromAnyToAnyCreate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.ipv4.any.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		IP().
		IPv4().
		PlaceFirst().
		MatchAnySourceAddress().
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationPort().
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleIPFromHostToHost(
					"acc.test.firewall.rule.ipv4.any.to.any",
					compute.FirewallRuleIPVersion4,
					compute.FirewallRuleMatchAny,
					compute.FirewallRuleMatchAny,
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (IPv4 from any address to any address):
//
// Create a firewall rule, update it, and verify that it gets updated in-place with the correct configuration.
func TestAccFirewallRuleIPv4FromAnyToAnyUpdate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.ipv4.any.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		IP().
		IPv4().
		PlaceFirst().
		MatchAnySourceAddress().
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationPort().
		Enable()

	testAccResourceUpdateInPlace(t, testAccResourceUpdate{
		ResourceName: "ddcloud_firewall_rule.acc_test_rule",
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),

		// Create
		InitialConfig: testAccDDCloudFirewallRuleIPFromHostToHost(
			"acc.test.firewall.rule.ipv4.any.to.any",
			compute.FirewallRuleIPVersion4,
			compute.FirewallRuleMatchAny,
			compute.FirewallRuleMatchAny,
			true, // Enabled
		),
		InitialCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
			testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
				expectedRuleConfiguration.ToFirewallRule(),
			),
		),

		// Update
		UpdateConfig: testAccDDCloudFirewallRuleIPFromHostToHost(
			"acc.test.firewall.rule.ipv4.any.to.any",
			compute.FirewallRuleIPVersion4,
			compute.FirewallRuleMatchAny,
			compute.FirewallRuleMatchAny,
			false, // Disabled
		),
		UpdateCheck: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
			testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
				expectedRuleConfiguration.Disable().ToFirewallRule(),
			),
		),
	})
}

// Acceptance test for ddcloud_firewall_rule (IPv4 from network to any address):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleIPv4FromNetworkToAnyCreate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.ipv4.network.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		IP().
		IPv4().
		PlaceFirst().
		MatchSourceNetwork("8.8.8.0", 24).
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationPort().
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleIPFromNetworkToHost(
					"acc.test.firewall.rule.ipv4.network.to.any",
					compute.FirewallRuleIPVersion4,
					"8.8.8.0/24",
					compute.FirewallRuleMatchAny,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (IPv6 from test address to any address):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleIPv6FromTestToAnyCreate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.ipv6.test.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		IP().
		IPv6().
		PlaceFirst().
		MatchSourceAddress(testIPv6Address).
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationPort().
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleIPFromHostToHost(
					"acc.test.firewall.rule.ipv6.test.to.any",
					compute.FirewallRuleIPVersion6,
					testIPv6Address,
					compute.FirewallRuleMatchAny,
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (TCP over IPv4 from any address to any address on port 80):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleTCP4FromAnyToAnyPort80Create(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.tcp4.80.any.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		TCP().
		IPv4().
		PlaceFirst().
		MatchAnySourceAddress().
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchDestinationPort(80).
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleTCPFromHostToHost(
					"acc.test.firewall.rule.tcp4.80.any.to.any",
					compute.FirewallRuleIPVersion4,
					compute.FirewallRuleMatchAny,
					compute.FirewallRuleMatchAny,
					80,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (TCP over IPv6 from test address to any address on port 80):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleTCP6FromTestToAnyPort80Create(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.tcp6.80.test.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		TCP().
		IPv6().
		PlaceFirst().
		MatchSourceAddress(testIPv6Address).
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchDestinationPort(80).
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleTCPFromHostToHost(
					"acc.test.firewall.rule.tcp6.80.test.to.any",
					compute.FirewallRuleIPVersion6,
					testIPv6Address,
					compute.FirewallRuleMatchAny,
					80,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (ICMPv4 from any address to any address):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleICMP4FromAnyToAnyCreate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.icmp4.any.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		ICMP().
		IPv4().
		PlaceFirst().
		MatchAnySourceAddress().
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationAddress().
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleICMPFromHostToHost(
					"acc.test.firewall.rule.icmp4.any.to.any",
					compute.FirewallRuleIPVersion4,
					compute.FirewallRuleMatchAny,
					compute.FirewallRuleMatchAny,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (ICMPv4 from a network to any address):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleICMP4FromNetworkToAnyCreate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.icmp4.network.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		ICMP().
		IPv4().
		PlaceFirst().
		MatchSourceNetwork("8.8.8.0", 24).
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationPort().
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleICMPFromNetworkToHost(
					"acc.test.firewall.rule.icmp4.network.to.any",
					compute.FirewallRuleIPVersion4,
					"8.8.8.0/24",
					compute.FirewallRuleMatchAny,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

// Acceptance test for ddcloud_firewall_rule (TCP over IPv6 from test address to any address on port 80):
//
// Create a firewall rule and verify that it gets created with the correct configuration.
func TestAccFirewallRuleICMP6FromTestToAnyCreate(t *testing.T) {
	expectedRuleConfiguration := &compute.FirewallRuleConfiguration{
		Name: "acc.test.firewall.rule.icmp6.test.to.any",
	}
	expectedRuleConfiguration.
		Accept().
		ICMP().
		IPv6().
		PlaceFirst().
		MatchSourceAddress(testIPv6Address).
		MatchAnySourcePort().
		MatchAnyDestinationAddress().
		MatchAnyDestinationPort().
		Enable()

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckDDCloudFirewallRuleDestroy,
			testCheckDDCloudNetworkDomainDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDDCloudFirewallRuleICMPFromHostToHost(
					"acc.test.firewall.rule.icmp6.test.to.any",
					compute.FirewallRuleIPVersion6,
					testIPv6Address,
					compute.FirewallRuleMatchAny,
				),
				Check: resource.ComposeTestCheckFunc(
					testCheckDDCloudFirewallRuleExists("ddcloud_firewall_rule.acc_test_rule", true),
					testCheckDDCloudFirewallRuleMatches("ddcloud_firewall_rule.acc_test_rule",
						expectedRuleConfiguration.ToFirewallRule(),
					),
				),
			},
		},
	})
}

/*
 * Acceptance-test checks.
 */

// Acceptance test check for ddcloud_firewall_rule:
//
// Check if the firewall rule exists.
func testCheckDDCloudFirewallRuleExists(name string, exists bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		firewallRuleID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		firewallRule, err := client.GetFirewallRule(firewallRuleID)
		if err != nil {
			return fmt.Errorf("Bad: Get firewall rule: %s", err)
		}
		if exists && firewallRule == nil {
			return fmt.Errorf("Bad: Firewall rule not found with Id '%s'.", firewallRuleID)
		} else if !exists && firewallRule != nil {
			return fmt.Errorf("Bad: Firewall rule still exists with Id '%s'.", firewallRuleID)
		}

		return nil
	}
}

// Acceptance test check for ddcloud_firewall_rule:
//
// Check if the firewall rule's configuration matches the expected configuration.
func testCheckDDCloudFirewallRuleMatches(name string, expected compute.FirewallRule) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		firewallRuleID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		firewallRule, err := client.GetFirewallRule(firewallRuleID)
		if err != nil {
			return fmt.Errorf("Bad: Get firewall rule: %s", err)
		}
		if firewallRule == nil {
			return fmt.Errorf("Bad: Firewall rule not found with Id '%s'", firewallRuleID)
		}

		if firewallRule.Name != expected.Name {
			return fmt.Errorf("Bad: Firewall rule '%s' has name '%s' (expected '%s')", firewallRuleID, firewallRule.Name, expected.Name)
		}

		if firewallRule.Action != expected.Action {
			return fmt.Errorf("Bad: Firewall rule '%s' has action '%s' (expected '%s')", firewallRuleID, firewallRule.Action, expected.Action)
		}

		if firewallRule.Action != expected.Action {
			return fmt.Errorf("Bad: Firewall rule '%s' has enablement '%t' (expected '%t')", firewallRuleID, firewallRule.Enabled, expected.Enabled)
		}

		sourceDifferences := firewallRule.Source.Diff(expected.Source)
		if len(sourceDifferences) > 0 {
			return fmt.Errorf("Bad: Firewall rule '%s' has unexpected source scope: %s",
				firewallRuleID,
				strings.Join(sourceDifferences, ", "),
			)
		}

		destinationDifferences := firewallRule.Destination.Diff(expected.Destination)
		if len(destinationDifferences) > 0 {
			return fmt.Errorf("Bad: Firewall rule '%s' has unexpected destination scope: %s",
				firewallRuleID,
				strings.Join(destinationDifferences, ", "),
			)
		}

		return nil
	}
}

// Acceptance test resource-destruction check for ddcloud_firewall_rule:
//
// Check all firewall rules specified in the configuration have been destroyed.
func testCheckDDCloudFirewallRuleDestroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != "ddcloud_firewall_rule" {
			continue
		}

		firewallRuleID := res.Primary.ID

		client := testAccProvider.Meta().(*providerState).Client()
		firewallRule, err := client.GetFirewallRule(firewallRuleID)
		if err != nil {
			return nil
		}
		if firewallRule != nil {
			return fmt.Errorf("Firewall rule '%s' still exists.", firewallRuleID)
		}
	}

	return nil
}
