package routingtable

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// RoutingTable is a binary trie that stores prefixes and their paths
type RoutingTable struct {
	root *node
	mu   sync.RWMutex
}

// NewRoutingTable creates a new routing table
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{}
}

// AddPath adds a path to the routing table
func (rt *RoutingTable) AddPath(pfx net.Prefix, p *route.Path) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.root == nil {
		rt.root = newNode(pfx, p, pfx.Pfxlen(), false)
		return nil
	}

	rt.root = rt.root.addPath(pfx, p)
	return nil
}

// RemovePath removes a path from the trie
func (rt *RoutingTable) RemovePath(pfx net.Prefix, p *route.Path) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.root.removePath(pfx, p)
	return nil
}

// LPM performs a longest prefix match for pfx on lpm
func (rt *RoutingTable) LPM(pfx net.Prefix) (res []*route.Route) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if rt.root == nil {
		return nil
	}

	rt.root.lpm(pfx, &res)
	return res
}

// Get get's the route for pfx from the LPM
func (rt *RoutingTable) Get(pfx net.Prefix) *route.Route {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if rt.root == nil {
		return nil
	}

	res := rt.root.get(pfx)
	if res == nil {
		return nil
	}
	return res.route
}

// GetLonger get's prefix pfx and all it's more specifics from the LPM
func (rt *RoutingTable) GetLonger(pfx net.Prefix) (res []*route.Route) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if rt.root == nil {
		return []*route.Route{}
	}

	return rt.root.get(pfx).dumpPfxs(res)
}

// Dump dumps all routes in table rt into a slice
func (rt *RoutingTable) Dump() []*route.Route {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	res := make([]*route.Route, 0)
	return rt.root.dump(res)
}
