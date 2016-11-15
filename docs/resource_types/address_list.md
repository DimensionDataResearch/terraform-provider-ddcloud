# ddcloud\_address\_list

An address list is part of the configuration for a network domain; it identifies one or more IP addresses (all from the same address family, e.g. IPv4 / IPv6).

The most common use for address lists is to group related addresses together to simplify the creation of firewall rules.

## Example Usage

### Simple
The following configuration creates an address list with several IPv4 addresses.

```
resource "ddcloud_address_list" "test_list" {
  name       = "TestAddresses"
  ip_version = "IPv4"

  addresses  = [
    "10.0.1.15",
    "10.0.1.16",
    "10.0.1.17",
    "10.0.1.18",
    "10.0.1.19",
    "10.0.1.20"
  ]

  networkdomain = "${ddcloud_networkdomain.test_domain.id}"
}
```

### Complex
The following configuration creates an address list with more complex IPv4 addresses (a single address, an address range, and a network).

```
resource "ddcloud_address_list" "test_list" {
  name       = "TestAddresses"
  ip_version = "IPv4"

  # A single address
  address {
    begin       = "192.168.1.1"
  }

  # An address range
  address {
    begin       = "10.0.1.15"
    end         = "10.0.1.20"
  }

  # An IP network
  address {
    network     = "10.15.7.0"
    prefix_size = 24
  }

  networkdomain = "${ddcloud_networkdomain.test_domain.id}"
}
```

### Nested
The following configuration creates 2 address lists each with a single IPv4 address. Then creates another address list with the original 2 address lists as children.

```
resource "ddcloud_address_list" "child1" {
  name         = "ChildList1"
  ip_version   = "IPv4"

  address {
    begin       = "192.168.1.20"
  }

  networkdomain = "${ddcloud_networkdomain.test_domain.id}"
}

resource "ddcloud_address_list" "child2" {
  name         = "ChildList2"
  ip_version   = "IPv4"

  address {
    begin       = "192.168.1.21"
  }

  networkdomain = "${ddcloud_networkdomain.test_domain.id}"
}

resource "ddcloud_address_list" "parent" {
  name         = "ParentList"
  ip_version   = "IPv4"

  child_lists  = [
    "${ddcloud_address_list.child1.id}",
    "${ddcloud_address_list.child2.id}"
  ]

  networkdomain = "${ddcloud_networkdomain.test_domain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the port list.
Note that port list names can only contain letters, numbers, and periods (`.`).
* `description` - (Optional) A description for the port list.
* `ip_version` - (Required) The IP version (IPv4 / IPv6) of the addresses that the list contains.
* `address` - (Optional) One or more entries to include in the address list.  
For a single address, specify `begin`. For an address range, specify `begin` and `end`. For an IP network, specify `network` and `prefix_size`.  
Must specify at least one address, or one child list Id.  
Either `address` or `addresses` can be specified, but not both.
* `addresses` - (Optional) One or simple IP addresses to include in the address list.  
Must specify at least one address, or one child list Id.  
Either `address` or `addresses` can be specified, but not both.
* `child_lists` - (Optional) A list of Ids representing address lists whose addresses will to be included in the port list.  
Must specify at least one address, or one child list Id.

## Attribute Reference

There are currently no additional attributes for `ddcloud_port_list`.
