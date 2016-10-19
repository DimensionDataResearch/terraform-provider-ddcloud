# ddcloud\_address\_list

An address list is part of the configuration for a network domain; it identifies one or more IP addresses (all from the same address family, e.g. IPv4 / IPv6).

The most common use for address lists is to group related addresses together to simplify the creation of firewall rules.

## Example Usage

The following configuration creates an address list with several IPv4 addresses (a single address, an address range, and a network).

```
resource "ddcloud_address_list" "test_list" {
	name        	= "TestAddresses"
	ip_version		= "IPv4"

    # A single address
	address {
		begin		= "192.168.1.1"
	}

    # An address range
	address {
		begin		= "10.0.1.15"
		end 		= "10.0.1.20"
	}

    # An IP network
	address {
		network		= "10.15.7.0"
		prefix_size	= 24
	}

    networkdomain 	= "${ddcloud_networkdomain.test_domain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the port list.
Note that port list names can only contain letters, numbers, and periods (`.`).
* `description` - (Optional) A description for the port list.
* `address` - (Required) One or more entries to include in the address list.  
For a single address, specify `begin`. For an address range, specify `begin` and `end`. For an IP network, specify `network` and `prefix_size`.
* `child_lists` - (Optional) A list of Ids representing address lists whose addresses will to be included in the port list.

## Attribute Reference

There are currently no additional attributes for `ddcloud_port_list`.
