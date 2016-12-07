#!/usr/bin/env python3

import json
import os
from os import path

"""
Migration script for terraform.tfstate (terraform-provider-ddcloud)
v1.1 -> v1.2
"""


def load_state():
    """
    Load state data from terraform.tfstate.
    """

    state_file_path = path.join(os.getcwd(), "terraform.tfstate")
    with open(state_file_path) as state_file:
        return json.load(state_file)


def get_resource_instances(state_data):
    """
    Get instance data for all resources declared in the state data.

    :param: state_data The Terraform state data.
    :returns: A sequence of (resource_type, resource_name, resource) tuples.
    """

    for module in state_data['modules']:
        resources = module['resources']
        for resource_key in sorted(resources.keys()):
            resource = resources[resource_key]
            resource_type = resource['type']
            resource_name = resource_key.split('.', maxsplit=1)[1]

            yield resource_type, resource_name, resource


def migrate_ddcloud_server_nics(state_data):
    """
    Migrate state for ddcloud_server_nics to a ddcloud_network_adapters.

    :param: state_data The Terraform state data.
    """

    for module in state_data['modules']:
        resources = module['resources']

        ddcloud_server_nic_keys = [
            resource_key for resource_key in resources.keys()
            if resource_key.startswith('ddcloud_server_nic.')
        ]

        for ddcloud_server_nic_key in ddcloud_server_nic_keys:
            resource = resources.pop(ddcloud_server_nic_key)
            resource['type'] = 'ddcloud_network_adapter'

            ddcloud_network_adapter_key = ddcloud_server_nic_key.replace(
                'ddcloud_server_nic.',
                'ddcloud_network_adapter.'
            )
            resources[ddcloud_network_adapter_key] = resource


if __name__ == "__main__":
    state_data = load_state()
    migrate_ddcloud_server_nics(state_data)

    resources = get_resource_instances(state_data)
    for resource_type, resource_name, resource in resources:
        print("{}: {}".format(resource_type, resource_name))

    print('')
    print(json.dumps(state_data, indent='  '))
