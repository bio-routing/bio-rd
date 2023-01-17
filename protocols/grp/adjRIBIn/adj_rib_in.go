package adjRIBIn

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/log"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	clientManager     *routingtable.ClientManager
	rt                *routingtable.RoutingTable
	mu                sync.RWMutex
	exportFilterChain filter.Chain
	vrf               *vrf.VRF
}

// New creates a new Adjacency RIB In
func New(exportFilterChain filter.Chain, vrf *vrf.VRF) *AdjRIBIn {
	a := &AdjRIBIn{
		rt:                routingtable.NewRoutingTable(),
		exportFilterChain: exportFilterChain,
		vrf:               vrf,
	}
	a.clientManager = routingtable.NewClientManager(a)
	return a
}

// ClientCount gets the number of registered clients
func (a *AdjRIBIn) ClientCount() uint64 {
	return a.clientManager.ClientCount()
}

// Dump dumps the RIB
func (a *AdjRIBIn) Dump() []*route.Route {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.rt.Dump()
}

// Flush drops all routes from the AdjRIBIn
func (a *AdjRIBIn) Flush() {
	a.mu.Lock()
	defer a.mu.Unlock()

	routes := a.rt.Dump()
	for _, route := range routes {
		for _, path := range route.Paths() {
			a.removePath(route.Prefix(), path)
		}
	}
}

// ReplaceFilterChain replaces the filter chain
func (a *AdjRIBIn) ReplaceFilterChain(c filter.Chain) {
	a.mu.Lock()
	defer a.mu.Unlock()

	routes := a.rt.Dump()
	for _, route := range routes {
		paths := route.Paths()
		for _, path := range paths {
			currentPath, currentReject := a.exportFilterChain.Process(route.Prefix(), path)
			newPath, newReject := c.Process(route.Prefix(), path)

			if currentReject && newReject {
				continue
			}

			if currentReject && !newReject {
				for _, client := range a.clientManager.Clients() {
					client.AddPath(route.Prefix(), newPath)
				}

				continue
			}

			if !currentReject && newReject {
				for _, client := range a.clientManager.Clients() {
					client.RemovePath(route.Prefix(), newPath)
				}
				continue
			}

			if !currentReject && !newReject {
				for _, client := range a.clientManager.Clients() {
					if !currentPath.Equal(newPath) {
						client.ReplacePath(route.Prefix(), currentPath, newPath)
					}
				}
			}
		}
	}

	a.exportFilterChain = c
}

func (a *AdjRIBIn) ReplacePath(pfx *net.Prefix, old *route.Path, new *route.Path) {

}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBIn) UpdateNewClient(client routingtable.RouteTableClient) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	routes := a.rt.Dump()
	for _, route := range routes {
		paths := route.Paths()
		for _, path := range paths {
			path, reject := a.exportFilterChain.Process(route.Prefix(), path)
			if reject {
				continue
			}

			err := client.AddPathInitialDump(route.Prefix(), path)
			if err != nil {
				log.WithFields(log.Fields{
					"sender": "AdjRIBOutAddPath"},
				).WithError(err).Error("Could not send update to client")
			}
		}
	}

	client.EndOfRIB()

	return nil
}

// RouteCount returns the number of stored routes
func (a *AdjRIBIn) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBIn) AddPath(pfx *net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.addPath(pfx, p)
}

// addPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBIn) addPath(pfx *net.Prefix, p *route.Path) error {
	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	p, reject := a.exportFilterChain.Process(pfx, p)
	if reject {
		p.HiddenReason = route.HiddenReasonFilteredByPolicy
		return nil
	}

	for _, client := range a.clientManager.Clients() {
		client.AddPath(pfx, p)
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBIn) RemovePath(pfx *net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.removePath(pfx, p)
}

// removePath removes the path for prefix `pfx`
func (a *AdjRIBIn) removePath(pfx *net.Prefix, p *route.Path) bool {
	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	removed := make([]*route.Path, 0)
	oldPaths := r.Paths()
	for _, path := range oldPaths {
		a.rt.RemovePath(pfx, path)
		removed = append(removed, path)
	}

	a.removePathsFromClients(pfx, removed)
	return true
}

func (a *AdjRIBIn) removePathsFromClients(pfx *net.Prefix, paths []*route.Path) {
	for _, path := range paths {
		// If this path wasn't eligible in the first place, we didn't announce it
		if path.HiddenReason != route.HiddenReasonNone {
			continue
		}

		path, reject := a.exportFilterChain.Process(pfx, path)
		if reject {
			continue
		}
		for _, client := range a.clientManager.Clients() {
			client.RemovePath(pfx, path)
		}
	}
}

func (a *AdjRIBIn) RT() *routingtable.RoutingTable {
	return a.rt
}

// Register registers a client for updates
func (a *AdjRIBIn) Register(client routingtable.RouteTableClient) {
	a.clientManager.RegisterWithOptions(client, routingtable.ClientOptions{BestOnly: true})
}

// RegisterWithOptions registers a client for updates
func (a *AdjRIBIn) RegisterWithOptions(client routingtable.RouteTableClient, options routingtable.ClientOptions) {
	a.clientManager.RegisterWithOptions(client, options)
}

// Unregister unregisters a client
func (a *AdjRIBIn) Unregister(client routingtable.RouteTableClient) {
	if !a.clientManager.Unregister(client) {
		return
	}

	for _, r := range a.rt.Dump() {
		for _, p := range r.Paths() {
			client.RemovePath(r.Prefix(), p)
		}
	}
}

// RefreshRoute is here to fultill an interface
func (a *AdjRIBIn) RefreshRoute(*net.Prefix, []*route.Path) {

}

// LPM performs a longest prefix match on the routing table
func (a *AdjRIBIn) LPM(pfx *net.Prefix) (res []*route.Route) {
	return a.rt.LPM(pfx)
}

// Get gets a route
func (a *AdjRIBIn) Get(pfx *net.Prefix) *route.Route {
	return a.rt.Get(pfx)
}

// GetLonger gets all more specifics
func (a *AdjRIBIn) GetLonger(pfx *net.Prefix) (res []*route.Route) {
	return a.rt.GetLonger(pfx)
}
