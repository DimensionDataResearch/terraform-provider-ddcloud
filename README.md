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
	"username"	= "my_username"
	"password"	= "my_password" # Watch out for escaping if your password contains characters such as "$".
	"region"	= "AU" # The DD compute region code (e.g. "AU", "NA", "EU")
}

resource "ddcloud_networkdomain" "test-net-domain" {
	"name"			= "My network domain"
	"description"	= "This is my shiny new network domain."
	"datacenter"	= "AU9" $ The ID of the data centre in which to create your network domain.
}
```

When I get time, I'll be moving the credentials out to a file that won't be under source control (with the ability to fall back to environment variables).
