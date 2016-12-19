package ddcloud

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"strconv"

	"github.com/hashicorp/terraform/terraform"
)

// Migrate state for ddcloud_server.
func resourceServerMigrateState(schemaVersion int, instanceState *terraform.InstanceState, provider interface{}) (migratedState *terraform.InstanceState, err error) {
	if instanceState.Empty() {
		log.Println("Empty Server state; nothing to migrate.")
		migratedState = instanceState

		return
	}

	const currentSchemaVersion = 3
	for schemaVersion < currentSchemaVersion {
		switch schemaVersion {
		case 0:
			log.Println("Found Server state v0; migrating to v1")
			migratedState, err = migrateServerStateV0toV1(instanceState)
		case 1:
			log.Println("Found Server state v1; migrating to v2")
			migratedState, err = migrateServerStateV1toV2(instanceState)
		case 2:
			log.Println("Found Server state v2; migrating to v3")
			migratedState, err = migrateServerStateV2toV3(instanceState)
		case 3:
			log.Println("Found Server state v3; migrating to v4")
			migratedState, err = migrateServerStateV3toV4(instanceState)
		default:
			err = fmt.Errorf("Unexpected schema version: %d", schemaVersion)
		}
		if err != nil {
			return
		}

		schemaVersion++
	}

	return
}

// Migrate state for ddcloud_server (v0 to v1).
//
// disk.HASH.xxx -> disk.INDEX.xxx (where INDEX is the 0-based index of the disk in the set).
// Note that we should really be sorting disks by SCSI unit Id, but that's a little complicated for now.
//
// Also, os_image_id / os_image_name / customer_image_id / customer_image_name -> image, image_type
func migrateServerStateV0toV1(instanceState *terraform.InstanceState) (migratedState *terraform.InstanceState, err error) {
	migratedState = instanceState

	// Convert disks from Set ("disk.HASH.property") to List ("disk.INDEX.property")
	//
	// Where INDEX is the 0-based index of the disk in the set.
	var keys []string
	for key := range migratedState.Attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	nextIndex := 0
	diskIndexesByHash := make(map[string]int)
	for _, key := range keys {
		if !strings.HasPrefix(key, "disk.") {
			continue
		}

		// Should be "disk.HASH.property".
		keyParts := strings.Split(key, ".")
		if len(keyParts) != 3 {
			continue
		}
		hash := keyParts[1]

		diskIndex, ok := diskIndexesByHash[hash]
		if !ok {
			nextIndex++
			diskIndex = nextIndex
			diskIndexesByHash[hash] = diskIndex
		}

		value := migratedState.Attributes[key]
		delete(migratedState.Attributes, key)

		// Convert to "disk.N.property"
		keyParts[1] = strconv.Itoa(diskIndex)
		key = strings.Join(keyParts, ".")
		migratedState.Attributes[key] = value
	}

	log.Printf("Server attributes after migration from v0 to v1: %#v",
		migratedState.Attributes,
	)

	return
}

// Migrate state for ddcloud_server (v1 to v2).
//
// os_image_id         = "xxx" -> image { id   = "xxx", type = "os" }
// os_image_name       = "xxx" -> image { name = "xxx", type = "os" }
// customer_image_id   = "xxx" -> image { id   = "xxx", type = "customer" }
// customer_image_name = "xxx" -> image { name = "xxx", type = "customer" }
func migrateServerStateV1toV2(instanceState *terraform.InstanceState) (migratedState *terraform.InstanceState, err error) {
	migratedState = instanceState

	var (
		osImageID         string
		osImageName       string
		customerImageID   string
		customerImageName string
	)
	osImageID, _ = migratedState.Attributes[resourceKeyServerOSImageID]
	delete(migratedState.Attributes, resourceKeyServerOSImageID)
	osImageName, _ = migratedState.Attributes[resourceKeyServerOSImageName]
	delete(migratedState.Attributes, resourceKeyServerOSImageName)
	customerImageID, _ = migratedState.Attributes[resourceKeyServerCustomerImageID]
	delete(migratedState.Attributes, resourceKeyServerCustomerImageID)
	customerImageName, _ = migratedState.Attributes[resourceKeyServerCustomerImageName]
	delete(migratedState.Attributes, resourceKeyServerCustomerImageName)

	// Single-item list.
	migratedState.Attributes[resourceKeyServerImage+".#"] = "1"
	setImageProperty := func(propertyName string, propertyValue string) {
		key := fmt.Sprintf("%s.0.%s",
			resourceKeyServerImage,
			propertyName,
		)

		migratedState.Attributes[key] = propertyValue
	}

	if osImageID != "" {
		setImageProperty("id", osImageID)
		setImageProperty(resourceKeyServerImageType, "os")
	} else if osImageName != "" {
		setImageProperty("name", osImageName)
		setImageProperty(resourceKeyServerImageType, "os")
	} else if customerImageID != "" {
		setImageProperty("id", customerImageID)
		setImageProperty(resourceKeyServerImageType, "customer")
	} else if customerImageName != "" {
		setImageProperty("name", customerImageName)
		setImageProperty(resourceKeyServerImageType, "customer")
	}

	log.Printf("Server attributes after migration from v1 to v2: %#v",
		migratedState.Attributes,
	)

	return
}

// Migrate state for ddcloud_server (v2 to v3).
//
// image { id   = "xxx", type = "os" }       -> image = "xxx"
// image { name = "xxx", type = "os" }       -> image = "xxx"
// image { id   = "xxx", type = "customer" } -> image = "xxx", image_type = "customer"
// image { name = "xxx", type = "customer" } -> image = "xxx", image_type = "customer"
func migrateServerStateV2toV3(instanceState *terraform.InstanceState) (migratedState *terraform.InstanceState, err error) {
	migratedState = instanceState

	const keyPrefix = "image."
	var image, imageType string
	for key := range migratedState.Attributes {
		if !strings.HasPrefix(key, keyPrefix) {
			continue
		}

		value := migratedState.Attributes[key]
		delete(migratedState.Attributes, key)

		if key == keyPrefix+"#" { // "image.#"
			continue // Count - nothing else to do.
		}

		switch key[len(keyPrefix+"0")+1:] { // "image.0."
		case "id":
		case "name":
			image = value
		case resourceKeyServerImageType:
			imageType = value
		}
	}

	if image != "" {
		migratedState.Attributes[resourceKeyServerImage] = image
		if imageType == "customer" {
			migratedState.Attributes[resourceKeyServerImageType] = "customer"
		}
	}

	log.Printf("Server attributes after migration from v2 to v3: %#v",
		migratedState.Attributes,
	)

	return
}

// Migrate state for ddcloud_server (v3 to v4).
//
// From:
// primary_adapter_ipv4 = "xxx"
// primary_adapter_vlan = "yyy"
// primary_adapter_type = "zzz"
//
// To:
// primary_network_adapter {
//     ipv4 = "xxx"
//     vlan = "yyy"
//     type = "zzz"
// }
func migrateServerStateV3toV4(instanceState *terraform.InstanceState) (migratedState *terraform.InstanceState, err error) {
	setPrimaryAdapterProperty := func(key string, value string) {
		if key != "#" { // Count
			key = "0." + key // First element of list
		}
		migratedState.Attributes[resourceKeyServerPrimaryNetworkAdapter+"."+key] = value
	}

	migratedState = instanceState

	var primaryAdapterIPv4, primaryAdapterVLAN, primaryAdapterType string
	for key := range migratedState.Attributes {
		switch key {
		case resourceKeyServerPrimaryAdapterIPv4:
			primaryAdapterIPv4 = migratedState.Attributes[key]
		case resourceKeyServerPrimaryAdapterVLAN:
			primaryAdapterVLAN = migratedState.Attributes[key]
		case resourceKeyServerPrimaryAdapterType:
			primaryAdapterType = migratedState.Attributes[key]
		}
	}

	if primaryAdapterIPv4 != "" || primaryAdapterVLAN != "" || primaryAdapterType != "" {
		setPrimaryAdapterProperty("#", "1")
	}
	if primaryAdapterIPv4 != "" {
		setPrimaryAdapterProperty(resourceKeyServerNetworkAdapterIPV4, primaryAdapterIPv4)
	}
	if primaryAdapterVLAN != "" {
		setPrimaryAdapterProperty(resourceKeyServerNetworkAdapterVLANID, primaryAdapterVLAN)
	}
	if primaryAdapterType != "" {
		setPrimaryAdapterProperty(resourceKeyServerNetworkAdapterType, primaryAdapterType)
	}

	log.Printf("Server attributes after migration from v3 to v4: %#v",
		migratedState.Attributes,
	)

	return
}
