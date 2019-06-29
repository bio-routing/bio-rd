package locRIB

import (
	"fmt"
	"math"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	log "github.com/sirupsen/logrus"
)

// LocRIB represents a routing information base
type LocRIB struct {
	name             string
	clientManager    *routingtable.ClientManager
	rt               *routingtable.RoutingTable
	mu               sync.RWMutex
	contributingASNs *routingtable.ContributingASNs
	countTarget      *countTarget
}

type countTarget struct {
	target uint64
	ch     chan struct{}
}

// New creates a new routing information base
func New(name string) *LocRIB {
	a := &LocRIB{
		name:             name,
		rt:               routingtable.NewRoutingTable(),
		contributingASNs: routingtable.NewContributingASNs(),
	}
	a.clientManager = routingtable.NewClientManager(a)

	return a
}

// Name gets the name of the LocRIB
func (a *LocRIB) Name() string {
	return a.name
}

// ClientCount gets the number of registered clients
func (a *LocRIB) ClientCount() uint64 {
	return a.clientManager.ClientCount()
}

// GetContributingASNs returns a pointer to the list of contributing ASNs
func (a *LocRIB) GetContributingASNs() *routingtable.ContributingASNs {
	return a.contributingASNs
}

// Count routes from the LocRIB
func (a *LocRIB) Count() uint64 {
	return uint64(a.rt.GetRouteCount())
}

// LPM performs a longest prefix match on the routing table
func (a *LocRIB) LPM(pfx net.Prefix) (res []*route.Route) {
	return a.rt.LPM(pfx)
}

// Get gets a route
func (a *LocRIB) Get(pfx net.Prefix) *route.Route {
	return a.rt.Get(pfx)
}

// GetLonger gets all more specifics
func (a *LocRIB) GetLonger(pfx net.Prefix) (res []*route.Route) {
	return a.rt.GetLonger(pfx)
}

// Dump dumps the RIB
func (a *LocRIB) Dump() []*route.Route {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.rt.Dump()
}

// SetCountTarget sets a target and a channel to send a message to when a certain route count is reached
func (a *LocRIB) SetCountTarget(count uint64, ch chan struct{}) {
	a.countTarget = &countTarget{
		target: count,
		ch:     ch,
	}
}

// UpdateNewClient sends current state to a new client
func (a *LocRIB) UpdateNewClient(client routingtable.RouteTableClient) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	routes := a.rt.Dump()
	for _, r := range routes {
		for _, p := range r.Paths() {
			client.AddPath(r.Prefix(), p)
		}
	}

	return nil
}

// RouteCount returns the number of stored routes
func (a *LocRIB) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *LocRIB) AddPath(pfx net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	log.WithFields(map[string]interface{}{
		"Prefix": pfx,
		"Route":  p,
	}).Debug("AddPath to locRIB")
	routeExisted := false
	oldRoute := &route.Route{}
	r := a.rt.Get(pfx)
	if r != nil {
		oldRoute = r.Copy()
		routeExisted = true
	}

	// FIXME: in AddPath() we assume that the same reference of route (r) is modified (not responsibility of locRIB). If this implementation changes in the future this code will break.
	a.rt.AddPath(pfx, p)
	if !routeExisted {
		r = a.rt.Get(pfx)
	}

	r.PathSelection()
	newRoute := r.Copy()

	a.propagateChanges(oldRoute, newRoute)
	if a.countTarget != nil {
		if a.RouteCount() == int64(a.countTarget.target) {
			a.countTarget.ch <- struct{}{}
		}
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *LocRIB) RemovePath(pfx net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	log.WithFields(map[string]interface{}{
		"Prefix": pfx,
		"Route":  p,
	}).Debug("Remove from locRIB")
	var oldRoute *route.Route
	r := a.rt.Get(pfx)
	if r != nil {
		oldRoute = r.Copy()
	} else {
		return true
	}

	a.rt.RemovePath(pfx, p)
	r.PathSelection()

	r = a.rt.Get(pfx)
	newRoute := r.Copy()

	a.propagateChanges(oldRoute, newRoute)
	return true
}

func (a *LocRIB) propagateChanges(oldRoute *route.Route, newRoute *route.Route) {
	a.removePathsFromClients(oldRoute, newRoute)
	a.addPathsToClients(oldRoute, newRoute)
}

func (a *LocRIB) addPathsToClients(oldRoute *route.Route, newRoute *route.Route) {
	for _, client := range a.clientManager.Clients() {
		opts := a.clientManager.GetOptions(client)
		oldMaxPaths := opts.GetMaxPaths(oldRoute.ECMPPathCount())
		newMaxPaths := opts.GetMaxPaths(newRoute.ECMPPathCount())

		oldPathsLimit := int(math.Min(float64(oldMaxPaths), float64(len(oldRoute.Paths()))))
		newPathsLimit := int(math.Min(float64(newMaxPaths), float64(len(newRoute.Paths()))))

		advertise := route.PathsDiff(newRoute.Paths()[0:newPathsLimit], oldRoute.Paths()[0:oldPathsLimit])

		for _, p := range advertise {
			client.AddPath(newRoute.Prefix(), p)
		}
	}
}

func (a *LocRIB) removePathsFromClients(oldRoute *route.Route, newRoute *route.Route) {
	for _, client := range a.clientManager.Clients() {
		opts := a.clientManager.GetOptions(client)
		oldMaxPaths := opts.GetMaxPaths(oldRoute.ECMPPathCount())
		newMaxPaths := opts.GetMaxPaths(newRoute.ECMPPathCount())

		oldPathsLimit := int(math.Min(float64(oldMaxPaths), float64(len(oldRoute.Paths()))))
		newPathsLimit := int(math.Min(float64(newMaxPaths), float64(len(newRoute.Paths()))))

		withdraw := route.PathsDiff(oldRoute.Paths()[0:oldPathsLimit], newRoute.Paths()[0:newPathsLimit])

		for _, p := range withdraw {
			client.RemovePath(oldRoute.Prefix(), p)
		}
	}
}

// ContainsPfxPath returns true if this prefix and path combination is
// present in this LocRIB.
func (a *LocRIB) ContainsPfxPath(pfx net.Prefix, p *route.Path) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	for _, path := range r.Paths() {
		if path.Equal(p) {
			return true
		}
	}

	return false
}

func (a *LocRIB) String() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ret := ""
	routes := a.rt.Dump()
	for idx, r := range routes {
		if idx < len(routes)-1 {
			ret += fmt.Sprintf("%s, ", r.Prefix().String())
		} else {
			ret += fmt.Sprintf("%s", r.Prefix().String())
		}
	}

	return ret
}

func (a *LocRIB) Print() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ret := "Loc-RIB DUMP:\n"
	routes := a.rt.Dump()
	for _, r := range routes {
		ret += fmt.Sprintf("%s\n", r.Prefix().String())
	}

	return ret
}

// Register registers a client for updates
func (a *LocRIB) Register(client routingtable.RouteTableClient) {
	a.clientManager.RegisterWithOptions(client, routingtable.ClientOptions{BestOnly: true})
}

// RegisterWithOptions registers a client with options for updates
func (a *LocRIB) RegisterWithOptions(client routingtable.RouteTableClient, opt routingtable.ClientOptions) {
	a.clientManager.RegisterWithOptions(client, opt)
}

// Unregister unregisters a client
func (a *LocRIB) Unregister(client routingtable.RouteTableClient) {
	a.clientManager.Unregister(client)
}
