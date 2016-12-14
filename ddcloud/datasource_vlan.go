package ddcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func dataSourceVLAN() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVLANRead,

		Schema: map[string]*schema.Schema{
			resourceKeyVLANName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the target VLAN",
			},
			resourceKeyVLANNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Id of the network domain that contains the target VLAN",
			},
			resourceKeyVLANDescription: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of the VLAN",
			},
			resourceKeyVLANIPv4BaseAddress: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VLAN's private IPv4 base address.",
			},
			resourceKeyVLANIPv4PrefixSize: &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VLAN's private IPv4 prefix length.",
			},
			resourceKeyVLANIPv6BaseAddress: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VLAN's IPv6 base address.",
			},
			resourceKeyVLANIPv6PrefixSize: &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VLAN's IPv6 prefix length.",
			},
		},
	}
}

// Read a network domain data source.
func dataSourceVLANRead(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyVLANName).(string)
	networkDomainID := data.Get(resourceKeyVLANNetworkDomainID).(string)

	log.Printf("Read VLAN '%s' in network domain '%s'.", name, networkDomainID)

	apiClient := provider.(*providerState).Client()

	vlan, err := apiClient.GetVLANByName(name, networkDomainID)
	if err != nil {
		return err
	}

	if vlan != nil {
		data.SetId(vlan.ID)
		data.Set(resourceKeyVLANIPv4BaseAddress, vlan.IPv4Range.BaseAddress)
		data.Set(resourceKeyVLANIPv4PrefixSize, vlan.IPv4Range.PrefixSize)
		data.Set(resourceKeyVLANIPv6BaseAddress, vlan.IPv6Range.BaseAddress)
		data.Set(resourceKeyVLANIPv6PrefixSize, vlan.IPv6Range.PrefixSize)
	} else {
		data.SetId("") // Mark resource as deleted.
	}

	return nil
}
