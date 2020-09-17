# ddcloud\_ip\_address\_block 

A block of Public IPv4 Addresses to the designated Network Domain.

## Example Usage



### IP Block
The following configuration adds a block of Public IPv4 Addresses to the designated Network Domain.

```hcl
resource "ddcloud_ip_address_block" "ipblock_example" {
    domain_id = "${data.ddcloud_networkdomain.mydomain.id}"
    description = "IP block for demo purpose"
    tag {
      name  = "TestTag1"
      value = "TestCustomer1"
    }
}
```
  
## Argument Reference

The following arguments are supported:

* `domain_id` - (Required) he "id" of the Network Domain which the Public IPv4 Address Block will be associated.
* `description` - (Optional) The description for the IP address block.
* `tag` - (Optional) Tagging is optional. 

## Attribute Reference

There are currently no additional attributes for `ddcloud_ip_address_reservation`.
