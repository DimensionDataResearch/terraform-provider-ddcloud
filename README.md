# terraform-provider-ddcloud

[Terraform](https://terraform.io/) provider for Dimension Data cloud compute.

Currently, the following resource types are supported:

* `ddcloud_networkdomain`: A network domain
* `ddcloud_vlan`: A VLAN
* `ddcloud_server`: A virtual machine
* `ddcloud_server_nic`: An additional server network adapter
* `ddcloud_server_anti_affinity`: An anti-affinity rule between 2 servers
* `ddcloud_nat`: A NAT rule (forwards traffic from a public IPv4 address to a server's internal IPv4 address)
* `ddcloud_firewall_rule`: A firewall rule
* `ddcloud_address_list`: A network address list
* `ddcloud_port_list`: A network port list
* `ddcloud_vip_node`: A Virtual IP (VIP) node.
* `ddcloud_vip_pool`: A Virtual IP (VIP) pool.
* `ddcloud_vip_pool_member`: A Virtual IP (VIP) pool membership (node -> pool).

And the following data-source types are supported:

* `ddcloud_networkdomain`: A network domain (lookup by name and data centre).

For more information, see the [provider documentation](docs/).

## Installing the provider

Download the [latest release](https://github.com/DimensionDataResearch/dd-cloud-compute-terraform/releases) and place the provider executable in the same directory as the main Terraform executable.

### Docker

If you prefer, you can run Terraform with the CloudControl provider [using Docker](https://hub.docker.com/r/ddresearch/terraform-provider-ddcloud/):
From the directory containing your configuration (`.tf` files, etc):

```bash
docker run --rm -it -v $PWD:/config ddresearch/terraform-provider-ddcloud /bin/terraform plan
```

## Building the provider yourself

If you want to build the provider yourself instead of installing a pre-built release, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Testing the provider

Create a folder containing a single `.tf` file:

```hcl
/*
 * This configuration will create a single server running CentOS and expose it to the internet on port 80.
 *
 * By default, the CentOS image does not have http installed (`yum install httpd`) so there's no problem exposing port 80.
 */

provider "ddcloud" {
  # User name and password can also be specified via MCP_USER and MCP_PASSWORD environment variables.
  "username"           = "my_username"
  "password"           = "my_password" # Watch out for escaping if your password contains characters such as "$".
  "region"             = "AU" # The DD compute region code (e.g. "AU", "NA", "EU")
}

resource "ddcloud_networkdomain" "my-domain" {
  name                 = "terraform-test-domain"
  description          = "This is my Terraform test network domain."
  datacenter           = "AU9" # The ID of the data centre in which to create your network domain.
}

resource "ddcloud_vlan" "my-vlan" {
  name                 = "terraform-test-vlan"
  description          = "This is my Terraform test VLAN."

  networkdomain        = "${ddcloud_networkdomain.my-domain.id}"

  # VLAN's default network: 192.168.17.1 -> 192.168.17.254 (netmask = 255.255.255.0)
  ipv4_base_address    = "192.168.17.0"
  ipv4_prefix_size     = 24

  depends_on           = [ "ddcloud_networkdomain.my-domain"]
}

resource "ddcloud_server" "my-server" {
  name                 = "terraform-server"
  description          = "This is my Terraform test server."
  admin_password       = "password"

  memory_gb            = 8
  cpu_count            = 2

  networkdomain        = "${ddcloud_networkdomain.my-domain.id}"
  primary_adapter_ipv4 = "192.168.17.10"
  dns_primary          = "8.8.8.8"
  dns_secondary        = "8.8.4.4"

  os_image_name        = "CentOS 7 64-bit 2 CPU"

  # The image disk (part of the original server image). If size_gb is larger than the image disk's original size, it will be expanded (specifying a smaller size is not supported).
  # You don't have to specify this but, if you don't, then Terraform will keep treating the ddcloud_server resource as modified.
  disk {
    scsi_unit_id       = 0
    size_gb            = 10
  }

  # An additional disk.
  disk {
    scsi_unit_id       = 1
    size_gb            = 20
  }

  depends_on           = [ "ddcloud_vlan.my-vlan" ]
}

resource "ddcloud_nat" "my-server-nat" {
  networkdomain       = "${ddcloud_networkdomain.my-domain.id}"
  private_ipv4        = "${ddcloud_server.my-server.primary_adapter_ipv4}"

  # public_ipv4 is computed at deploy time.

  depends_on          = [ "ddcloud_vlan.my-vlan" ]
}

resource "ddcloud_firewall_rule" "my-vm-http-in" {
  name                = "my_server.HTTP.Inbound"
  placement           = "first"
  action              = "accept" # Valid values are "accept" or "drop."
  enabled             = true

  ip_version          = "ipv4"
  protocol            = "tcp"

  # source_address is computed at deploy time (not specified = "any").
  # source_port is computed at deploy time (not specified = "any).
  # You can also specify source_network (e.g. 10.2.198.0/24) or source_address_list instead of source_address.
  # For a ddcloud_vlan, you can obtain these values using the ipv4_baseaddress and ipv4_prefixsize properties.

  # You can also specify destination_network or destination_address_list instead of source_address.
  destination_address = "${ddcloud_nat.my-server-nat.public_ipv4}"
  destination_port    = "80"

  networkdomain       = "${ddcloud_networkdomain.my-domain.id}"
}
```

1. Run `terraform plan -out tf.plan`.
2. Verify that everything looks ok.
3. Run `terraform apply tf.plan`
4. Have a look around and, when it's time to clean up...
5. Run `terraform plan -destroy -out tf.plan`
6. Verify that everything looks ok.
7. Run `terraform apply tf.plan`

## Interactive walkthrough for using .tf files

For a more extensive learning experience in working with Terraform and `.tf` files, please reference the GitHub project below.  There are example files complete with commentary on each resource method currently available while it builds on prior builds before getting to a fully completed Terraform infrastructure build.

https://github.com/wninobla/ddcloud-terraform-examples

## Troubleshooting

For the most part the CloudControl provider for Terraform directly displays any error messages returned by the CloudControl API (and so you can look at the CloudControl documentation for information about what they might mean). These CloudControl API errors will always include an error code of some sort.

If you want to gather additional information to better diagnose your issue, you can try re-running your Terraform command after setting a couple of environment variables:
  * Set `TF_LOG` to `INFO` or `DEBUG`
  * Set `TF_LOG_PATH` to the full path of the log file to write
  * As a last resort, you can also set `MCP_EXTENDED_LOGGING` to `1` (if you set this, then `TF_LOG` must be `DEBUG`) to log requests to and responses from the CloudControl API

The resulting log file will contain detailed information

## Reporting problems

If it's practical to do so, try to make the issue title as short and to-the-point as possible; this helps us to more quickly categorise and prioritise your issue.

### Reporting a bug?

If you have encountered a bug (or behaviour that you believe to be a bug), it's often helpful if you can capture as much of the following information as possible.

### Expected behaviour and actual behaviour

* What were you trying to do when you encountered the problem?
* What did you expect Terraform to do?
* What did it do instead?
* Have you encountered this problem more than once?
* Have you encountered this problem with other sets of `.tf` files or on other machines?

### Steps to reproduce the problem

* The exact Terraform command-line you were running
* The exact program output produced by Terraform
* The terraform configuration you were working with  
If you can't share this, see if you can create a stripped-down version of your configuration that still causes the problem to occur

### System specifications

* CloudControl provider version  
Run `terraform-provider-ddcloud --version`.
* Terraform version  
Run `terraform -version`.
* Operating system  
E.g. OSX 10.11.6, Windows 10, or Ubuntu 12.04
* Anything else you think might be relevant
