# ddcloud\_server\_nic

Represents an additional network adapters in an existing server.

## Example Usage

```
resource "ddcloud_server_nic" "server_nic_test" {
  server       = "${ddcloud_server.test_server.id}"
  private_ipv4 = "192.168.18.100"
  vlan         = "${ddcloud_vlan.test_vlan.id}"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) ID of the server to which the network adapter needs to be added
* `private_ipv4` - (Optional) IPv4 Address for the new network adapter.  
At least one of `private_ipv4` or `vlan` must be specified.
* `vlan` - (Optional) VLAN ID of the new network adapter.  
At least one of `private_ipv4` or `vlan` must be specified.  
`primary_adapter_vlan` is ignored if `primary_adapter_ipv4` is also specified.  
It's still useful to supply both, though, since it sets up a dependency between the NIC and the VLAN.
* `adapter_type` - (Optional) The type of network adapter (`E1000` or `VMXNET3`).  
**Note**: Changing this property will result in the adapter being destroyed and re-created.

## Attribute Reference

There are currently no additional attributes for `ddcloud_server_nic`.
