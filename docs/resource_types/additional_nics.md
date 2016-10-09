# ddcloud\_additional\_nics

This resource to add additional network adapters to the existing server

## Example Usage

```
resource "ddcloud_additional_nics" "additional_nic_test" {
	server   					= "${ddcloud_server.test_server.id}"
	private_ipv4 			= "192.168.18.100"
  vlan_id 			= "${ddcloud_vlan.test_vlan.id}"
  shutdown_ok   = true
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) ID of the server to which the network adapter needs to be added
* `private_ipv4` - (Optional) IPv4 Address for the new network adapter. Exactly one of `private_ipv4` or `vlan_id` or both must be specified.
* `vlan_id` - (Optional) VLAN ID of the new network adapter. Exactly one of `private_ipv4` or `vlan_id` or both must be specified.
* `shutdown_ok` - (Required) Its to confirm that you are ok to shutdown the server to add/remove the nics. If it's false, then the network adapter will be added or removed only if the server is powered off

## Attribute Reference

There are currently no additional attributes for `ddcloud_additional_nics`.
