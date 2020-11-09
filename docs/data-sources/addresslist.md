# ddcloud\_addresslist

The `ddcloud_networklist` data-source enables lookup of a addresslist by name and network domain.
An IP Address List is a flexible object, which can be used on the Source or Destination of multiple Firewall Rules 
as a way of re-using network structures across Firewall Rules.


## Example Usage
Lookup an existing addresslist by name

```
data "ddcloud_addresslist" "my-addresslist" {
    name                    = "AddressList_A"
    networkdomain          = "${data.ddcloud_networkdomain.my-domain.id}"

}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the address list.
* `networkdomain` - (Required) The Id of the network in which the AddressList exists.

## Attribute Reference

The following attributes are exported:

* `description` - Additional notes (if any) for the address list.
* `ip_version` - Either IPv4 or IPv6.
