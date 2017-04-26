# ddcloud\_storage\_controller

A storage controller represents a SCSI adapter in a [Server](server.md). Each storage controller emulates a specific adapter type, and 0 or more attached disks.

## Notes
* You can either declare your server's disks directly on the `ddcloud_server` resource, or use 1 or more `ddcloud_storage_controller` resources and declare the disks inside them. _Do not declare disks in both places or you run the risk of confusing the Terraform provider and damaging your server's configuration._
* If you are using more than storage controller in your server, you _must_ use `ddcloud_storage_controller`.
* There is a minimum number of disks per server (usually 1, but can vary by datacenter).  
  If a `ddcloud_storage_controller` is being deleted and the provider needs to remove its corresponding storage controller then, if removing that storage controller would involve removing too many disks from the server (e.g. the server's last disk), the controller will not be removed (but will be treated as if it has been).

## Example Usage

```hcl
resource "ddcloud_server" "myserver" {
  name                 = "terraform-server"
  
  // Do not declare disks here if you are using ddcloud_storage controller
}

resource "ddcloud_storage_controller "myserver_controller_0" {
  server          = "${ddcloud_server.myserver.id}"
  scsi_bus_number = 0

  // You can omit adapter_type if you are using LSI_LOGIC_PARALLEL (this is the default value).
  
  // Instead, declare them here.
  disk {
      scsi_unit_id     = 0
      size_gb          = 10
      speed            = "STANDARD"
  }
}

resource "ddcloud_storage_controller "myserver_controller_0" {
  server          = "${ddcloud_server.myserver.id}"
  scsi_bus_number = 1

  adapter_type    = "LSI_LOGIC_SAS"
  
  disk {
      scsi_unit_id     = 0
      size_gb          = 50
      speed            = "STANDARD"
  }

  disk {
      scsi_unit_id     = 1
      size_gb          = 500
      speed            = "ECONOMY"
  }
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) The Id of the server that the storage controller is attached to.  
**Note**: Changing this property will result in the storage controller (and its attached disks) being destroyed and recreated.
* `scsi_bus_number` - (Required) The storage controller's SCSI bus number.  
Behaviour is undefined if you have more than one `ddcloud_storage_controller` pointing to the same server with the same `scsi_bus_number`.  
**Note**: Changing this property will result in the storage controller (and its attached disks) being destroyed and recreated.
* `adapter_type` - (Optional) The type of adapter emulated by the storage controller (one of `BUSLOGIC_PARALLEL`, `LSI_LOGIC_PARALLEL`, `LSI_LOGIC_SAS`, or `VMWARE_PARAVIRTUAL`).  
Default value: `LSI_LOGIC_PARALLEL`.  
**Note**: Changing this property will result in the storage controller (and its attached disks) being destroyed and recreated.
* `disk` - (Optional) The set of disks attached to the storage controller.
    * `scsi_unit_id` - (Required) The SCSI Logical Unit Number (LUN) for the disk. Must be unique across the controller's disks.
    * `size_gb` - (Required) The size (in GB) of the disk. This value can be increased (to expand the disk) but not decreased.
    * `speed` - (Required) The disk speed. Usually one of `STANDARD`, `ECONOMY`, or `HIGHPERFORMANCE` (but varies between data centres).

## Attribute Reference

This resource does not expose any additional attributes.
