# ddcloud\_server

A Server is a virtual machine. It is deployed in a network domain, and each of its network adapters are connected to a VLAN. Each server is created from an OS or customer image that also specifies sensible defaults for the server configuration.

**Note**: Currently, the provider can only manage a server's primary network adapter.

## Example Usage

```
resource "ddcloud_server" "myserver" {
  name                 = "terraform-server"
  description          = "My Terraform test server."
  admin_password       = "password"

  memory_gb            = 8
  cpu_count            = 2
  cpu_speed            = "STANDARD"
  cores_per_cpu        = 1

  networkdomain        = "${ddcloud_networkdomain.mydomain.id}"
  primary_adapter_vlan = "${ddcloud_vlan.myvlan.id}"
  primary_adapter_ipv4 = "192.168.17.10"

  dns_primary          = "8.8.8.8"
  dns_secondary        = "8.8.4.4"

  disk {
      scsi_unit_id     = 0
      size_gb          = 10
      speed            = "STANDARD"
  }

  os_image_name        = "CentOS 7 64-bit 2 CPU"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the server.
* `description` - (Optional) A description for the server.
* `admin_password` - (Required) The initial administrative password for the deployed server.  
Has no effect after deployment.
* `memory_gb` - (Optional) The amount of memory (in GB) allocated to the server.  
Defaults to the memory specified by the image from which the server is created.
* `cpu_count` - (Optional) The number of CPUs allocated to the server.  
Defaults to the CPU count specified by the image from which the server is created.
* `cores_per_cpu` - (Optional) The number of cores per virtual CPU socket allocated to the server.  
Defaults to the number of cores specified by the image from which the server is created.
* `cpu_speed` - (Optional) The speed of the CPU(s) allocated to the server (`STANDARD` or `HIGHPERFORMANCE`).  
Default is `STANDARD`.
* `disk` - (Optional) The set of virtual disks attached to the server.
    * `scsi_unit_id` - (Required) The SCSI Logical Unit Number (LUN) for the disk. Must be unique across the server's disks.
    * `size_gb` - (Required) The size (in GB) of the disk. This value can be increased (to expand the disk) but not decreased.
    * `speed` - (Required) The disk speed. Usually one of `STANDARD`, `ECONOMY`, or `HIGHPERFORMANCE` (but varies between data centres).
* `networkdomain` - (Required) The Id of the network domain in which the server is deployed.
* `primary_adapter_ipv4` - (Optional) The IPv4 address for the server's primary network adapter.  
Must specify at least one of `primary_adapter_ipv4` or `primary_adapter_vlan`.
* `primary_adapter_vlan` - (Optional) The Id of the VLAN to which the server's primary network adapter will be attached (the first available IPv4 address will be allocated).  
Must specify at least one of `primary_adapter_vlan` or `primary_adapter_ipv4`.  
`primary_adapter_vlan` is ignored if `primary_adapter_ipv4` is also specified.  
It's still useful to supply both, though, since it sets up a dependency between the server and the VLAN.
**Note**: Changing this property will result in the server being destroyed and re-created.
* `primary_adapter_type` - (Optional) The type of the server's primary network adapter (`E1000` or `VMXNET3`).  
**Note**: Changing this property will result in the server being destroyed and re-created.
* `dns_primary` - (Required) The IP address of the server's primary DNS server.  
If not specified, the default (region-dependent) CloudControl DNS is used.
* `dns_secondary` - (Required) The IP address of the server's secondary DNS.  
If not specified, the default (region-dependent) CloudControl DNS is used.
* `os_image_id` - (Required) The Id of the OS (built-in) image from which the server will be created. Must specify exactly one of `os_image_id`, `os_image_name`, `customer_image_id`, `customer_image_name`.
* `os_image_name` - (Required) The name of the OS (built-in) image from which the server will be created (the name must be unique within the data center in which the network domain is deployed). Must specify exactly one of `os_image_id`, `os_image_name`, `customer_image_id`, `customer_image_name`.
* `customer_image_id` - (Required) The Id of the customer (custom) image from which the server will be created. Must specify exactly one of `os_image_id`, `os_image_name`, `customer_image_id`, `customer_image_name`.
* `customer_image_name` - (Required) The name of the customer (custom) image from which the server will be created (the name must be unique within the data center in which the network domain is deployed). Must specify exactly one of `os_image_id`, `os_image_name`, `customer_image_id`, `customer_image_name`.
* `auto_start` - (Optional) Automatically start the server once it is deployed (default is false).
* `tag` - (Optional) A set of tags to apply to the server.
    * `name` - (Required) The tag name. **Note**: The tag name must already be defined for your organisation.
    * `value` - (Required) The tag value.

## Attribute Reference

* `os_image_id` - The Id of the OS image (if any) from which the server was created. Calculated if `os_image_name` is specified.
* `os_image_name` - The name of the OS image (if any) from which the server was created. Calculated if `os_image_id` is specified.
* `customer_image_id` - The Id of the customer image (if any) from which the server was created. Calculated if `customer_image_name` is specified.
* `customer_image_name` - The name of the customer image (if any) from which the server was created. Calculated if `customer_image_id` is specified.
* `primary_adapter_ipv4` - The IPv4 address of the server's primary network adapter. Calculated if `primary_adapter_vlan` is specified.
* `primary_adapter_ipv6` - The IPv6 address of the server's primary network adapter.
* `primary_adapter_vlan` - The Id of the VLAN to which the server's primary network adapter is attached. Calculated if `primary_adapter_ipv4` is specified.
* `public_ipv4` - The server's public IPv4 address (if any). Calculated if there is a NAT rule that points to any of the server's private IPv4 addresses. **Note**: Due to an incompatibility between the CloudControl resource model and Terraform life-cycle model, this attribute is only available after a subsequent refresh (not when the server is first deployed).
