# ddcloud\_port\_list

A port list is part of the configuration for a network domain; it identifies one or more network ports (.e.g TCP ports).

The most common use for port lists is to group related ports together to simplify the creation of firewall rules.

## Example Usage

### Simple
The following configuration creates a port list with HTTP and HTTPS ports.

```
resource "ddcloud_port_list" "http_https" {
  name                = "http.and.https"
  description         = "HTTP and HTTPS"

  ports               = [ 80, 443 ]
}
```

### Complex
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

### Nested
```hcl
resource "ddcloud_port_list" "child1" {
  name        = "child.list.1"
  description = "Child port list 1"

  port {
      begin = 80
  }
}

resource "ddcloud_port_list" "child2" {
  name        = "child.list.2"
  description = "Child port list 2"

  port {
      begin = 443
  }
}

resource "ddcloud_port_list" "parent" {
  name         = "parent.list"
  description  = "Parent port list"

  child_lists  = [
    "${ddcloud_port_list.child1.id}",
    "${ddcloud_port_list.child2.id}"
  ]

  networkdomain = "${ddcloud_networkdomain.test_domain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the port list.
Note that port list names can only contain letters, numbers, and periods (`.`).
* `description` - (Required) A description for the port list.
* `port` - (Optional) One or more complex entries to include in the port list.  
For a single port, specify `begin`. For a port range, specify `begin` and `end`.  
Must specify at least one port, or one child list Id.  
Either `port` or `ports` can be specified, but not both.
* `ports` - (Optional) One or more simple entries to include in the port list.  
For a single port, specify `begin`. For a port range, specify `begin` and `end`.  
Must specify at least one port, or one child list Id.  
Either `port` or `ports` can be specified, but not both.
* `child_lists` - (Optional) A list of Ids representing port lists whose ports will to be included in the port list.  
Must specify at least one child list Id, or one port / port-range.

## Attribute Reference

There are currently no additional attributes for `ddcloud_port_list`.
