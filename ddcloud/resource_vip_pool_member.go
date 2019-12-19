package ddcloud

import (
	"fmt"
	"log"
	"strconv"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				ForceNew: true,
			},
			resourceKeyVIPPoolMemberPoolName: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			resourceKeyVIPPoolMemberNodeID: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			resourceKeyVIPPoolMemberNodeName: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			resourceKeyVIPPoolMemberPort: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
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

	var (
		memberID string
		err      error
	)

	operationDescription := fmt.Sprintf("Add node '%s' to VIP pool '%s'", nodeID, poolID)

	err = providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		var addError error
		memberID, addError = apiClient.AddVIPPoolMember(poolID, nodeID, status, port)
		if compute.IsResourceBusyError(addError) {
			context.Retry()
		} else if addError != nil {
			context.Fail(addError)
		}
		asyncLock.Release()
	})
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
		return fmt.Errorf("unable to find newly-created pool member '%s'", memberID)
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

	if !data.HasChange(resourceKeyVIPPoolMemberStatus) {
		return nil
	}

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	status := data.Get(resourceKeyVIPPoolMemberStatus).(string)

	operationDescription := fmt.Sprintf("Edit VIP pool member '%s'", id)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		editError := apiClient.EditVIPPoolMember(id, status)
		if compute.IsResourceBusyError(editError) {
			context.Retry()
		} else if editError != nil {
			context.Fail(editError)
		}
		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	return nil
}

func resourceVIPPoolMemberDelete(data *schema.ResourceData, provider interface{}) error {
	memberID := data.Id()
	poolID := data.Get(resourceKeyVIPPoolMemberPoolID).(string)

	log.Printf("Delete member '%s' from VIP pool '%s'.", memberID, poolID)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	operationDescription := fmt.Sprintf("Remove member '%s' from VIP pool '%s'", memberID, poolID)
	err := providerState.RetryAction(operationDescription, func(context retry.Context) {
		asyncLock := providerState.AcquireAsyncOperationLock(operationDescription)
		defer asyncLock.Release() // Released at the end of the current attempt.

		removeError := apiClient.RemoveVIPPoolMember(memberID)
		if compute.IsResourceBusyError(removeError) {
			context.Retry()
		} else if removeError != nil {
			context.Fail(removeError)
		}
		asyncLock.Release()
	})
	if err != nil {
		return err
	}

	return nil
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
