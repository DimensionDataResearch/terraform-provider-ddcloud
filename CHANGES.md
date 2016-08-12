# Changes

## v0.7

New features:

* Full support for load-balancer configuration:
  * `ddcloud_vip_node`
  * `ddcloud_vip_pool`
  * `ddcloud_vip_pool_member`
  * `ddcloud_virtual_listener`

Fixes:

* `ddcloud_nat` resource is now fully idempotent (previously, if the NAT rule went missing, then Terraform did not correctly detect this).

## v0.6

* Extended logging of CloudControl API requests and responses can now be enabled by setting the `DD_COMPUTE_EXTENDED_LOGGING` environment variable to any non-empty value.

## v.04

Fixes:

* Changes to `ddcloud_server.tag` and `ddcloud_server.disk` are now correctly detected.

New features:

* `ddcloud_server` can now be deployed from a customer image (use `customer_image_id` / `customer_image_name` instead of `os_image_id`, `os_image_name`).

Breaking changes:

* `ddcloud_server.osimage_id` and `ddcloud_server.osimage_name` have been renamed to `ddcloud_server.os_image_id` and `ddcloud_server.os_image_id_name` (this is to be consistent with `customer_image_id` and `customer_image_name`).

## v0.3

* `ddcloud_server` now has a `public_ipv4` attribute that is resolved by matching any NAT rule that targets the server's primary network adapter's private IPv4 address. If the NAT rule gets the private IPv4 address from the server (rather than the other way around) then this attribute will not be available until you run `terraform refresh`.
