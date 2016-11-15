# ddcloud\_vlan

A virtual LAN (VLAN) is a partitioned and isolated broadcast domain within a Managed Cloud Platform network domain.

**Note:** Due to current platform limitations, organisations that use MCP 2.0 cannot perform more than one concurrent deployment operation for network domains or VLANs (all other operations can however be performed concurrently). If necessary, use the `depends_on` attribute to ensure that resources that relate to the same VLAN are not run in parallel.

## Example Usage

```
resource "ddcloud_vlan" "my-vlan" {
    name                    = "terraform-test-vlan"
    description             = "This is my Terraform test VLAN."

	# The Id of the network domain in which the VLAN will be deployed.
    networkdomain           = "${ddcloud_networkdomain.my-domain.id}"

    # VLAN's default network: 192.168.17.0/24 = 192.168.17.1 -> 192.168.17.254 (netmask = 255.255.255.0)
    ipv4_base_address       = "192.168.17.0"
    ipv4_prefix_size        = 24
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the VLAN.
* `description` - (Optional) A description for the VLAN.
* `networkdomain` - (Required) The Id of the network domain in which the VLAN is deployed.
* `ipv4_base_address` - (Required) The base address of the VLAN's IPv4 network.
* `ipv4_prefix_size` - (Required) The prefix size of the VLAN's IPv4 network.

## Attribute Reference

The following additional attributes are exported:

* `ipv6_base_address` - The base address of the VLAN's IPv6 network.
* `ipv6_prefix_size` - The prefix size of the VLAN's IPv6 network.
