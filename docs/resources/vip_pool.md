# ddcloud\_vip\_pool

A VIP Pool is a fundamental component of a VIP used to group [VIP Nodes](vip_node.md) so that they can be addressed through a [Virtual Listener](virtual_listener.md) according to the rules defined by the Virtual Listener and VIP Pool.

A Pool can be added to a Virtual Listener as a `pool` or a `clientClonePool` where `pool` is the primary recipient of the traffic and the `clientClonePool` receives a clone of the same traffic for additional processing such as intrusion detection.

** Note: the current release of the `ddcloud` provider does not support `clientClonePool` yet (but the next release will).

VIP Pools are only supported in Network Domains on the `ADVANCED` plan.

## Example Usage

```
resource "ddcloud_vip_pool" "test_pool" {
	name					= "my_terraform_pool"
	description 			= "Adam's Terraform test VIP pool (do not delete)."
	load_balance_method		= "ROUND_ROBIN"
	service_down_action		= "NONE",
	slow_ramp_time			= 5,

	networkdomain 			= "${ddcloud_networkdomain.test_net_domain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the VIP pool. **Note**: Changing this value will cause the pool to be destroyed and re-created.
* `description` - (Optional) A description of the VIP pool.
* `networkdomain` - (Required) The Id of the network domain in which the VIP pool is created.
* The load-balancing method used by the VIP pool.  
  Must be one of:
	* `ROUND_ROBIN` - requests will be directed to each VIP node, in turn.
	* `LEAST_CONNECTIONS_NODE` - requests will be directed to the pool node that has the smallest number of active connections at the moment of connection.  
	  All connections to the pool are considered.
	* `LEAST_CONNECTIONS_MEMBER` - requests will be directed to the pool node that has the smallest number of active connections at the moment of connection.  
	  Only connections to the node as a member of the current pool are considered.
	* `OBSERVED_NODE` - requests will be directed to the pool node that has the smallest number of active connections over time.  
	  All connections to the node are considered.
	* `OBSERVED_MEMBER` - requests will be directed to the pool node that has the smallest number of active connections over time.  
	  Only connections to the node as a member of the current pool are considered.
	* `PREDICTIVE_NODE` - requests will be directed to the pool node that is predicted to have the smallest number of active connections.  
	  All connections to the pool are considered.
	* `PREDICTIVE_MEMBER` - requests will be directed to the pool node that is predicted to have the smallest number of active connections over time.  
	  Only connections to the pool as a member of the current pool are considered.
* `health_monitors` (Optional) The names of any health monitors used by the pool.
   Must be one or more of the following:
	 * `CCDEFAULT.Http`
	 * `CCDEFAULT.Https`
	 * `CCDEFAULT.Tcp`
	 * `CCDEFAULT.TcpHalfOpen`
	 * `CCDEFAULT.Udp`
* `service_down_action` (Optional) The action to take when the service on a node is unavailable. Must be one of:
	* `NONE` - (Default) no action will be taken.
	* `DROP` - the node will be dropped from service.
	* `RESELECT` - the node will be reselected.
* `slow_ramp_time` (Required) The period of time, in seconds, over which requests to new nodes are ramped up to the full rate (allows each node to gradually warm up).

## Attribute Reference

There are currently no additional attributes for `ddcloud_vip_pool`.
