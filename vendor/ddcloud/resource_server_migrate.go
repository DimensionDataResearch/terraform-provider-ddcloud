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

	switch schemaVersion {
	case 0:
		log.Println("Found Server state v0; migrating to v2")
		migratedState, err = migrateServerStateV0toV1(instanceState)
	case 1:
		log.Println("Found Server state v1; migrating to v2")
		migratedState, err = migrateServerStateV1toV2(instanceState)
	default:
		err = fmt.Errorf("Unexpected schema version: %d", schemaVersion)
	}

	return
}

// Migrate state for ddcloud_server (v0 to v1).
//
// disk.HASH.xxx -> disk.INDEX.xxx (where INDEX is the 0-based index of the disk in the set).
//
// Note that we should really be sorting disks by SCSI unit Id, but that's a little complicated for now.
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
		setImageProperty(resourceKeyServerImageID, osImageID)
		setImageProperty(resourceKeyServerImageType, "os")
	} else if osImageName != "" {
		setImageProperty(resourceKeyServerImageName, osImageName)
		setImageProperty(resourceKeyServerImageType, "os")
	} else if customerImageID != "" {
		setImageProperty(resourceKeyServerImageID, customerImageID)
		setImageProperty(resourceKeyServerImageType, "customer")
	} else if customerImageName != "" {
		setImageProperty(resourceKeyServerImageName, customerImageName)
		setImageProperty(resourceKeyServerImageType, "customer")
	}

	log.Printf("Server attributes after migration from v1 to v2: %#v",
		migratedState.Attributes,
	)

	return
}
