# ddcloud\_static_route

Network Domain Static Routes is an advanced topic limited to ENTERPRISE type Network Domains. For further information please refer to Introduction to IP Addressing and Routing in MCP 2.0
(https://docs.mcp-services.net/x/GwIu).

This resource creates a new CLIENT Static Route on a Network Domain in an MCP 2.0 data center location. Client created Static Routes co-exist with any SYSTEM Static Routes that are present on the Network Domain.
CloudControl supports creation of overlapping Static Routes that point to different nextHopAddress destinations. In such cases routing for an IP address that resides in both Static Routes, which default to the nextHopAddress for the more specific (higher order prefix) Static Route.



## Example Usage



```
resource "ddcloud_static_route" "my_static_route_A" {
    name = "MyStaticRouteA"
    description = "Test static route A"
    networkdomain = "1cecad11-0d41-4abf-a0c2-475563eef108"
    ip_version = "IPV4"
    destination_network_address = "100.102.0.0"
    destination_prefix_size = 16
    next_hop_address = "100.65.1.119"   
}
```





## Argument Reference

The following arguments are supported:

For a single address, specify `begin`. 
For an address range, specify `begin` and `end`. 
For an IP network, specify `network` and `prefix_size`.  

| Arguments    | Required | Description |
| ---------    | ----------- | ----------|
| networkdomain | Yes       |  UUID of a Network Domain belonging to {org-id} within which the Static Route is to be created. |
| name  | Yes| Must be between 1 and 75 characters in length. Cannot start with a number, a period ('.'), 'CCSYSTEM.' or 'CCDEFAULT.'. Can only include alphanumeric characters, '_' and '.'. Spaces are not supported. |
| description  | No (Optional)|    |  Maximum length: 255 characters.|
| ip_version  |  Yes | Either 'IPV4' or 'IPV6' (Upper case)|
|destination_network_address| Yes | Either a valid IPv4 address in dot-decimal notation or an IPv6 address in compressed or extended format. In conjunction with the destinationPrefixSize this must represent a CIDR boundary. |
|destination_prefix_size| Yes | Integer prefix defining the size of the network. |
|next_hop_address| Yes |Gateway address in the form of an INET gateway, CPNC gateway or an address on an Attached VLAN in the same Network Domain. Cannot be a system-reserved address on the Attached VLAN if an Attached VLAN is referenced. For details of system-reserved addresses please refer to Introduction to IP Addressing and Routing in MCP 2.0 (https://docs.mcp-services.net/x/GwIu).|

## Attribute Reference

There are currently no additional attributes for `ddcloud_static_route`.


