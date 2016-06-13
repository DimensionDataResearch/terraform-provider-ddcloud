# dd-cloud-compute-terraform
Terraform provider Dimension Data cloud compute.

This is a work in progress. Currently, the following resource types are supported:

* `ddcloud_networkdomain`: A network domain

To get started:

* On windows, create / update `$HOME\terraform.rc`
* On Linux / OSX, create / update `~/.terraformrc`

And add the following contents:

```
providers {
	ddcloud = "path-to-the-folder/containing/terraform-provider-ddcloud"
}
```

Create a folder containing a single `.tf` file:

```
provider "ddcloud" {
	"username"				= "my_username"
	"password"				= "my_password" # Watch out for escaping if your password contains characters such as "$".
	"region"				= "AU" # The DD compute region code (e.g. "AU", "NA", "EU")
}

resource "ddcloud_networkdomain" "my-domain" {
	name					= "terraform-test-domain"
	description				= "This is my Terraform test network domain."
	datacenter				= "AU9" # The ID of the data centre in which to create your network domain.
}

resource "ddcloud_vlan" "my-vlan" {
	name					= "terraform-test-vlan"
	description 			= "This is my Terraform test VLAN."

	networkdomain 			= "${ddcloud_networkdomain.my-domain.id}"

	# VLAN's default network: 192.168.17.1 -> 192.168.17.254 (netmask = 255.255.255.0)
	ipv4_base_address		= "192.168.17.0"
	ipv4_prefix_size		= 24

	depends_on				= [ "ddcloud_networkdomain.my-domain"]
}

resource "ddcloud_server" "my-server" {
	name					= "terraform-server"
	description 			= "This is my Terraform test server."
	admin_password			= "password"

	networkdomain 			= "${ddcloud_networkdomain.test-domain.id}"
	primary_adapter_ipv4	= "192.168.17.10"
	dns_primary				= "8.8.8.8"
	dns_secondary			= "8.8.4.4"

	osimage_name			= "CentOS 7 64-bit 2 CPU"

	depends_on				= [ "ddcloud_networkdomain.my-domain", "ddcloud_vlan.my-vlan" ]
}
```

1. Run `terraform plan -out tf.plan`.
2. Verify that everything looks ok.
3. Run `terraform apply tf.plan`
4. Have a look around and, when it's time to clean up...
5. Run `terraform plan -destroy -out tf.plan`
6. Verify that everything looks ok.
7. Run `terraform apply tf.plan`

When I get time, I'll also enable specifying credentials via environment variables.
