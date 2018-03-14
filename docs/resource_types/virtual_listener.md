# ddcloud\_virtual\_listener

Virtual Listener is the top level component of a VIP.

It is used to expose the underlying Pool(s) to external network traffic via the address indicated in its `ipv4` property.

VIP Pools are only supported in Network Domains on the `ADVANCED` plan.

## Example Usage

```
resource "ddcloud_virtual_listener" "test_virtual_listener" {
	name                 	= "my_terraform_listener"
	type									= "STANDARD"
	protocol             	= "HTTP"
	optimization_profile 	= "TCP"
	pool                 	= "${ddcloud_vip_pool.test_pool.id}"
	ipv4                	= "192.168.18.10"

	networkdomain 		 	= "${ddcloud_networkdomain.test_domain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the virtual listener.  
  **Note**: Changing this value will cause the listener to be destroyed and re-created.
* `description` - (Optional) A description of the virtual listener.
* `type` - (Optional) The listener type.  
  Must be one of:
	* `STANDARD`
	* `PERFORMANCE_LAYER_4`
* `protocol` - (Required) The protocol to be supported by the listener.  
	* If `type` is `STANDARD`:
		* `ANY`
		* `TCP`
		* `UDP`
		* `HTTP`
		* `FTP`
		* `SMTP`
	* If `type` is `PERFORMANCE_LAYER_4`:
		* `ANY`
		* `TCP`
		* `UDP`
		* `HTTP`
* `pool` - (Optional) The Id of the underlying VIP pool to which the listener forwards traffic.
* `ipv4` - (Optional) The IPv4 address from which the listener will accept traffic.  
  The address can be either:
	* Public  
	  `ipv4` is optional; if not specified, the first free public IPv4 address will be used. Will fail if there are no available public IPv4 addresses.
	* Private  
	  `ipv4` is required, and must be neither already be in use by a Node on the Network Domain nor fall within the IP space of a VLAN deployed on the Network Domain.
* `port` - (Optional)
* `enabled` - (Optional)
* `ssl_offload_profile` - (Optional) The Id of an SSL-offload profile (if any) to assign to the virtual listener.
* `connection_limit` (Optional) - The listener total connection limit.
* `connection_rate_limit` (Optional) - The listener connection rate limit.
* `source_port_preservation` (Optional) - Preserve source port information (if possible)?
* `persistence_profile` (Optional) - The name of the persistence profile (if any) to use.
* `irules`
* `optimization_profile` (Optional) - The listener optimisation profile.  
  Required if `type` is `STANDARD` and `protocol` is `TCP`.  
	See the CloudControl documentation for further information.
* `networkdomain` - (Required) The Id of the network domain in which the VIP pool is created.

## Attribute Reference

There are currently no additional attributes for `ddcloud_virtual_listener`.
