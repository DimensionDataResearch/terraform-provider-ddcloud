package ddcloud

import (
	"fmt"
	"log"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/validators"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyIPAddressReservationVLANID      = "vlan"
	resourceKeyIPAddressReservationAddress     = "address"
	resourceKeyIPAddressReservationAddressType = "address_type"

	addressTypeIPv4 = "ipv4"
	addressTypeIPv6 = "ipv6"
)

func resourceIPAddressReservation() *schema.Resource {
	return &schema.Resource{
		Exists: resourceIPAddressReservationExists,
		Create: resourceIPAddressReservationCreate,
		Read:   resourceIPAddressReservationRead,
		Delete: resourceIPAddressReservationDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyIPAddressReservationVLANID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Id of the VLAN in which the IP address is reserved.",
			},
			resourceKeyIPAddressReservationAddress: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The reserved IP address.",
			},
			resourceKeyIPAddressReservationAddressType: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The reserved IP address type ('ipv4' or 'ipv6').",
				ValidateFunc: validators.StringIsOneOf("IP address type",
					addressTypeIPv4,
					addressTypeIPv6,
				),
			},
		},
	}
}

func resourceIPAddressReservationExists(data *schema.ResourceData, provider interface{}) (exists bool, err error) {
	vlanID := data.Get(resourceKeyIPAddressReservationVLANID).(string)
	address := data.Get(resourceKeyIPAddressReservationAddress).(string)
	addressType := data.Get(resourceKeyIPAddressReservationAddressType).(string)

	log.Printf("Check if IP address '%s' ('%s') is reserved in VLAN '%s'...",
		address,
		addressType,
		vlanID,
	)

	providerState := provider.(*providerState)

	var reservedIPAddresses map[string]compute.ReservedIPAddress
	reservedIPAddresses, err = getReservedIPAddresses(vlanID, addressType, providerState)
	if err != nil {
		return
	}

	_, exists = reservedIPAddresses[address]

	if exists {
		log.Printf("IP address '%s' ('%s') is reserved in VLAN '%s'.",
			address,
			addressType,
			vlanID,
		)
	} else {
		log.Printf("IP address '%s' ('%s') is not reserved in VLAN '%s'.",
			address,
			addressType,
			vlanID,
		)
	}

	return
}

func resourceIPAddressReservationCreate(data *schema.ResourceData, provider interface{}) (err error) {
	vlanID := data.Get(resourceKeyIPAddressReservationVLANID).(string)
	address := data.Get(resourceKeyIPAddressReservationAddress).(string)
	addressType := data.Get(resourceKeyIPAddressReservationAddressType).(string)

	log.Printf("Reserving IP address '%s' ('%s') in VLAN '%s'...",
		address,
		addressType,
		vlanID,
	)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	switch addressType {
	case addressTypeIPv4:
		err = apiClient.ReservePrivateIPv4Address(vlanID, address)
	case addressTypeIPv6:
		err = apiClient.ReserveIPv6Address(vlanID, address)
	default:
		err = fmt.Errorf("Invalid address type '%s'", addressType)
	}

	if err != nil {
		return
	}

	data.SetId(fmt.Sprintf("%s/%s",
		vlanID, addressType,
	))

	log.Printf("Reserved IP address '%s' ('%s') in VLAN '%s'.",
		address,
		addressType,
		vlanID,
	)

	return
}

func resourceIPAddressReservationRead(data *schema.ResourceData, provider interface{}) error {
	// Nothing to do.

	return nil
}

func resourceIPAddressReservationUpdate(data *schema.ResourceData, provider interface{}) error {
	// Nothing to do.

	return nil
}

func resourceIPAddressReservationDelete(data *schema.ResourceData, provider interface{}) (err error) {
	vlanID := data.Get(resourceKeyIPAddressReservationVLANID).(string)
	address := data.Get(resourceKeyIPAddressReservationAddress).(string)
	addressType := data.Get(resourceKeyIPAddressReservationAddressType).(string)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	switch addressType {
	case addressTypeIPv4:
		err = apiClient.UnreservePrivateIPv4Address(vlanID, address)
	case addressTypeIPv6:
		err = apiClient.UnreserveIPv6Address(vlanID, address)
	default:
		err = fmt.Errorf("Invalid address type '%s'", addressType)
	}

	if err != nil {
		return
	}

	data.SetId("")

	return
}

func getReservedIPAddresses(vlanID string, addressType string, providerState *providerState) (map[string]compute.ReservedIPAddress, error) {
	switch addressType {
	case addressTypeIPv4:
		return getReservedPrivateIPv4Addresses(vlanID, providerState)
	case addressTypeIPv6:
		return getReservedIPv6Addresses(vlanID, providerState)
	default:
		return nil, fmt.Errorf("Invalid address type '%s'", addressType)
	}
}

func getReservedPrivateIPv4Addresses(vlanID string, providerState *providerState) (map[string]compute.ReservedIPAddress, error) {
	apiClient := providerState.Client()

	reservedIPAddresses := make(map[string]compute.ReservedIPAddress)

	page := compute.DefaultPaging()
	for {
		reservations, err := apiClient.ListReservedPrivateIPv4AddressesInVLAN(vlanID)
		if err != nil {
			return nil, err
		}
		if reservations.IsEmpty() {
			break
		}

		for _, reservedIPAddress := range reservations.Items {
			reservedIPAddresses[reservedIPAddress.IPAddress] = reservedIPAddress
		}

		page.Next()
	}

	return reservedIPAddresses, nil
}

func getReservedIPv6Addresses(vlanID string, providerState *providerState) (map[string]compute.ReservedIPAddress, error) {
	apiClient := providerState.Client()

	reservedIPAddresses := make(map[string]compute.ReservedIPAddress)

	page := compute.DefaultPaging()
	for {
		reservations, err := apiClient.ListReservedIPv6AddressesInVLAN(vlanID)
		if err != nil {
			return nil, err
		}
		if reservations.IsEmpty() {
			break
		}

		for _, reservedIPAddress := range reservations.Items {
			reservedIPAddresses[reservedIPAddress.IPAddress] = reservedIPAddress
		}

		page.Next()
	}

	return reservedIPAddresses, nil
}
