# Changes

## v.04

Fixes:

* Changes to `ddcloud_server.tag` and `ddcloud_server.disk` are now correctly detected.

New features:

* `ddcloud_server` can now be deployed from a customer image (use `customer_image_id` / `customer_image_name` instead of `os_image_id`, `os_image_name`).

Breaking changes:

* `ddcloud_server.osimage_id` and `ddcloud_server.osimage_name` have been renamed to `ddcloud_server.os_image_id` and `ddcloud_server.os_image_id_name` (this is to be consistent with `customer_image_id` and `customer_image_name`).

## v0.3

* `ddcloud_server` now has a `public_ipv4` attribute that is resolved by matching any NAT rule that targets the server's primary network adapter's private IPv4 address. If the NAT rule gets the private IPv4 address from the server (rather than the other way around) then this attribute will not be available until you run `terraform refresh`.
