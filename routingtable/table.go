package routingtable

import (
	"sync"
	"sync/atomic"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// RoutingTable is a binary trie that stores prefixes and their paths
type RoutingTable struct {
	routeCount int64
	root       *node
	mu         sync.RWMutex
}

// NewRoutingTable creates a new routing table
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{}
}

// GetRouteCount gets the amount of stored routes
func (rt *RoutingTable) GetRouteCount() int64 {
	return atomic.LoadInt64(&rt.routeCount)
}

// AddPath adds a path to the routing table
func (rt *RoutingTable) AddPath(pfx *net.Prefix, p *route.Path) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	return rt.addPath(pfx, p)
}

func (rt *RoutingTable) addPath(pfx *net.Prefix, p *route.Path) error {
	if rt.root == nil {
		rt.root = newNode(pfx, p, pfx.Len(), false)
		atomic.AddInt64(&rt.routeCount, 1)
		return nil
	}

	root, isNew := rt.root.addPath(pfx, p)
	rt.root = root
	if isNew {
		atomic.AddInt64(&rt.routeCount, 1)
	}
	return nil
}

// ReplacePath replaces all paths for prefix `pfx` with path `p`
func (rt *RoutingTable) ReplacePath(pfx *net.Prefix, p *route.Path) []*route.Path {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	r := rt.get(pfx)
	if r == nil {
		rt.addPath(pfx, p)
		return nil
	}

	oldPaths := r.Paths()
	rt.removePaths(pfx, oldPaths)

	rt.addPath(pfx, p)
	return oldPaths
}

// RemovePath removes a path from the trie
func (rt *RoutingTable) RemovePath(pfx *net.Prefix, p *route.Path) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.removePath(pfx, p)
}

func (rt *RoutingTable) removePath(pfx *net.Prefix, p *route.Path) {
	if rt.root.removePath(pfx, p) {
		atomic.AddInt64(&rt.routeCount, -1)
	}
}

func (rt *RoutingTable) removePaths(pfx *net.Prefix, paths []*route.Path) {
	for _, p := range paths {
		rt.removePath(pfx, p)
	}
}

// RemovePfx removes all paths for prefix `pfx`
func (rt *RoutingTable) RemovePfx(pfx *net.Prefix) []*route.Path {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	r := rt.get(pfx)
	if r == nil {
		return nil
	}

	oldPaths := r.Paths()
	rt.removePaths(pfx, oldPaths)
	return oldPaths
}

// LPM performs a longest prefix match for pfx on lpm
func (rt *RoutingTable) LPM(pfx *net.Prefix) (res []*route.Route) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if rt.root == nil {
		return nil
	}

	rt.root.lpm(pfx, &res)
	return res
}

// Get gets the route for pfx from the LPM
func (rt *RoutingTable) Get(pfx *net.Prefix) *route.Route {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	return rt.get(pfx)
}

func (rt *RoutingTable) get(pfx *net.Prefix) *route.Route {
	if rt.root == nil {
		return nil
	}

	res := rt.root.get(pfx)
	if res == nil {
		return nil
	}
	return res.route
}

// GetLonger gets prefix pfx and all it's more specifics from the LPM
func (rt *RoutingTable) GetLonger(pfx *net.Prefix) (res []*route.Route) {
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
