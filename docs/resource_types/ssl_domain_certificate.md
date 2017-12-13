# ddcloud\_ssl\_domain\_certificate

An X.509 certificate (with private key) used for SSL offloading.

## Example Usage

```
resource "ddcloud_ssl_domain_certificate" "my_cert" {
  name        = "MyCertificate"
  certificate = "${file("./certificate.pem")}"
  private_key = "${file("./private-key.pem")}"

  networkdomain = "${ddcloud_networkdomain.my_networkdomain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `networkdomain` - (Required) The Id of the network domain in which the SSL domain certificate will be used for SSL offload.
* `name` - (Required) A name for the certificate.
* `description` - (Optional) A description for the certificate.
* `certificate` - (Required) The X.509 certificate (in PEM format, use `ddcloud_pfx` data source if you need to use a certificate from a `.pfx` file).
* `private_key` - (Required) The private key (in PEM format).

## Attribute Reference

There are currently no additional attributes for `ddcloud_ssl_domain_certificate`.

## Import

Once declared in configuration, `ddcloud_ssl_domain_certificate` instances can be imported using their Id.

For example:

```bash
$ terraform import ddcloud_ssl_domain_certificate.my_certificate 87d42402-6bec-494d-b365-31971e415bc4
```
