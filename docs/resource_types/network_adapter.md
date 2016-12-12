# ddcloud\_network\_adapter

Represents an additional network adapter in an existing server.

**Note**: Using both `ddcloud_network_adapter` _and_ `ddcloud_server.additional_network_adapter` for the same server is not supported.

## Example Usage

```
resource "ddcloud_network_adapter" "test_server_adapter2" {
  server       = "${ddcloud_server.test_server.id}"
  private_ipv4 = "192.168.18.100"
  vlan         = "${ddcloud_vlan.test_vlan.id}"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) ID of the server to which the network adapter needs to be added
* `ipv4` - (Optional) IPv4 Address for the new network adapter.  
At least one of `ipv4` or `vlan` must be specified.
* `vlan` - (Optional) VLAN ID of the new network adapter.  
At least one of `ipv4` or `vlan` must be specified.  
**Note**: Changing this property will result in the adapter being destroyed and re-created.
`vlan` is ignored if `ipv4` is also specified.  
It's still useful to supply both, though, since it sets up a dependency between the NIC and the VLAN.
* `type` - (Optional) The type of network adapter (`E1000` or `VMXNET3`).

## Attribute Reference

The following attributes are exposed:

* `mac` - The network adapter's MAC address.
