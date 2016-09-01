# ddcloud\_vip\_pool\_member

A VIP Pool Member links VIP Node (and, optionally, a port) with a VIP Pool. A VIP Node can be a member of multiple VIP Pools (and / or the same VIP Pool on multiple ports).

VIP Nodes and VIP Pools are only supported in Network Domains on the `ADVANCED` plan.

## Example Usage

```
resource "ddcloud_vip_pool_member" "test_pool_test_node" {
	pool 					= "${ddcloud_vip_pool.test_pool.id}"
	node 					= "${ddcloud_vip_node.test_node.id}"
	port 					= 80
	status 					= "ENABLED"
}
```

## Argument Reference

The following arguments are supported:

* `pool` - (Required) The Id of the target VIP pool.
* `node` - (Required) The Id of the target VIP node.
* `port` - (Optional) Port on which the node is added to the pool.
* `status` - (Optional) The status assigned to the VIP node, as a member of the VIP pool.  
  Must be one of:
	* `ENABLED` - the pool member is enabled, and will receive requests.
	* `DISABLED` - the pool member is disabled, and will not receive requests.
	* `FORCED_OFFLINE` - the pool member has encountered an error, and has been forced offline. It will not receive requests.

## Attribute Reference

There are currently no additional attributes for `ddcloud_vip_pool_member`.
