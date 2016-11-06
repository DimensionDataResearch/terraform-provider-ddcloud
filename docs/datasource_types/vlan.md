# ddcloud\_vlan

A VLAN isa virtual network.

The `ddcloud_vlan` data-source enables lookup of a VLAN by name and network domain.

## Example Usage

```
// Existing VLAN (not managed by Terraform)
data "ddcloud_vlan" "my-vlan" {
    name                 = "terraform-test-vlan"
    networkdomain        = "AU9" # The ID of the network domain in which the VLAN exists.
}

// New server attached existing VLAN
resource "ddcloud_server" "my-server" {
	// Other properties

	primary_adapter_vlan = "${data.ddcloud_vlan.my-vlan.id}"
}
```

Note that the `data.` prefix is required to reference data-source properties.

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the VLAN.
* `networkdomain` - (Required) The Id of the network in which the VLAN exists.

## Attribute Reference

The following attributes are exported:

* `description` - Additional notes (if any) for the VLAN.
* `ipv4_base_address` - (Required) The base address of the VLAN's IPv6 network.
* `ipv4_prefix_size` - (Required) The prefix size of the VLAN's IPv6 network.
* `ipv6_base_address` - (Required) The base address of the VLAN's IPv6 network.
* `ipv6_prefix_size` - (Required) The prefix size of the VLAN's IPv6 network.
