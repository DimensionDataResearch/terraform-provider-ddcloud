# ddcloud\_vip\_node

A VIP Node is a low level component of a VIP and a NAT in an MCP 2.0 data centre.

A VIP Node is added to a VIP Pool as a VIP Pool Member, where it can receive designated network traffic from a Virtual Listener.

VIP Nodes are only supported in Network Domains on the `ADVANCED` plan.

## Example Usage

```
resource "ddcloud_vip_node" "test_node" {
	name					= "my_terraform_vip_node"
	description 			= "My Terraform test VIP node."
    networkdomain 			= "${ddcloud_networkdomain.test_net_domain.id}"

	ipv4_address			= "192.168.17.10"
	status					= "ENABLED"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the VIP node. **Note**: Changing this value will cause the node to be destroyed and re-created.
* `description` - (Optional) A description of the VIP node.
* `networkdomain` - (Required) The Id of the network domain in which the VIP node is created.
* `ipv4_address` - (Required) The VIP node's IPv4 address. Exactly one of `ipv4_address` or `ipv6_address` must be specified.
* `ipv6_address` - (Required) The VIP node's IPv6 address. Exactly one of `ipv4_address` or `ipv6_address` must be specified.
* `health_monitor` - (Optional) The name of the VIP node's associated health monitor (By Default 'CCDEFAULT.Icmp' health monitor will be associated).
* `connection_limit` - (Optional) The number of active connections that the node supports.
* `connection_rate_limit` - (Optional) The number of new connections per second that the node supports.
* `status` - (Required) The VIP node status. Must be one of `ENABLED`, `DISABLED`, or `FORCED_OFFLINE`.

## Attribute Reference

There are currently no additional attributes for `ddcloud_vip_node`.
