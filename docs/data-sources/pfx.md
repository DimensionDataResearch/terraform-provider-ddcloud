# ddcloud\_pfx

The PFX (also known as PKCS12) format is used to store one or more certificates and / or private keys, protected by a password.

The `ddcloud_pfx` data-source enables decoding of a `.pfx` file into PEM-format certificate and private key.

**Note:** only the first certificate and private key in the `.pfx` file will be returned.

## Example Usage

```hcl
// Extract certificate and private key from test.pfx
data "ddcloud_pfx" "server_cert" {
    file        = "./server.pfx"
    password    = "Hello123"
}

output "certificate_pem" {
	value = "${data.ddcloud_pfx.server_cert.certificate}"
}
output "private_key_pem" {
	value = "${data.ddcloud_pfx.server_cert.private_key}"
}
```

Note that the `data.` prefix is required to reference data-source properties.

## Argument Reference

The following arguments are supported:

* `file` - (Required) The path to the `.pfx` file.
* `password` - (Required) The password used to decrypt the file's contents.

## Attribute Reference

The following attributes are exported:

* `certificate` - The first certificate found in the `.pfx` file, in PEM format.
* `private_key` - The first private key found in the `.pfx` file, in PEM format.
