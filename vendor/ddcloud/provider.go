package ddcloud

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DimensionDataResearch/dd-cloud-compute-terraform/retry"
	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider creates the Dimension Data Cloud resource provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		// Provider settings schema
		Schema: map[string]*schema.Schema{
			"region": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The region code that identifies the target end-point for the Dimension Data CloudControl API.",
				ConflictsWith: []string{"cloudcontrol_endpoint"},
			},
			"cloudcontrol_endpoint": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The base URL of a custom target end-point for the Dimension Data CloudControl API.",
				ConflictsWith: []string{"region"},
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The user name used to authenticate to the Dimension Data CloudControl API (if not specified, then the MCP_USER environment variable will be used).",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Default:     "",
				Description: "The password used to authenticate to the Dimension Data CloudControl API (if not specified, then the MCP_PASSWORD environment variable will be used).",
			},
			"allow_server_reboot": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Allow rebooting of ddcloud_server instances (e.g. for adding / removing NICs)?",
			},
			"retry_timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10 * 60, // 10 minutes
				Description: "The number of seconds before retrying an operation times out.",
			},
			"retry_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "The delay, in seconds, between retries of operations that fail due to a RESOURCE_BUSY response from CloudControl.",
			},
		},

		// Provider resource definitions
		ResourcesMap: map[string]*schema.Resource{
			// A network domain.
			"ddcloud_networkdomain": resourceNetworkDomain(),

			// A VLAN.
			"ddcloud_vlan": resourceVLAN(),

			// A server (virtual machine).
			"ddcloud_server": resourceServer(),

			// An additional network interface card (NIC) in a server.
			"ddcloud_server_nic": resourceServerNIC(),

			// A server anti-affinity rule.
			"ddcloud_server_anti_affinity": resourceServerAntiAffinityRule(),

			// A Network Address Translation (NAT) rule.
			"ddcloud_nat": resourceNAT(),

			// A firewall rule.
			"ddcloud_firewall_rule": resourceFirewallRule(),

			// An IP address list.
			"ddcloud_address_list": resourceAddressList(),

			// A port list.
			"ddcloud_port_list": resourcePortList(),

			// A VIP node.
			"ddcloud_vip_node": resourceVIPNode(),

			// A VIP pool.
			"ddcloud_vip_pool": resourceVIPPool(),

			// A VIP pool member (links pool, node, and optionally port).
			"ddcloud_vip_pool_member": resourceVIPPoolMember(),

			// A virtual listener is the top-level entity for load-balancing functionality.
			"ddcloud_virtual_listener": resourceVirtualListener(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			// A network domain.
			"ddcloud_networkdomain": dataSourceNetworkDomain(),

			// A virtual network (VLAN).
			"ddcloud_vlan": dataSourceVLAN(),
		},

		// Provider configuration
		ConfigureFunc: configureProvider,
	}
}

// Configure the provider.
// Returns the provider's compute API client.
func configureProvider(providerSettings *schema.ResourceData) (interface{}, error) {
	// Log provider version (for diagnostic purposes).
	log.Print("ddcloud provider version is " + ProviderVersion)

	region := strings.ToLower(
		providerSettings.Get("region").(string),
	)
	customEndPoint := providerSettings.Get("cloudcontrol_endpoint").(string)
	if region == "" && customEndPoint == "" {
		return nil, fmt.Errorf("Neither the 'region' nor the 'cloudcontrol_endpoint' provider properties were specified (the 'ddcloud' provider requires exactly one of these properties to be configured).")
	}

	username := providerSettings.Get("username").(string)
	if isEmpty(username) {
		username = os.Getenv("MCP_USER")
		if isEmpty(username) {
			return nil, fmt.Errorf("The 'username' property was not specified for the 'ddcloud' provider, and the 'MCP_USER' environment variable is not present. Please supply either one of these to configure the user name used to authenticate to Dimension Data CloudControl.")
		}
	}

	password := providerSettings.Get("password").(string)
	if isEmpty(password) {
		password = os.Getenv("MCP_PASSWORD")
		if isEmpty(password) {
			return nil, fmt.Errorf("The 'password' property was not specified for the 'ddcloud' provider, and the 'MCP_PASSWORD' environment variable is not present. Please supply either one of these to configure the password used to authenticate to Dimension Data CloudControl.")
		}
	}

	var client *compute.Client
	if region != "" {
		client = compute.NewClient(region, username, password)
	} else {
		client = compute.NewClientWithBaseAddress(customEndPoint, username, password)
	}

	// Configure retry, if required.
	retryCount := 0
	retryDelay := 30 // Seconds

	// Override default retry configuration with environment variables, if required.
	retryValue, err := strconv.Atoi(os.Getenv("MCP_MAX_RETRY"))
	if err == nil {
		retryCount = retryValue

		retryValue, err = strconv.Atoi(os.Getenv("MCP_RETRY_DELAY"))
		if err == nil {
			retryDelay = retryValue
		}
	}
	client.ConfigureRetry(retryCount, time.Duration(retryDelay)*time.Second)

	settings := &ProviderSettings{
		RetryDelay:         time.Duration(providerSettings.Get("retry_delay").(int)) * time.Second,
		RetryTimeout:       time.Duration(providerSettings.Get("retry_timeout").(int)) * time.Second,
		AllowServerReboots: providerSettings.Get("allow_server_reboot").(bool),
	}

	// Override server reboot behaviour with environment variables, if required.
	allowRebootValue, err := strconv.ParseBool(os.Getenv("MCP_ALLOW_SERVER_REBOOT"))
	if err == nil {
		settings.AllowServerReboots = allowRebootValue
	}

	provider := newProvider(client, settings)

	return provider, nil
}

// ProviderSettings represents the configuration for the ddcloud provider.
type ProviderSettings struct {
	// Allow rebooting of ddcloud_server instances if required during an update?
	//
	// For example, servers must be rebooted to add or remove network adapters.
	AllowServerReboots bool

	// The period of time between retry attempts for asynchronous operations.
	RetryDelay time.Duration

	// The period of time before retrying of asynchronous operations time out.
	RetryTimeout time.Duration
}

type providerState struct {
	// The CloudControl API client.
	apiClient *compute.Client

	// The provider settings.
	settings *ProviderSettings

	// Global lock for provider state.
	stateLock *sync.Mutex

	// Global lock for initiating asynchronous operations.
	asyncOperationLock *sync.Mutex

	// Lock per server (prevent parallel provisioning operations for a given ddcloud_server resource).
	serverLocks map[string]*sync.Mutex

	// Provider-global retry executor for asynchronous operations.
	retry retry.Do
}

func newProvider(client *compute.Client, settings *ProviderSettings) *providerState {
	state := &providerState{
		apiClient:          client,
		settings:           settings,
		stateLock:          &sync.Mutex{},
		asyncOperationLock: &sync.Mutex{},
		serverLocks:        make(map[string]*sync.Mutex),
		retry:              retry.NewDo(settings.RetryDelay),
	}

	return state
}

// Client retrieves the CloudControl API client from provider state.
func (state *providerState) Client() *compute.Client {
	return state.apiClient
}

// Settings retrieves a copy of the provider settings.
func (state *providerState) Settings() ProviderSettings {
	return *state.settings // We return a copy because these settings should be read-only once the provider has been created.
}

// Retry retrieves the provider's operation-retry executor.
func (state *providerState) Retry() retry.Do {
	return state.retry
}

// AcquireAsyncOperationLock acquires (locks) the global lock used to synchronise initiation of global operations.
func (state *providerState) AcquireAsyncOperationLock(ownerNameOrFormat string, formatArgs ...interface{}) *asyncOperationLock {
	asyncLock := &asyncOperationLock{
		ownerName:   fmt.Sprintf(ownerNameOrFormat, formatArgs...),
		lock:        state.asyncOperationLock,
		releaseOnce: &sync.Once{},
	}

	log.Printf("%s acquiring global asynchronous operation lock...", asyncLock.ownerName)
	asyncLock.lock.Lock()
	log.Printf("%s acquired global asynchronous operation lock.", asyncLock.ownerName)

	return asyncLock
}

type asyncOperationLock struct {
	ownerName   string
	lock        *sync.Mutex
	releaseOnce *sync.Once
}

// Release the global asynchronous operation lock.
//
// Safe to call multiple times - subsequent calls to Release have no effect (call providerState.AcquireAsyncOperationLock to reacquire the lock).
func (asyncLock *asyncOperationLock) Release() {
	asyncLock.releaseOnce.Do(func() {
		log.Printf("%s acquiring global asynchronous operation lock...", asyncLock.ownerName)
		asyncLock.lock.Unlock()
		log.Printf("%s acquired global asynchronous operation lock.", asyncLock.ownerName)
	})
}

// GetServerLock retrieves the global lock for the specified server.
func (state *providerState) GetServerLock(id string, ownerNameOrFormat string, formatArgs ...interface{}) *providerServerLock {
	state.stateLock.Lock()
	defer state.stateLock.Unlock()

	lock, ok := state.serverLocks[id]
	if !ok {
		lock = &sync.Mutex{}
		state.serverLocks[id] = lock
	}

	return &providerServerLock{
		serverID:  id,
		ownerName: fmt.Sprintf(ownerNameOrFormat, formatArgs...),
		lock:      lock,
	}
}

type providerServerLock struct {
	serverID  string
	ownerName string
	lock      *sync.Mutex
}

// Acquire the server lock.
func (serverLock *providerServerLock) Lock() {
	log.Printf("%s acquiring lock for server '%s'...", serverLock.ownerName, serverLock.serverID)
	serverLock.lock.Lock()
	log.Printf("%s acquired lock for server '%s'.", serverLock.ownerName, serverLock.serverID)
}

// Release the server lock.
func (serverLock *providerServerLock) Unlock() {
	log.Printf("%s releasing lock for server '%s'...", serverLock.ownerName, serverLock.serverID)
	serverLock.lock.Unlock()
	log.Printf("%s released lock for server '%s'.", serverLock.ownerName, serverLock.serverID)
}
