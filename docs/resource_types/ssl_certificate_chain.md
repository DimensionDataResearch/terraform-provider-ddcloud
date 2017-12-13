# ddcloud\_ssl\_certificate\_chain

An X.509 certificate chain used for SSL offloading.

## Example Usage

```hcl
resource "ddcloud_ssl_certificate_chain" "my_chain" {
  name          = "MyChain"
  chain         = "${file("./chain.pem")}"

  networkdomain = "${ddcloud_networkdomain.my_networkdomain.id}"
}
```

## Argument Reference

The following arguments are supported:

* `networkdomain` - (Required) The Id of the network domain in which the SSL domain certificate will be used for SSL offload.
* `name` - (Required) A name for the certificate.
* `description` - (Optional) A description for the certificate.
* `chain` - (Required) The X.509 certificate chain (in PEM format).

## Attribute Reference

There are currently no additional attributes for `ddcloud_ssl_certificate_chain`.

## Import

Once declared in configuration, `ddcloud_ssl_certificate_chain` instances can be imported using their Id.

For example:

```bash
$ terraform import ddcloud_ssl_certificate_chain.my_chain 87d42402-6bec-494d-b365-31971e415bc4
```
