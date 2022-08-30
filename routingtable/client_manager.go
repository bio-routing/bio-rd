package routingtable

import (
	"sync"
)

type ClientManagerMaster interface {
	UpdateNewClient(RouteTableClient) error
}

// ClientOptions represents options for a client
type ClientOptions struct {
	BestOnly bool
	EcmpOnly bool
	MaxPaths uint
}

// GetMaxPaths calculates the maximum amount of wanted paths given that ecmpPaths paths exist
func (c *ClientOptions) GetMaxPaths(ecmpPaths uint) uint {
	if c.BestOnly {
		return 1
	}

	if c.EcmpOnly {
		return ecmpPaths
	}

	return c.MaxPaths
}

// ClientManager manages clients of routing tables (observer pattern)
type ClientManager struct {
	clients   map[RouteTableClient]ClientOptions
	master    ClientManagerMaster
	mu        sync.RWMutex
	endOfLife bool // do not accept new clients when EOL
}

// NewClientManager creates and initializes a new client manager
func NewClientManager(master ClientManagerMaster) *ClientManager {
	return &ClientManager{
		clients: make(map[RouteTableClient]ClientOptions, 0),
		master:  master,
	}
}

// ClientCount gets the number of registred clients
func (c *ClientManager) ClientCount() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return uint64(len(c.clients))
}

// GetOptions gets the options for a registered client
func (c *ClientManager) GetOptions(client RouteTableClient) ClientOptions {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clients[client]
}

// RegisterWithOptions registers a client with options for updates
func (c *ClientManager) RegisterWithOptions(client RouteTableClient, opt ClientOptions) {
	c.mu.Lock()

	if c.endOfLife {
		return
	}

	c.clients[client] = opt
	c.mu.Unlock()

	c.master.UpdateNewClient(client)
}

// Unregister unregisters a client
func (c *ClientManager) Unregister(client RouteTableClient) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c._unregister(client)
}

func (c *ClientManager) _unregister(client RouteTableClient) bool {
	if _, ok := c.clients[client]; !ok {
		return false
	}
	delete(c.clients, client)
	return true
}

// Clients returns a list of registered clients
func (c *ClientManager) Clients() []RouteTableClient {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ret := make([]RouteTableClient, 0)
	for rtc := range c.clients {
		ret = append(ret, rtc)
	}

	return ret
}

func (c *ClientManager) Dispose() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.endOfLife = true

	for cli := range c.clients {
		c._unregister(cli)
	}
}
