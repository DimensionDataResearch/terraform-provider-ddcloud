# Changes

### v3.0.0
Breaking Changes:
* Upgrade to terraform 0.12

### v1.3.8

* Handle missing server when checking for existence of server network adapter (DimensionDataResearch/dd-cloud-compute-terraform#114).

### v1.3.7

* Bug-fix: create / update of `ddcloud_server` resource fails if server is running and disk configuration needs to be updated (DimensionDataResearch/dd-cloud-compute-terraform#116).

### v1.3.6

* Bug-fix: "invalid UTF-8" error if user account details contain non-ASCI characters.

### v1.3.5

* If the initial connection to the CloudControl API returns an invalid response, this response is now always written to the log (this helps when troubleshooting interaction with CloudControl).

### v1.3.4

* Implement `ddcloud_server_backup` resource type (DimensionDataResearch/dd-cloud-compute-terraform#26).
* Expose downoad URLs for backup clients on computed `backup_client_urls` attribute of `ddcloud_server` resource type  (DimensionDataResearch/dd-cloud-compute-terraform#26).

### v1.3.4-preview1

* Persist private key for `ddcloud_ssl_domain_certificate` in state data.  
  Otherwise, Terraform either supplies an incorrect value to the provider (which causes creation to fail) or always sees the key as having changed (which causes a destroy-and-recreate of the resource).

### v1.3.3

* Defer `ddcloud_server_backup` until v1.3.4 (final).
* Bug-fix: `ssl_offload_profile` property is not being persisted by `ddcloud_virtual_listener` (DimensionDataResearch/dd-cloud-compute-terraform#110).
* Use v2.6 API for ddcloud_virtual_listener.
* Ignore changes to `private_key` property of `ddcloud_ssl_domain_certificate`.  
  This is required because the private key is not persisted in state data.

### v1.3.3-preview1

* Implement `ddcloud_server_backup` resource type (DimensionDataResearch/dd-cloud-compute-terraform#26).
* Expose downoad URLs for backup clients on computed `backup_client_urls` attribute of `ddcloud_server` resource type  (DimensionDataResearch/dd-cloud-compute-terraform#26).

### v1.3.2

* Expose SSL-offload profile on `ddcloud_virtual_listener` (DimensionDataResearch/dd-cloud-compute-terraform#104).
* Implement `ddcloud_ssl_offload_profile` resource type (DimensionDataResearch/dd-cloud-compute-terraform#104).
* Implement `ddcloud_ssl_domain_certificate` resource type (DimensionDataResearch/dd-cloud-compute-terraform#104).
* Implement `ddcloud_ssl_certificate_chain` resource type (DimensionDataResearch/dd-cloud-compute-terraform#104).
* Implement `ddcloud_pfx` data source (DimensionDataResearch/dd-cloud-compute-terraform#104).

### v1.3.0

* Switch to Terraform v0.9.11.
* Implement resource import for `ddcloud_networkdomain`  (DimensionDataResearch/dd-cloud-compute-terraform#73).
* Implement resource import for `ddcloud_vlan` (DimensionDataResearch/dd-cloud-compute-terraform#90).
* Implement resource import for `ddcloud_server` (DimensionDataResearch/dd-cloud-compute-terraform#91).
* Implement resource import for `ddcloud_nat` (DimensionDataResearch/dd-cloud-compute-terraform#92).
* Implement resource import for `ddcloud_firewall_rule` (DimensionDataResearch/dd-cloud-compute-terraform/issues/93).

### v1.3.0-beta1

* Add support for deploying `ddcloud_server` from an uncustomised image.

### v1.3.0-alpha3

* `ddcloud_networkdomain` and `ddcloud_vlan` datasources now correctly return an error if the target entity is not found.

### v1.3.0-alpha2

* Support for multiple storage controllers (using `ddcloud_storage_controller` resource).

### v1.3.0-alpha1

* Support for additional disk storage tier (`ECONOMY`) and network adapter types (`E1000E`, `ENHANCED_VMXNET2`, and `FLEXIBLE_PCNET32`).

### v1.3.0-preview2

* Improvements to behaviour of `ddcloud_network_adapter` (matching up resources with corresponding server network adapters by Id).

### v1.3.0-preview1

## Breaking changes

* Switch to Terraform v0.8.2 (_only_ works with v0.8+ now).

### v1.2.0

Final (RTM) release of v1.2.0.

### v1.2.0-rc2

* Migrate properties for `ddcloud_server.primary_network_adapter`.

### v1.2.0-rc1

* Fix broken acceptance tests for `ddcloud_server_anti_affinity`.

### v1.2.0-beta3

Enhancements:

* Add new (experimental) `ddcloud_ip_address_reservation` resource type.  
This resource type allows specific IP addresses to be reserved on VLANs. It is only intended to support advanced usage scenarios.

### v1.2.0-beta2

Enhancements:

* Enable changing of adapter type for existing network adapters, where supported (#80).

#### v1.2.0-beta1

Enhancements:

* Provider now checks for disks with duplicate SCSI unit Ids before deployment (#63).

#### v1.2.0-alpha4

Bug fixes:

* Fix crash when deploying a customer image using `image_type="auto"` (#66).

#### v1.2.0-alpha3

Bug fixes:

* Fix INVALID\_INPUT\_DATA when creating a ddcloud\_address\_list using complex addresses (#72).

#### v1.2.0-alpha2

* Simplify image properties for `ddcloud_server` (#66).  
There are now just 2 image-related properties: `image` (image name or Id), and `image_type`.  
See the provider documentation for details.
* Change simple ports in `ddcloud_port_list` from strings to integers (#71).

#### v1.2.0-alpha1

* Fix crash when deploying `ddcloud_server` from a customer image (#66).

#### v1.2-preview5

This is a preview release intended to gather feedback on changes to `ddcloud_server.primary_network_adapter`, `ddcloud_server.additional_network_adapter`, and `ddcloud_network_adapter`.

Enhancements:

* MAC address is now exposed on network adapters.

Breaking changes:

* Primary network adapter is now exposed via `ddcloud_server.primary_network_adapter` (#56).
* Additional network adapters are now exposed via `ddcloud_server.additional_network_adapter` and `ddcloud_network_adapter` (#56).  
**Note**: You can specify additional adapters for a given `ddcloud_server` via either the `ddcloud_server.additional_network_adapter` _or_ `ddcloud_network_adapter` (but not both). Use `ddcloud_network_adapter` if you want to be able to modify additional network adapters after deployment.

#### v1.2-preview4

This is a preview release intended to gather feedback on changes to `ddcloud_server.image` and `ddcloud_server.admin_password`.

Enhancements:

* The provider will now attempt to automatically migrate the data for `ddcloud_server.disk` from a Set to a List.  
**Note** - if you have not used v1.2-preview3, then you can now disregard the warning about `terraform.tfstate` incompatibility. 
* The image for a `ddcloud_server` is now configured via `ddcloud_server.image` (#66).  
You now only need to set 1 of 2 properties (and there's an optional 3rd property if you want to customise behaviour).
  * `id` - The Id of the target image to use.
  * `name` - The name of the target image to use.
  * `type` - The type of image to use (optional, defaults to `auto`).
    * `os` - Use an OS image with the specified Id or name.
    * `customer` - Use an OS image with the specified Id or name.
    * `auto` - Use an OS or customer image with the specified Id or name (auto-detect image type).  
    Note that if an OS and customer image have the same name or Id, the OS image will be used.
* `ddcloud_server.admin_password` is now optional for server images that don't require an initial administrator password (#65).

Breaking changes:

* `ddcloud_server` properties relating to image have been moved into `ddcloud_server.image` (#66).

Bug fixes:

* Fixed crash when deploying a `ddcloud_server`.

#### v1.2-preview3

This is a preview release intended to gather feedback on changes to `ddcloud_server.disk` and `ddcloud_server._network_adapter` behaviour.

Breaking changes:

* Network adapters now modelled as `ddcloud_server.network_adapter` (see documentation on `ddcloud_server` resource type for details).
* Change `ddcloud_server.disk` from a set to a list in order to fix errors when performing subsequent `terraform apply` that modifies an existing disk (#63).  
**Note** - this will break any existing `terraform.tfstate` values you have for `ddcloud_server` resources.

#### v1.2-preview2

This is a preview release intended to gather feedback on changes to `ddcloud_server.disk` and `ddcloud_server._network_adapter` behaviour.

Bug fixes:

* Fix some major bugs due to `ddcloud_server.disk` changes (#63).

Enhancements:

* Network adapters are added to server as part of initial deployment rather than having to wait until after deployment to add them (#56).
* Disk speed is now applied to disks that are part of the initial deployment.

#### v1.2-preview1

This is a preview release intended to gather feedback on changes to `ddcloud_server.disk` behaviour.

Enhancements:

* Major improvements (robustness and maintainability) to add / modify server disks (#63).  
This logic was getting waaaay too complicated and is now greatly simplified using lessons learned over the last couple of months.

Bug fixes:

* Removal of a `ddcloud_server` disk now actually removes the disk (#63).

Breaking changes:

* `ddcloud_server_nic` resource has been removed (#56).  
Its functionality will be merged back into `ddcloud_server` before v1.2 is released.  
This is mainly due to the size and scope of the changes required to do this. Now that we've figured out what's involved in simplifying the `disk` work for `ddcloud_server` is complete, the `network_adapter` work should be a lot easier.

#### v1.1.9

New features:

* Automatically retry network operations that fail due to `RESOURCE_BUSY` response from CloudControl (#11)
* Use default DNS when deploying a server if no specific DNS IPs are configured (#62) 

Bug fixes:

* Fix crashes when customer_image_id or customer_image_name is used with ddcloud_server (#64)
* Fix `UNEXPECTED_ERROR` responses from CloudControl due to simultaneously initiating multiple asynchronous operations  
Multiple asynchronous operations can run in parallel, but only 1 can be initiated at a time

Breaking changes:

* `ddcloud_server` properties `primary_adapter_vlan` and `primary_adapter_ipv4` can be specified together but if `primary_adapter_ipv4` is specified, then `primary_adapter_vlan` is ignored (this is OK because specifying one implies the other)
* `ddcloud_server_nic` properties `vlan` and `private_ipv4` can be specified together but if `primary_adapter_ipv4` is specified, then `primary_adapter_vlan` is ignored (this is OK because specifying one implies the other)
* `retry_count` and `retry_delay` provider properties have been removed (use the `MCP_MAX_RETRY` and `MCP_RETRY_DELAY` environment variables)  
`retry_delay` is now used to control retry of operations that fail due to `RESOURCE_BUSY` response from CloudControl

#### v1.1.8

Breaking changes:

* `ddcloud_server` properties `primary_adapter_vlan` and `primary_adapter_ipv4` are now mutually exclusive (this is OK because specifying one implies the other)
* `ddcloud_server_nic` properties `vlan` and `private_ipv4` are now mutually exclusive (this is OK because specifying one implies the other)

Bug fixes:

* `ddcloud_server_nic` now respects per-server locks (so the error message about servers already being rebooted should no longer occur).
* `ddcloud_server_nic` now correctly applies its `adapter_type` property

#### v1.1.7

New features:

* Add the ability to manage default firewall rules (#54)

#### v1.1.6

Bug fixes:

* Settings such as `allow_server_reboots` are now correctly honoured by the provider (previously, they were incorrectly overridden to `false`).
* Correctly apply partial state when deploying a new `ddcloud_server`
* Removed `auto_create_tag_keys` setting and the ability to automatically create tag keys (#53)  
CloudControl support for this feature is too fragile (e.g. only works in home region), and considering how infrequently it's used it's not worth the effort.

#### v1.1.5

New features:

* Enable specifying the network adapter type on `ddcloud_server` and `ddcloud_server_nic` (#52).

#### v1.1.4

New features:

* `ddcloud_port_list` now also supports specifying a list of simple ports (#50).
* Build executable for 32-bit Windows (Go calls this windows-386).

#### v1.1.3

New features:

* `ddcloud_address_list` now also supports specifying a list of simple IP addresses (#49).
* Enable specifying of custom URL for CloudControl API end point (#48).

Bug fixes:

* Server name and description can now be modified after deployment (#47).

Changes:

* The `ddcloud` provider will now automatically retry requests to the CloudControl API if they fail due to network errors (default value for `retry_count` is now 3).

Breaking changes:

* Environment variables that previously had a `DDCLOUD_` prefix will now have an `MCP_` prefix (required for [docker-machine-driver-ddcloud #4](https://github.com/DimensionDataResearch/docker-machine-driver-ddcloud/issues/4))

#### v1.1.2

New features:

* The provider now logs its version information at startup (for diagnostic purposes).
* Add `ddcloud_vlan` data source (#45)

Bug fixes:

* #44: Order of child address or port lists is not preserved

Breaking changes:

* The `child_lists` property on `ddcloud_address_list` and `ddcloud_port_list` is now a Set instead of a list. If you have existing Terraform state for this property it will not be retained (sorry).

#### v1.1.1

New features:

* Add support for automatically creating tag keys if they are not already defined.
* Fixed bug #42 - unable to create `ddcloud_firewall_rule` with source address list or destination address list.  
This was due to inconsistencies in the CloudControl API for firewall rules (create vs read returns different field structure). The CloudControl client for Go has been updated and the new version included in the Terraform provider.

#### v1.1

New features:

* Implement address lists (`ddcloud_address_list` resource type).
* Implement port lists (`ddcloud_port_list` resource type).
* Add address list and port list support to `ddcloud_firewall_rule` resource type.
* Add support for additional server network adapters (`ddcloud_server_nic` resource type).
* Expose CPU speed and cores-per-socket on `ddcloud_server` resource type (`cpu_speed` and `cores_per_cpu` properties).

#### v1.0.2

New features:

* Implemented server anti-affinity rules (`ddcloud_server_anti_affinity` resource type).

#### v1.0.1

Fixes:

* Fix for incorrect behaviour when adding VIP pool members with a specific port ([#17](https://github.com/DimensionDataResearch/dd-cloud-compute-terraform/pull/17/))
* `ddcloud_virtual_listener`'s `ipv4` address property is now computable (and captured during create / read).
* If there are no available public IP addresses when creating a `ddcloud_virtual_listener` without explicitly specifying an IPv4 address for the listener (i.e. CloudControl will allocate an IPv4 address), the provider will now automatically allocate a public IP block (similar to the behaviour of `ddcloud_nat_rule`).

#### v0.7

New features:

* Full support for load-balancer configuration:
  * `ddcloud_vip_node`
  * `ddcloud_vip_pool`
  * `ddcloud_vip_pool_member`
  * `ddcloud_virtual_listener`

Fixes:

* `ddcloud_nat` resource is now fully idempotent (previously, if the NAT rule went missing, then Terraform did not correctly detect this).

#### v0.6

* Extended logging of CloudControl API requests and responses can now be enabled by setting the `MCP_EXTENDED_LOGGING` environment variable to any non-empty value.

#### v.04

Fixes:

* Changes to `ddcloud_server.tag` and `ddcloud_server.disk` are now correctly detected.

New features:

* `ddcloud_server` can now be deployed from a customer image (use `customer_image_id` / `customer_image_name` instead of `os_image_id`, `os_image_name`).

Breaking changes:

* `ddcloud_server.osimage_id` and `ddcloud_server.osimage_name` have been renamed to `ddcloud_server.os_image_id` and `ddcloud_server.os_image_id_name` (this is to be consistent with `customer_image_id` and `customer_image_name`).

#### v0.3

* `ddcloud_server` now has a `public_ipv4` attribute that is resolved by matching any NAT rule that targets the server's primary network adapter's private IPv4 address. If the NAT rule gets the private IPv4 address from the server (rather than the other way around) then this attribute will not be available until you run `terraform refresh`.
