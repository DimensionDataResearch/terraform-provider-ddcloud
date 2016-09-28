# ddcloud\_anti_affinity

An anti-affinity rule ensures that 2 [Servers](server.md) are not run on the same physical hardware.

## Example Usage

```
resource "ddcloud_server_anti_affinity" "acc_test_anti_affinity_rule" {
	server1 = "${ddcloud_server.acc_test_server1.id}"
	server2 = "${ddcloud_server.acc_test_server2.id}"
}
```

## Argument Reference

The following arguments are supported:

* `server1` - (Required) The Id of the first server that the rule relates to.
* `server2` - (Required) The Id of the second server that the rule relates to.

## Attribute Reference

* `server1_name` - The name of the first server that the rule relates to.
* `server2_name` - The name of the second server that the rule relates to.
* `networkdomain` - The Id of the network domain in which the rule applies.
