package routingtable

import (
	"sync"
)

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
	clients map[RouteTableClient]ClientOptions
	master  RouteTableClient
	mu      sync.RWMutex
}

// NewClientManager creates and initializes a new client manager
func NewClientManager(master RouteTableClient) ClientManager {
	return ClientManager{
		clients: make(map[RouteTableClient]ClientOptions, 0),
		master:  master,
	}
}

// GetOptions gets the options for a registred client
func (c *ClientManager) GetOptions(client RouteTableClient) ClientOptions {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clients[client]
}

// Register registers a client for updates
func (c *ClientManager) Register(client RouteTableClient) {
	c.RegisterWithOptions(client, ClientOptions{BestOnly: true})
}

// RegisterWithOptions registers a client with options for updates
func (c *ClientManager) RegisterWithOptions(client RouteTableClient, opt ClientOptions) {
	c.mu.Lock()
	c.clients[client] = opt
	c.mu.Unlock()

	c.master.UpdateNewClient(client)
}

// Unregister unregisters a client
func (c *ClientManager) Unregister(client RouteTableClient) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.clients[client]; !ok {
		return
	}
	delete(c.clients, client)
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
