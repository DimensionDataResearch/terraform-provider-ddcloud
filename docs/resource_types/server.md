# ddcloud\_server

A Server is a virtual machine. It is deployed in a network domain, and each of its network adapters are connected to a VLAN. Each server is created from an OS or customer image that also specifies sensible defaults for the server configuration.

## Example Usage

```hcl
resource "ddcloud_server" "myserver" {
  name                 = "terraform-server"
  description          = "My Terraform test server."
  admin_password       = "password"

  memory_gb            = 8
  cpu_count            = 2
  cpu_speed            = "STANDARD"
  cores_per_cpu        = 1

  image                = "CentOS 7 64-bit 2 CPU"

  networkdomain        = "${ddcloud_networkdomain.mydomain.id}"

  primary_network_adapter {
    vlan               = "${ddcloud_vlan.myvlan.id}"
    ipv4               = "192.168.17.10"
  }

  dns_primary          = "8.8.8.8"
  dns_secondary        = "8.8.4.4"

  disk {
      scsi_unit_id     = 0
      size_gb          = 10
      speed            = "STANDARD"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the server.
* `description` - (Optional) A description for the server.
* `admin_password` - (Optional) The initial administrative password for the deployed server.  
Has no effect after deployment.
  * Required for all OS images.
  * Required for Windows Server 2008 customer images.
  * Required for Windows Server 2012 customer images.
  * Required for Windows Server 2012 R2 customer images.
  * Optional for Linux customer images.
* `memory_gb` - (Optional) The amount of memory (in GB) allocated to the server.  
Defaults to the memory specified by the image from which the server is created.
* `cpu_count` - (Optional) The number of CPUs allocated to the server.  
Defaults to the CPU count specified by the image from which the server is created.
* `cores_per_cpu` - (Optional) The number of cores per virtual CPU socket allocated to the server.  
Defaults to the number of cores specified by the image from which the server is created.
* `cpu_speed` - (Optional) The speed of the CPU(s) allocated to the server (`STANDARD` or `HIGHPERFORMANCE`).  
Default is `STANDARD`.
* `image` - (Required) The name or Id of the image used to create the server.  
If `image` is a GUID / UUID, then it is treated as the image Id. Otherwise, it is treated as the image name.
* `image_type` - (Optional) The type of image used to create the server.  
If specified, must be `os`, `customer`, or `auto` (default). 
* `disk` - (Optional) The set of virtual disks attached to the server.  
  **Note**: If you list _any_ of the server's disks here, you must specify _all_ of its disks (including ones included in the original image).  
  Additionally, if your server has (or is likely to have) multiple storage controllers (i.e. SCSI buses) then you should define one or more [ddcloud\_storage\_controller](storage_controller.md) resources and declare your disks there instead.  
    * `scsi_unit_id` - (Required) The SCSI Logical Unit Number (LUN) for the disk. Must be unique across the server's disks.
    * `size_gb` - (Required) The size (in GB) of the disk. This value can be increased (to expand the disk) but not decreased.
    * `speed` - (Required) The disk speed. Usually one of `STANDARD`, `ECONOMY`, or `HIGHPERFORMANCE` (but varies between data centres).
* `networkdomain` - (Required) The Id of the network domain in which the server is deployed.
* `primary_network_adapter` - (Required) The primary network adapter attached to the server
  * `vlan` - (Optional) The Id of the VLAN that the primary network adapter is attached to.  
  Must specify at least one of `vlan` or `ipv4`.  
  **Note**: Changing this property will result in the server being destroyed and recreated.
  * `ipv4` - (Optional) The IPv4 address for the primary network adapter.  
  Note that if `ipv4` is specified, the VLAN will be inferred from this value.  
  Must specify at least one of `ipv4` or `vlan`.
  * `type` - (Optional) The primary network adapter type.  
  Must be either `E1000` (default) or `VMXNET3`.  
* `additional_network_adapter` - (Optional 0..\*) Additional network adapters (if any) attached to the server  
  **Note**: Changing this property will result in the server being destroyed and recreated.  
  If you want to support modifying of additional network adapters, use `ddcloud_network_adapter` resources instead.  
  **Note**: Using both `additional_network_adapter` _and_ the `ddcloud_network_adapter` resource type for the same server is not supported.
  * `vlan` - (Optional) The Id of the VLAN that the network adapter is attached to.  
  Must specify at least one of `vlan` or `ipv4`.
  * `ipv4` - (Optional) The IPv4 address for the network adapter.  
  Note that if `ipv4` is specified, the VLAN will be inferred from this value.  
  Must specify at least one of `ipv4` or `vlan`.
  * `type` - (Optional) The network adapter type.  
  Must be either `E1000` (default) or `VMXNET3`.
* `dns_primary` - (Required) The IP address of the server's primary DNS server.  
If not specified, Google DNS (`8.8.8.8`) is used.
* `dns_secondary` - (Required) The IP address of the server's secondary DNS.  
If not specified, Google DNS (`8.8.4.4`) is used.
* `power_state` - (Optional) Sets the Power state of the server (default is off).
Available Options.
  * `start` - Starts the server, auto starts if server is being created.
  * `off'` (default) Hard stops the server.
  * `shutdown` - Graceful shutdown of the server.

* `tag` - (Optional) A set of tags to apply to the server.
    * `name` - (Required) The tag name. **Note**: The tag name must already be defined for your organisation.
    * `value` - (Required) The tag value.

## Attribute Reference

* `os_type` - The server operating system type (e.g. `CENTOS7/64`).
* `os_family` - The server operating system family (e.g. `UNIX`, `WINDOWS`).
* `primary_adapter_ipv4` - The IPv4 address of the server's primary network adapter.
* `primary_adapter_ipv6` - The IPv6 address of the server's primary network adapter.
* `primary_adapter_vlan` - The Id of the VLAN to which the server's primary network adapter is attached. Calculated if `primary_adapter_ipv4` is specified.
* `public_ipv4` - The server's public IPv4 address (if any). Calculated if there is a NAT rule that points to any of the server's private IPv4 addresses. **Note**: Due to an incompatibility between the CloudControl resource model and Terraform life-cycle model, this attribute is only available after a subsequent refresh (not when the server is first deployed).

## Import

Once declared in configuration, `ddcloud_server` instances can be imported using their Id.

For example:

```bash
$ terraform import ddcloud_server.my-server a79a9273-5362-4af3-91dd-9853b986872c
```
