# ddcloud\_ip\_address\_reservation 

Normally, CloudControl automatically assigns IP addresses when a server's network adapter(s) to a VLAN. When it assigns an address, it marks it as reserved so it cannot be reassigned.
For some (advanced) usage scenarios it is sometimes useful to manually reserve an IP address because it will be assigned through other means (e.g. DHCP, or as an additional IP address on a network adapter).

## Example Usage

### IPv6
The following configuration reserves an IPv6 address on a VLAN.

```hcl
resource "ddcloud_ip_address_reservation" "server1_6address2" {
  address      = "2001:44b8:31ab:b500:740e:425:dafa:896a"
  address_type = "ipv6"

  vlan         = "${ddcloud_vlan.test_vlan.id}"
}
```

### Private IPv4
The following configuration reserves a private IPv4 address on a VLAN.

```hcl
resource "ddcloud_ip_address_reservation" "server1_4address2" {
  address      = "10.19.21.12"
  address_type = "ipv4"
  description = "Reserved IP address for demo" 
  vlan         = "${ddcloud_vlan.test_vlan.id}"
}
```
  
## Argument Reference

The following arguments are supported:

* `address` - (Required) The IP address to reserve.
* `address_type` - (Required) The IP address type (IPv6 or private IPv4).  
Must be either `ipv4` or `ipv6`.
* `vlan` - (Required) The Id of the VLAN on which to reserve the IP address.
* `description` - The description for the Reserved IP address.

## Attribute Reference

There are currently no additional attributes for `ddcloud_ip_address_reservation`.
