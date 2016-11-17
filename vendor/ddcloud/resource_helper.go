package ddcloud

import (
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
	"strings"
)

// resourcePropertyHelper provides commonly-used functionality for working with Terraform's schema.ResourceData.
type resourcePropertyHelper struct {
	data *schema.ResourceData
}

func propertyHelper(data *schema.ResourceData) resourcePropertyHelper {
	return resourcePropertyHelper{data}
}

func (helper resourcePropertyHelper) HasProperty(key string) bool {
	_, ok := helper.data.GetOk(key)

	return ok
}

func (helper resourcePropertyHelper) GetOptionalString(key string, allowEmpty bool) *string {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case string:
		if len(typedValue) > 0 || allowEmpty {
			return &typedValue
		}
	}

	return nil
}

func (helper resourcePropertyHelper) GetOptionalInt(key string, allowZero bool) *int {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case int:
		if typedValue != 0 || allowZero {
			return &typedValue
		}
	}

	return nil
}

func (helper resourcePropertyHelper) GetOptionalBool(key string) *bool {
	value := helper.data.Get(key)
	switch typedValue := value.(type) {
	case bool:
		return &typedValue
	default:
		return nil
	}
}

func (helper resourcePropertyHelper) GetStringSetItems(key string) (items []string) {
	value, ok := helper.data.GetOk(key)
	if !ok || value == nil {
		return
	}
	rawItems := value.(*schema.Set).List()

	items = make([]string, len(rawItems))
	for index, item := range rawItems {
		items[index] = item.(string)
	}

	return
}

func (helper resourcePropertyHelper) SetStringSetItems(key string, items []string) error {
	rawItems := make([]interface{}, len(items))
	for index, item := range items {
		rawItems[index] = item
	}

	return helper.data.Set(key,
		schema.NewSet(schema.HashString, rawItems),
	)
}

func (helper resourcePropertyHelper) GetIntSetItems(key string) (items []int) {
	value, ok := helper.data.GetOk(key)
	if !ok || value == nil {
		return
	}
	rawItems := value.(*schema.Set).List()

	items = make([]int, len(rawItems))
	for index, item := range rawItems {
		items[index] = item.(int)
	}

	return
}

func (helper resourcePropertyHelper) SetIntSetItems(key string, items []int) error {
	rawItems := make([]interface{}, len(items))
	for index, item := range items {
		rawItems[index] = item
	}

	hashInt := func(value interface{}) int {
		return value.(int)
	}
	return helper.data.Set(key,
		schema.NewSet(hashInt, rawItems),
	)
}

func (helper resourcePropertyHelper) GetStringListItems(key string) (items []string) {
	value, ok := helper.data.GetOk(key)
	if !ok || value == nil {
		return
	}

	rawItems := value.([]interface{})
	items = make([]string, len(rawItems))
	for index, item := range rawItems {
		items[index] = item.(string)
	}

	return
}

func (helper resourcePropertyHelper) SetStringListItems(key string, items []string) error {
	rawItems := make([]interface{}, len(items))
	for index, item := range items {
		rawItems[index] = item
	}

	return helper.data.Set(key, rawItems)
}

func (helper resourcePropertyHelper) SetPartial(key string) {
	helper.data.SetPartial(key)
}

func (helper resourcePropertyHelper) GetTags(key string) (tags []compute.Tag) {
	value, ok := helper.data.GetOk(key)
	if !ok {
		return
	}
	tagData := value.(*schema.Set).List()

	tags = make([]compute.Tag, len(tagData))
	for index, item := range tagData {
		tagProperties := item.(map[string]interface{})
		tag := &compute.Tag{}

		value, ok = tagProperties[resourceKeyServerTagName]
		if ok {
			tag.Name = value.(string)
		}

		value, ok = tagProperties[resourceKeyServerTagValue]
		if ok {
			tag.Value = value.(string)
		}

		tags[index] = *tag
	}

	return
}

func (helper resourcePropertyHelper) SetTags(key string, tags []compute.Tag) {
	tagProperties := &schema.Set{F: hashServerTag}

	for _, tag := range tags {
		tagProperties.Add(map[string]interface{}{
			resourceKeyServerTagName:  tag.Name,
			resourceKeyServerTagValue: tag.Value,
		})
	}
	helper.data.Set(key, tagProperties)
}

func (helper resourcePropertyHelper) GetAddressListAddresses() (addresses []compute.IPAddressListEntry) {
	value, ok := helper.data.GetOk(resourceKeyAddressListAddresses)
	if !ok {
		return
	}
	portListAddresses := value.([]interface{})

	addresses = make([]compute.IPAddressListEntry, len(portListAddresses))
	for index, item := range portListAddresses {
		entryProperties := item.(map[string]interface{})
		entry := &compute.IPAddressListEntry{}

		value, ok := entryProperties[resourceKeyAddressListAddressBegin]
		if ok {
			begin := value.(string)
			if len(begin) > 0 {
				log.Printf("Have address Begin '%s'", begin)
				entry.Begin = value.(string)

				value, ok = entryProperties[resourceKeyAddressListAddressEnd]
				if ok {
					endAddress := value.(string)
					log.Printf("Have address End '%s'", endAddress)
					if endAddress != "" {
						entry.End = &endAddress
					}
				}
			}
		}

		value, ok = entryProperties[resourceKeyAddressListAddressNetwork]
		if ok {
			network := value.(string)
			if len(network) > 0 {
				entry.Begin = network
				log.Printf("Have address Network '%s'", entry.Begin)

				value, ok = entryProperties[resourceKeyAddressListAddressPrefixSize]
				if ok {
					prefixSize := value.(int)
					log.Printf("Have address PrefixSize '%d'", prefixSize)
					entry.PrefixSize = &prefixSize
				}
			}
		}

		addresses[index] = *entry
	}

	return
}

func (helper resourcePropertyHelper) SetAddressListAddresses(addresses []compute.IPAddressListEntry) {
	addressProperties := make([]interface{}, len(addresses))
	for index, address := range addresses {
		if address.PrefixSize == nil {
			addressProperties[index] = map[string]interface{}{
				resourceKeyAddressListAddressBegin: address.Begin,
				resourceKeyAddressListAddressEnd:   address.End,
			}
		} else {
			addressProperties[index] = map[string]interface{}{
				resourceKeyAddressListAddressNetwork:    address.Begin,
				resourceKeyAddressListAddressPrefixSize: *address.PrefixSize,
			}
		}
	}

	helper.data.Set(resourceKeyAddressListAddresses, addressProperties)
}

func (helper resourcePropertyHelper) GetPortListPorts() (ports []compute.PortListEntry) {
	value, ok := helper.data.GetOk(resourceKeyPortListPort)
	if !ok {
		return
	}
	portListPorts := value.([]interface{})

	ports = make([]compute.PortListEntry, len(portListPorts))
	for index, item := range portListPorts {
		portProperties := item.(map[string]interface{})
		port := &compute.PortListEntry{}

		value, ok := portProperties[resourceKeyPortListPortBegin]
		if ok {
			port.Begin = value.(int)
		}

		value, ok = portProperties[resourceKeyPortListPortEnd]
		if ok {
			endPort := value.(int)
			if endPort != 0 {
				port.End = &endPort
			}
		}

		ports[index] = *port
	}

	return
}

func (helper resourcePropertyHelper) SetPortListPorts(ports []compute.PortListEntry) {
	portProperties := make([]interface{}, len(ports))
	for index, port := range ports {
		portProperties[index] = map[string]interface{}{
			resourceKeyPortListPortBegin: port.Begin,
			resourceKeyPortListPortEnd:   port.End,
		}
	}

	helper.data.Set(resourceKeyPortListPort, portProperties)
}

func (helper resourcePropertyHelper) GetServerDisks() (disks []compute.VirtualMachineDisk) {
	value, ok := helper.data.GetOk(resourceKeyServerDisk)
	if !ok {
		return
	}
	serverDisks := value.(*schema.Set).List()

	disks = make([]compute.VirtualMachineDisk, len(serverDisks))
	for index, item := range serverDisks {
		diskProperties := item.(map[string]interface{})
		disk := &compute.VirtualMachineDisk{}

		value, ok = diskProperties[resourceKeyServerDiskID]
		if ok {
			disk.ID = stringToPtr(value.(string))
		}

		value, ok = diskProperties[resourceKeyServerDiskUnitID]
		if ok {
			disk.SCSIUnitID = value.(int)

		}
		value, ok = diskProperties[resourceKeyServerDiskSizeGB]
		if ok {
			disk.SizeGB = value.(int)
		}

		value, ok = diskProperties[resourceKeyServerDiskSpeed]
		if ok {
			disk.Speed = value.(string)
		}

		disks[index] = *disk
	}

	return
}

func (helper resourcePropertyHelper) SetServerDisks(disks []compute.VirtualMachineDisk) {
	diskProperties := &schema.Set{F: hashDisk}

	for _, disk := range disks {
		diskProperties.Add(map[string]interface{}{
			resourceKeyServerDiskID:     *disk.ID,
			resourceKeyServerDiskSizeGB: disk.SizeGB,
			resourceKeyServerDiskUnitID: disk.SCSIUnitID,
			resourceKeyServerDiskSpeed:  disk.Speed,
		})
	}
	helper.data.Set(resourceKeyServerDisk, diskProperties)
}

func (helper resourcePropertyHelper) GetVirtualListenerIRuleIDs(apiClient *compute.Client) (iRuleIDs []string, err error) {
	var iRules []compute.EntityReference
	iRules, err = helper.GetVirtualListenerIRules(apiClient)
	if err != nil {
		return
	}

	iRuleIDs = make([]string, len(iRules))
	for index, iRule := range iRules {
		iRuleIDs[index] = iRule.ID
	}

	return
}

func (helper resourcePropertyHelper) GetVirtualListenerIRuleNames(apiClient *compute.Client) (iRuleNames []string, err error) {
	var iRules []compute.EntityReference
	iRules, err = helper.GetVirtualListenerIRules(apiClient)
	if err != nil {
		return
	}

	iRuleNames = make([]string, len(iRules))
	for index, iRule := range iRules {
		iRuleNames[index] = iRule.Name
	}

	return
}

func (helper resourcePropertyHelper) GetVirtualListenerIRules(apiClient *compute.Client) (iRules []compute.EntityReference, err error) {
	value, ok := helper.data.GetOk(resourceKeyVirtualListenerIRuleNames)
	if !ok {
		return
	}
	iRuleNames := value.(*schema.Set)
	if iRuleNames.Len() == 0 {
		return
	}

	networkDomainID := helper.data.Get(resourceKeyVirtualListenerNetworkDomainID).(string)

	page := compute.DefaultPaging()
	for {
		var results *compute.IRules
		results, err = apiClient.ListDefaultIRules(networkDomainID, page)
		if err != nil {
			return
		}
		if results.IsEmpty() {
			break // We're done
		}

		for _, iRule := range results.Items {
			if iRuleNames.Contains(iRule.Name) {
				iRules = append(iRules, iRule.ToEntityReference())
			}
		}

		page.Next()
	}

	return
}

func (helper resourcePropertyHelper) SetVirtualListenerIRules(iRuleSummaries []compute.EntityReference) {
	iRuleNames := &schema.Set{F: schema.HashString}

	for _, iRuleSummary := range iRuleSummaries {
		iRuleNames.Add(iRuleSummary.Name)
	}

	helper.data.Set(resourceKeyVirtualListenerIRuleNames, iRuleNames)
}

func (helper resourcePropertyHelper) GetVirtualListenerPersistenceProfileID(apiClient *compute.Client) (persistenceProfileID *string, err error) {
	persistenceProfile, err := helper.GetVirtualListenerPersistenceProfile(apiClient)
	if err != nil {
		return nil, err
	}

	if persistenceProfile != nil {
		return &persistenceProfile.ID, nil
	}

	return nil, nil
}

func (helper resourcePropertyHelper) GetVirtualListenerPersistenceProfile(apiClient *compute.Client) (persistenceProfile *compute.EntityReference, err error) {
	value, ok := helper.data.GetOk(resourceKeyVirtualListenerPersistenceProfileName)
	if !ok {
		return
	}
	persistenceProfileName := value.(string)

	networkDomainID := helper.data.Get(resourceKeyVirtualListenerNetworkDomainID).(string)

	page := compute.DefaultPaging()
	for {
		var persistenceProfiles *compute.PersistenceProfiles
		persistenceProfiles, err = apiClient.ListDefaultPersistenceProfiles(networkDomainID, page)
		if err != nil {
			return
		}
		if persistenceProfiles.IsEmpty() {
			break // We're done
		}

		for _, profile := range persistenceProfiles.Items {
			if profile.Name == persistenceProfileName {
				persistenceProfileReference := profile.ToEntityReference()
				persistenceProfile = &persistenceProfileReference

				return
			}
		}

		page.Next()
	}

	return
}

func (helper resourcePropertyHelper) SetVirtualListenerPersistenceProfile(persistenceProfile compute.EntityReference) (err error) {
	return helper.data.Set(resourceKeyVirtualListenerPersistenceProfileName, persistenceProfile.Name)
}

func normalizeSpeed(value interface{}) string {
	speed := value.(string)

	return strings.ToUpper(speed)
}

func normalizeVIPMemberPort(port *int) string {
	if port != nil {
		return strconv.Itoa(*port)
	}

	return "ANY"
}
