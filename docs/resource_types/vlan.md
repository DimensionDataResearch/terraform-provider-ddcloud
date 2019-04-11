# ddcloud\_vlan

A virtual LAN (VLAN) is a partitioned and isolated broadcast domain within a Managed Cloud Platform network domain.

**Note:** Due to current platform limitations, organisations that use MCP 2.0 cannot perform more than one concurrent deployment operation for network domains or VLANs (all other operations can however be performed concurrently). If necessary, use the `depends_on` attribute to ensure that resources that relate to the same VLAN are not run in parallel.

## Example Usage

```
resource "ddcloud_vlan" "my-vlan-attached" {
    name                    = "terraform-test-attached-vlan"
    description             = "This is my Terraform test VLAN attached."
    networkdomain           = "${data.ddcloud_networkdomain.my-domain.id}"
    ipv4_base_address       = "172.16.0.0"
    ipv4_prefix_size        = 16
    
    ## attached_vlan_gateway_addressing and detached_vlan_gateway_address are mutually exclusive
    # Valid input either "HIGH" or "LOW"
    attached_vlan_gateway_addressing = "HIGH"
}

resource "ddcloud_vlan" "my-vlan-detached" {
    name                    = "terraform-test-detached-vlan"
    description             = "This is my Terraform test VLAN detached."
    networkdomain           = "${data.ddcloud_networkdomain.my-domain.id}"
    ipv4_base_address       = "10.1.1.0"
    ipv4_prefix_size        = 28
    
    ## attached_vlan_gateway_addressing and detached_vlan_gateway_address are mutually exclusive
    detached_vlan_gateway_address = "10.0.0.1"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the VLAN.
* `description` - (Optional) A description for the VLAN.
* `networkdomain` - (Required) The Id of the network domain in which the VLAN is deployed.
* `ipv4_base_address` - (Required) The base address of the VLAN's IPv4 network.
* `ipv4_prefix_size` - (Required) The prefix size of the VLAN's IPv4 network.

** (Either `detached_vlan_gateway_address` or `attached_vlan_gateway_addressing` must be specified. They are mutually exclusive) 

* `detached_vlan_gateway_address` - (Optional) The system will use this IP address as the IPv4 gateway when the Deploy Server API is used referencing a NIC to the VLAN.
* `attached_vlan_gateway_addressing` - (Optional) LOW gatewayAddressing has small VLAM with IP addresses x.x.x.1-x.x.x.3 reserved at the bottom of the VLAN range. HIGH gatewayAddressing has VLAN with IP addresses x.x.x.252-x.x.x.254 reserved at the top of the VLAN range.

Available Options include.
* `tag` - (Optional) A set of tags to apply to the vlan.
    * `name` - (Required) The tag name. **Note**: The tag name must already be defined for your organisation.
    * `value` - (Required) The tag value.

## Attribute Reference

The following additional attributes are exported:

* `ipv6_base_address` - The base address of the VLAN's IPv6 network.
* `ipv6_prefix_size` - The prefix size of the VLAN's IPv6 network.

## Import

Once declared in configuration, `ddcloud_vlan` instances can be imported using their Id.

For example:

```bash
$ terraform import ddcloud_vlan.my-vlan 1b7e3191-bac2-4efb-9c9e-87fdd7f86ded
```
