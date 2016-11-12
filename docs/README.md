# Managed Cloud Platform

The Dimension Data Managed Cloud Platform provider is used to interact with Dimension Data's Managed Cloud Platform resources.

## Provider configuration

See the documentation for [provider configuration](provider.md).

## Resource types

The `ddcloud` provider supports the following resource types:

* [ddcloud_networkdomain](resource_types/networkdomain.md) - A CloudControl network domain.
* [ddcloud_vlan](resource_types/vlan.md) - A CloudControl Virtual LAN (VLAN).
* [ddcloud_server](resource_types/server.md) - A CloudControl Server (virtual machine).
* [ddcloud_server_nic](resource_types/server_nic.md) - An additional network interface card (NIC) for a CloudControl Server.
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

And the following data-source types:

* [ddcloud_networkdomain](datasource_types/networkdomain.md) - A CloudControl network domain (lookup by name and data centre).
* [ddcloud_vlan](datasource_types/vlan.md) - A CloudControl Virtual LAN (VLAN) (lookup by name and network domain).
