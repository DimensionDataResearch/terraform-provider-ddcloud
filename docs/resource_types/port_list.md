# ddcloud\_port\_list

A port list is part of the configuration for a network domain; it identifies one or more network ports (.e.g TCP ports).

The most common use for port lists is to group related ports together to simplify the creation of firewall rules.

## Example Usage

The following configuration creates a port list with HTTP and HTTPS ports, as well as ports in the range 8000-9600.

```
resource "ddcloud_port_list" "http_https" {
  name                = "http.and.https"
  description         = "HTTP and HTTPS"

  port {
      begin = 80
  }

  port {
      begin = 443
  }

  port {
      begin = 8000
      end   = 9600
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the port list.
Note that port list names can only contain letters, numbers, and periods (`.`).
* `description` - (Required) A description for the port list.
* `port` - (Optional) One or more entries to include in the port list.  
For a single port, specify `begin`. For a port range, specify `begin` and `end`.
* `child_lists` - (Optional) A list of Ids representing port lists whose ports will to be included in the port list.

## Attribute Reference

There are currently no additional attributes for `ddcloud_port_list`.
