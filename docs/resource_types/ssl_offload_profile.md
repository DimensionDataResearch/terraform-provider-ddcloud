# ddcloud\_ssl\_offload\_profile

A profile for SSL offloading used by a virtual listener.

See the [CloudControl documentation](https://docs.mcp-services.net/display/CCD/Introduction+to+SSL+Offload%2C+including+SSL+Domain+Certificate%2C+SSL+Certificate+Chain%2C+and+SSL+Offload+Profiles) for more information about SSL offloading.

## Example Usage

```hcl
resource "ddcloud_ssl_offload_profile" "my_offload_profile" {
  name        = "MyOffloadProfile"
  certificate = "${ddcloud_ssl_domain_certificate.my_certificate.id}"
  chain       = "${ddcloud_ssl_certificate_chain.my_chain.id}"
  ciphers     = "MEDIUM:HIGH:!EXPORT:!ADH:!MD5:!RC4:!SSLv2:!SSLv3:!ECDHE+AES-GCM:!ECDHE+AES:!ECDHE+3DES:!ECDHE_ECDSA:!ECDH_RSA:!ECDH_ECDSA:@SPEED"

  networkdomain = "${data.ddcloud_networkdomain.primary.id}"
}
```

## Argument Reference

The following arguments are supported:

* `networkdomain` - (Required) The Id of the network domain in which the SSL domain certificate will be used for SSL offload.
* `name` - (Required) A name for the certificate.
* `description` - (Optional) A description for the certificate.
* `certificate` - (Required) The Id of the SSL domain certificate to use.
* `chain` - (Optional) The Id of the SSL certificate chain (if any) to use.
* `ciphers` - (Optional, Computed) SSL ciphers to use.  
  If not specified, then CloudControl will a default selection of ciphers (see the CloudControl documentation for details).

## Attribute Reference

There are currently no additional attributes for `ddcloud_ssl_offload_profile`.

## Import

Once declared in configuration, `ddcloud_ssl_offload_profile` instances can be imported using their Id.

For example:

```bash
$ terraform import ddcloud_ssl_offload_profile.my_offload_profile 87d42402-6bec-494d-b365-31971e415bc4
```
