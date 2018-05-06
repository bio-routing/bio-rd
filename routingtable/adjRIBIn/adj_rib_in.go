package adjRIBIn

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	rt *routingtable.RoutingTable
	routingtable.ClientManager
	mu sync.RWMutex
}

// NewAdjRIBIn creates a new Adjacency RIB In
func NewAdjRIBIn() *AdjRIBIn {
	a := &AdjRIBIn{
		rt: routingtable.NewRoutingTable(),
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBIn) UpdateNewClient(client routingtable.RouteTableClient) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	routes := a.rt.Dump()
	for _, route := range routes {
		paths := route.Paths()
		for _, path := range paths {
			client.AddPath(route.Prefix(), path)
		}
	}
	return nil
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBIn) AddPath(pfx net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	for _, client := range a.ClientManager.Clients() {
		client.AddPath(pfx, p)
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBIn) RemovePath(pfx net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	oldPaths := r.Paths()
	for _, path := range oldPaths {
		a.rt.RemovePath(pfx, path)
	}

	a.removePathsFromClients(pfx, oldPaths)
	return true
}

func (a *AdjRIBIn) removePathsFromClients(pfx net.Prefix, paths []*route.Path) {
	for _, path := range paths {
		for _, client := range a.ClientManager.Clients() {
			client.RemovePath(pfx, path)
		}
	}
}
