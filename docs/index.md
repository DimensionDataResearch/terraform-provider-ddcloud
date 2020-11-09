# Managed Cloud Platform

The Dimension Data Managed Cloud Platform provider is used to interact with Dimension Data's Managed Cloud Platform resources.

## Provider configuration

See the documentation for [provider configuration](guides/provider.md).

## Resource types

The `ddcloud` provider supports the following resource types:

* [ddcloud_networkdomain](resources/networkdomain.md) - A CloudControl network domain.
* [ddcloud_vlan](resources/vlan.md) - A CloudControl Virtual LAN (VLAN).
* [ddcloud_server](resources/server.md) - A CloudControl Server (virtual machine).
* [ddcloud_storage_controller](resources/storage_controller.md) - A SCSI controller in a CloudControl Server.
* [ddcloud_network_adapter](resources/network_adapter.md) - An additional network adapter for a CloudControl Server.
* [ddcloud_server_backup](resources/server_backup.md) - Backup configuration for a CloudControl Server.
* [ddcloud_server_anti_affinity](resources/server_anti_affinity.md) - Anti-affinity rule for 2 CloudControl Servers (virtual machines).
* [ddcloud_nat](resources/nat.md) - A CloudControl Network Address Translation (NAT) rule.
* [ddcloud_firewall_rule](resources/firewall_rule.md) - A CloudControl firewall rule.
* [ddcloud_address_list](resources/address_list.md) - A CloudControl network address list.
* [ddcloud_port_list](resources/port_list.md) - A CloudControl network port list.
* [ddcloud_vip_node](resources/vip_node.md) - A CloudControl Virtual IP (VIP) node.
* [ddcloud_vip_pool](resources/vip_pool.md) - A CloudControl Virtual IP (VIP) pool.
* [ddcloud_vip_pool_member](resources/vip_pool_member.md) - A CloudControl Virtual IP (VIP) pool membership.  
Links a `ddcloud_vip_node` (and optionally a port) to a `ddcloud_vip_pool`.
* [ddcloud_virtual_listener](resources/virtual_listener.md) - A CloudControl Virtual Listener.
* [ssl_offload_profile](resources/ssl_offload_profile.md) - An SSL-offload profile used by a Virtual Listener.
* [ssl_domain_certificate](resources/ssl_domain_certificate.md) - An X.509 certificate (with private key) for SSL offload.
* [ssl_certificate_chain](resources/ssl_certificate_chain.md) - An X.509 certificate chain for SSL offload.
* [ddcloud_ip_address_reservation](resources/ip_address_reservation.md) - An IP address reservation on a CloudControl VLAN (experimental, for advanced usage scenarios only).

And the following data-source types:

* [ddcloud_networkdomain](data-sources/networkdomain.md) - A CloudControl network domain (lookup by name and data centre).
* [ddcloud_vlan](data-sources/vlan.md) - A CloudControl Virtual LAN (VLAN) (lookup by name and network domain).
* [ddcloud_pfx](data-sources/pfx.md) - Enables decoding of a `.pfx` file into PEM-format certificate and private key (useful for SSL-offload resources).

## Migration

For information about migrating from v1.0 or v1.1 to v1.2, see the [v1.2 migration docs](guides/migrating/v1.1-v1.2.md).
