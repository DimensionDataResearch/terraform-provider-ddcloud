package ddcloud

import (
	"fmt"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
)

const (
	resourceKeyVIPPoolMemberPoolID   = "pool"
	resourceKeyVIPPoolMemberPoolName = "pool_name"
	resourceKeyVIPPoolMemberNodeID   = "node"
	resourceKeyVIPPoolMemberNodeName = "node_name"
	resourceKeyVIPPoolMemberPort     = "port"
	resourceKeyVIPPoolMemberStatus   = "status"
)

func resourceVIPPoolMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceVIPPoolMemberCreate,
		Read:   resourceVIPPoolMemberRead,
		Exists: resourceVIPPoolMemberExists,
		Update: resourceVIPPoolMemberUpdate,
		Delete: resourceVIPPoolMemberDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyVIPPoolMemberPoolID: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			resourceKeyVIPPoolMemberPoolName: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			resourceKeyVIPPoolMemberNodeID: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			resourceKeyVIPPoolMemberNodeName: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			resourceKeyVIPPoolMemberPort: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ANY",
			},
			resourceKeyVIPPoolMemberStatus: &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      compute.VIPNodeStatusEnabled,
				ValidateFunc: vipStatusValidator("VIP pool member"),
			},
		},
	}
}

func resourceVIPPoolMemberCreate(data *schema.ResourceData, provider interface{}) error {
	propertyHelper := propertyHelper(data)

	nodeID := data.Get(resourceKeyVIPPoolMemberNodeID).(string)
	poolID := data.Get(resourceKeyVIPPoolMemberPoolID).(string)
	port := propertyHelper.GetOptionalInt(resourceKeyVIPPoolMemberPort, false)
	status := data.Get(resourceKeyVIPPoolMemberStatus).(string)

	log.Printf("Add node '%s' as a member of VIP pool '%s'.", nodeID, poolID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	memberID, err := apiClient.AddVIPPoolMember(poolID, nodeID, status, port)
	if err != nil {
		return err
	}

	data.SetId(memberID)

	log.Printf("Successfully added node '%s' to VIP pool '%s' as member '%s'.", nodeID, poolID, memberID)

	member, err := apiClient.GetVIPPoolMember(memberID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("Unable to find newly-created pool member '%s'.", memberID)
	}

	data.Set(resourceKeyVIPPoolMemberNodeName, member.Node.Name)
	data.Set(resourceKeyVIPPoolMemberPoolName, member.Pool.Name)
	data.Set(resourceKeyVIPPoolMemberStatus, member.Status)

	return nil
}

func resourceVIPPoolMemberExists(data *schema.ResourceData, provider interface{}) (bool, error) {
	id := data.Id()

	log.Printf("Check if VIP pool member '%s' exists...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	member, err := apiClient.GetVIPPoolMember(id)
	if err != nil {
		return false, err
	}

	exists := member != nil

	log.Printf("VIP pool member '%s' exists: %t.", id, exists)

	return exists, nil
}

func resourceVIPPoolMemberRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()

	log.Printf("Read VIP pool '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	member, err := apiClient.GetVIPPoolMember(id)
	if err != nil {
		return err
	}
	if member == nil {
		data.SetId("") // VIP pool member has been deleted

		return nil
	}

	data.Set(resourceKeyVIPPoolMemberNodeName, member.Node.Name)
	data.Set(resourceKeyVIPPoolMemberPoolName, member.Pool.Name)
	data.Set(resourceKeyVIPPoolMemberStatus, member.Status)

	return nil
}

func resourceVIPPoolMemberUpdate(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	log.Printf("Update VIP pool member '%s'...", id)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	if !data.HasChange(resourceKeyVIPPoolMemberStatus) {
		return nil
	}

	status := data.Get(resourceKeyVIPPoolMemberStatus).(string)

	return apiClient.EditVIPPoolMember(id, status)
}

func resourceVIPPoolMemberDelete(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	poolID := data.Get(resourceKeyVIPPoolMemberPoolID).(string)

	log.Printf("Delete member '%s' from VIP pool '%s'.", id, poolID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	return apiClient.RemoveVIPPoolMember(id)
}

func hashVIPPoolMember(item interface{}) int {
	member, ok := item.(compute.VIPPoolMember)
	if ok {
		port := "ANY"
		if member.Port != nil {
			port = strconv.Itoa(*member.Port)
		}

		return schema.HashString(fmt.Sprintf(
			"%s/%s/%s/%s",
			member.Pool.ID,
			member.Node.ID,
			port,
			member.Status,
		))
	}

	memberData := item.(map[string]interface{})

	return schema.HashString(fmt.Sprintf(
		"%s/%s/%s/%s",
		memberData[resourceKeyVIPPoolMemberPoolID].(string),
		memberData[resourceKeyVIPPoolMemberNodeID].(string),
		memberData[resourceKeyVIPPoolMemberPort].(string),
		memberData[resourceKeyVIPPoolMemberStatus].(string),
	))
}
