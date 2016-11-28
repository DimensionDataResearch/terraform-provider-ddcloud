# Managed Cloud Platform

Managed Cloud Platform is Dimension Data's cloud virtualisation service. The API for controlling the Managed Cloud Platform is called CloudControl.
Use the navigation to the left to read about the available resources.

## Example Usage

The following configuration will create a single server running CentOS and expose it to the internet on port 80.

By default, the Managed Cloud Platform's CentOS image does not have httpd installed (`yum install httpd`) so there should be no problem exposing port 80.

```
provider "ddcloud" {
  "username" = "my_username"
  "password" = "my_password"
  "region"   = "AU"
}

resource "ddcloud_networkdomain" "mydomain" {
  name        = "terraform-test-domain"
  description = "My Terraform test network domain."
  datacenter  = "AU9"
}

resource "ddcloud_vlan" "myvlan" {
  name              = "terraform-test-vlan"
  description       = "My Terraform test VLAN."

  networkdomain     = "${ddcloud_networkdomain.mydomain.id}"

  ipv4_base_address = "192.168.17.0"
  ipv4_prefix_size  = 24
}

resource "ddcloud_server" "myserver" {
  name                 = "terraform-server"
  description          = "My Terraform test server."
  admin_password       = "password"

  memory_gb            = 8
  cpu_count            = 2

  networkdomain        = "${ddcloud_networkdomain.mydomain.id}"
  primary_adapter_ipv4 = "192.168.17.10"
  dns_primary          = "8.8.8.8"
  dns_secondary        = "8.8.4.4"

  os_image_name        = "CentOS 7 64-bit 2 CPU"
}

resource "ddcloud_nat" "mynat" {
  networkdomain = "${ddcloud_networkdomain.mydomain.id}"
  private_ipv4  = "${ddcloud_server.myserver.primary_adapter_ipv4}"

  depends_on    = [ "ddcloud_vlan.myvlan" ]
}

resource "ddcloud_firewall_rule" "myvm_http_in" {
  name                = "my_server.http.inbound"
  placement           = "first"
  action              = "accept"
  enabled             = true

  ip_version          = "ipv4"
  protocol            = "tcp"

  destination_address = "${ddcloud_nat.mynat.public_ipv4}"
  destination_port    = "80"

  networkdomain       = "${ddcloud_networkdomain.mydomain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The Managed Cloud Platform region code (e.g. 'AU' - Australia, 'EU' - Europe, 'NA' - North America) that identifies the CloudControl end-point to connect to.  
Must specify exactly one of either `region` or `cloudcontrol_endpoint`.
* `cloudcontrol_endpoint` - (Optional) The base URL of the CloudControl end-point to connect to.  
Use this property if you are using PCEE or some other custom end-point that does not follow the standard pattern (`https://api-<region>.dimensiondata.com/`).  
Must specify exactly one of either `cloudcontrol_endpoint` or `region`.
* `username` - (Optional) The user name for authenticating to CloudControl.  
If not specified, the `MCP_USER` environment variable will be used instead.
* `password` - (Optional) The password for authenticating to CloudControl.  
If not specified, the `MCP_PASSWORD` environment variable will be used instead.
* `region` - (Optional) The Managed Cloud Platform region code (e.g. 'AU' - Australia, 'EU' - Europe, 'NA' - North America) that identifies the CloudControl end-point to connect to.
* `retry_timeout` - (Optional) The time (in seconds) to wait before before retrying an operation due to a `RESOURCE_BUSY` response from CloudControl times out.    
Default is 10 minutes.
* `retry_delay` - (Optional) The time (in seconds) to delay between operation retries due to `RESOURCE_BUSY` responses from CloudControl.  
Default is 30 seconds.
* `allow_server_reboot` - (Optional) Allow servers to be rebooted due to configuration changes?  
  If `false`, then the provider will fail any operation (except deletion) that requires a server to be rebooted.  
  Default is `true`.
