# ddcloud\_address

Can be used to define one or more individual IP addresses or ranges of IP addresses. Duplicates are not supported.

##ipVersion: IPV4:
Must be a valid IPv4 address in dot-decimal notation (w.x.y.z).
If provided, prefixSize must be an integer from 1-32 inclusive and must be consistent with the associated IP address.

##ipVersion: IPV6:
Must be a valid IPv6 address in compressed or extended format.
If provided, prefixSize must be an integer from 1-128 inclusive and must be consistent with the associated IP address.
Common:
prefixSize and end are mutually exclusive. end must be greater than begin for a range.


The most common use for address is to add an address to an existing AddressList


## Example Usage

### Single IP
The following configuration creates an address with single IP.

```

resource "ddcloud_address" "single_ip" {
  begin   = "10.1.1.61"
  
  networkdomain = "${data.ddcloud_networkdomain.my-domain.id}"
  addresslist_id = "${data.ddcloud_addresslist.my-addresslist.id}"
}
```

### IP Range
The following configuration creates an address with IP range.

```
resource "ddcloud_address" "ip_range" {
  begin   = "10.1.1.71"
  end     = "10.1.1.73" 

  networkdomain = "${data.ddcloud_networkdomain.my-domain.id}"
  addresslist_id = "${data.ddcloud_addresslist.my-addresslist.id}"
}
```

### IP Network
The following configuration creates an address with IP Network.
```
resource "ddcloud_address" "ip_subnet" {
  network = "198.51.100.0"
  prefix_size = 24
  
  networkdomain = "${data.ddcloud_networkdomain.my-domain.id}"
  addresslist_id = "${data.ddcloud_addresslist.my-addresslist.id}"
}
```




## Argument Reference

The following arguments are supported:

For a single address, specify `begin`. 
For an address range, specify `begin` and `end`. 
For an IP network, specify `network` and `prefix_size`.  

| Arguments    | Description |
| ---------    | ----------- |
| begin        |  IP Address e.g. 10.0.1.1 |
| end          | IP Address e.g. 10.0.1.3 |
| network      | subnet e.g. 10.1.3.0 |
| prefix_size  | e.g. 32 |

## Attribute Reference

There are currently no additional attributes for `ddcloud_address`.

## Notes
IPv6 is supported. Make sure you input IPv6 address as how DD Cloud will store it.
DD Cloud will parse the input of IPv6 addresss and save it in a simplified format. 
>For example, if you input ipv6 '2001:db8:abcd:0012:0000:0000:0000:0000', 
DD Cloud will be format it into '2001:db8:abcd:12:0:0:0:0'.
This will cause terraform to treat them as a different resource value. 
