# ddcloud\_networkdomain

A Network Domain is the container for all other resource types in your MCP 2.0 deployment.

The `ddcloud_networkdomain` data-source enables lookup of a network domain by name and data centre.

## Example Usage

```
// Existing network domain (not managed by Terraform)
data "ddcloud_networkdomain" "my-domain" {
    name                    = "terraform-test-domain"
    datacenter              = "AU9" # The ID of the data centre in which the network domain exists.
}

// New VLAN in existing network domain
resource "ddcloud_vlan" "my-vlan" {
	// Other properties

	networkdomain           = "${data.ddcloud_networkdomain.my-domain.id}"
}
```

Note that the `data.` prefix is required to reference data-source properties.

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the network domain.
* `datacenter` - (Required) The Id of the MCP 2.0 datacenter in which the network domain is created.

## Attribute Reference

The following attributes are exported:

* `description` - Additional notes (if any) for the network domain.
* `plan` - The plan (service level) for the network domain (`ESSENTIALS` or `ADVANCED`).
* `nat_ipv4_address` - The IPv4 address for the network domain's IPv6->IPv4 Source Network Address Translation (SNAT). This is the IPv4 address of the network domain's IPv4 egress.
