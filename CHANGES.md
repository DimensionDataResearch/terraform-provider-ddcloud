# Changes

## v.04

Fixes:

* Changes to `ddcloud_server.tag` and `ddcloud_server.disk` are now correctly detected.

## v0.3

* `ddcloud_server` now has a `public_ipv4` attribute that is resolved by matching any NAT rule that targets the server's primary network adapter's private IPv4 address. If the NAT rule gets the private IPv4 address from the server (rather than the other way around) then this attribute will not be available until you run `terraform refresh`.
