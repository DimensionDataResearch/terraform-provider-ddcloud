# Changes

## v1.1.0

New features:

* Implemented address lists (`ddcloud_address_list` resource type).
* Implemented port lists (`ddcloud_port_list` resource type).
* Added address list and port list support to `ddcloud_firewall_rule` resource type.
* Added support for additional server network adapters (`ddcloud_server_nic` resource type).

## v1.0.2

New features:

* Implemented server anti-affinity rules (`ddcloud_server_anti_affinity` resource type).

## v1.0.1

Fixes:

* Fix for incorrect behaviour when adding VIP pool members with a specific port ([#17](https://github.com/DimensionDataResearch/dd-cloud-compute-terraform/pull/17/))
* `ddcloud_virtual_listener`'s `ipv4` address property is now computable (and captured during create / read).
* If there are no available public IP addresses when creating a `ddcloud_virtual_listener` without explicitly specifying an IPv4 address for the listener (i.e. CloudControl will allocate an IPv4 address), the provider will now automatically allocate a public IP block (similar to the behaviour of `ddcloud_nat_rule`).

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
