# Managed Cloud Platform

The Dimension Data Managed Cloud Platform provider is used to interact with Dimension Data's Managed Cloud Platform resources.

## Provider configuration

See the documentation for [provider configuration](provider.md).

## Resource types

The `ddcloud` provider supports the following resource types:

* [ddcloud_networkdomain](resource_types/networkdomain.md) - A CloudControl network domain.
* [ddcloud_vlan](resource_types/vlan.md) - A CloudControl Virtual LAN (VLAN).
* [ddcloud_server](resource_types/server.md) - A CloudControl Server (virtual machine).
* [ddcloud_storage_controller](resource_types/storage_controller.md) - A SCSI controller in a CloudControl Server.
* [ddcloud_network_adapter](resource_types/network_adapter.md) - An additional network adapter for a CloudControl Server.
* [ddcloud_server_anti_affinity](resource_types/server_anti_affinity.md) - Anti-affinity rule for 2 CloudControl Servers (virtual machines).
* [ddcloud_nat](resource_types/nat.md) - A CloudControl Network Address Translation (NAT) rule.
* [ddcloud_firewall_rule](resource_types/firewall_rule.md) - A CloudControl firewall rule.
* [ddcloud_address_list](resource_types/address_list.md) - A CloudControl network address list.
* [ddcloud_port_list](resource_types/port_list.md) - A CloudControl network port list.
* [ddcloud_vip_node](resource_types/vip_node.md) - A CloudControl Virtual IP (VIP) node.
* [ddcloud_vip_pool](resource_types/vip_pool.md) - A CloudControl Virtual IP (VIP) pool.
* [ddcloud_vip_pool_member](resource_types/vip_pool_member.md) - A CloudControl Virtual IP (VIP) pool membership.  
Links a `ddcloud_vip_node` (and optionally a port) to a `ddcloud_vip_pool`.
* [ddcloud_virtual_listener](resource_types/virtual_listener.md) - A CloudControl Virtual Listener.
* [ssl_domain_certificate](resource_types/ssl_domain_certificate.md) - A certificate (with private key) for SSL offload.
* [ddcloud_ip_address_reservation](resource_types/ip_address_reservation.md) - An IP address reservation on a CloudControl VLAN (experimental, for advanced usage scenarios only).

And the following data-source types:

* [ddcloud_networkdomain](datasource_types/networkdomain.md) - A CloudControl network domain (lookup by name and data centre).
* [ddcloud_vlan](datasource_types/vlan.md) - A CloudControl Virtual LAN (VLAN) (lookup by name and network domain).
* [ddcloud_pfx](datasource_types/pfx.md) - Enables decoding of a `.pfx` file into PEM-format certificate and private key (useful for SSL-offload resources).

## Migration

For information about migrating from v1.0 or v1.1 to v1.2, see the [v1.2 migration docs](migrating/v1.1-v1.2.md).
