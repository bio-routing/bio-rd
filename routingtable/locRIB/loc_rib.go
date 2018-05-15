package locRIB

import (
	"fmt"
	"math"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// LocRIB represents a routing information base
type LocRIB struct {
	routingtable.ClientManager
	rt *routingtable.RoutingTable
	mu sync.RWMutex
}

// New creates a new routing information base
func New() *LocRIB {
	a := &LocRIB{
		rt: routingtable.NewRoutingTable(),
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// UpdateNewClient sends current state to a new client
func (a *LocRIB) UpdateNewClient(client routingtable.RouteTableClient) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	routes := a.rt.Dump()
	for _, r := range routes {
		a.propagateChanges(&route.Route{}, r)
	}

	return nil
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *LocRIB) AddPath(pfx net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()

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

	fmt.Printf("NEW: %v\n", newRoute.Paths())
	fmt.Printf("OLD: %v\n", oldRoute.Paths())

	a.propagateChanges(oldRoute, newRoute)
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *LocRIB) RemovePath(pfx net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	var oldRoute *route.Route
	r := a.rt.Get(pfx)
	if r != nil {
		oldRoute = r.Copy()
	}

	a.rt.RemovePath(pfx, p)
	r.PathSelection()

	r = a.rt.Get(pfx)
	newRoute := r.Copy()

	fmt.Printf("NEW: %v\n", newRoute.Paths())
	fmt.Printf("OLD: %v\n", oldRoute.Paths())

	a.propagateChanges(oldRoute, newRoute)
	return true
}

func (a *LocRIB) propagateChanges(oldRoute *route.Route, newRoute *route.Route) {
	a.removePathsFromClients(oldRoute, newRoute)
	a.addPathsToClients(oldRoute, newRoute)
}

func (a *LocRIB) addPathsToClients(oldRoute *route.Route, newRoute *route.Route) {
	for _, client := range a.ClientManager.Clients() {
		opts := a.ClientManager.GetOptions(client)
		oldMaxPaths := opts.GetMaxPaths(oldRoute.ECMPPathCount())
		newMaxPaths := opts.GetMaxPaths(newRoute.ECMPPathCount())

		oldPathsLimit := int(math.Min(float64(oldMaxPaths), float64(len(oldRoute.Paths()))))
		newPathsLimit := int(math.Min(float64(newMaxPaths), float64(len(newRoute.Paths()))))

		advertise := route.PathsDiff(newRoute.Paths()[0:newPathsLimit], oldRoute.Paths()[0:oldPathsLimit])
		fmt.Printf("ADVERTISING PATHS %v TO CLIENTS\n", advertise)

		for _, p := range advertise {
			client.AddPath(newRoute.Prefix(), p)
		}
	}
}

func (a *LocRIB) removePathsFromClients(oldRoute *route.Route, newRoute *route.Route) {
	for _, client := range a.ClientManager.Clients() {
		opts := a.ClientManager.GetOptions(client)
		oldMaxPaths := opts.GetMaxPaths(oldRoute.ECMPPathCount())
		newMaxPaths := opts.GetMaxPaths(newRoute.ECMPPathCount())

		oldPathsLimit := int(math.Min(float64(oldMaxPaths), float64(len(oldRoute.Paths()))))
		newPathsLimit := int(math.Min(float64(newMaxPaths), float64(len(newRoute.Paths()))))

		withdraw := route.PathsDiff(oldRoute.Paths()[0:oldPathsLimit], newRoute.Paths()[0:newPathsLimit])

		for _, p := range withdraw {
			client.RemovePath(newRoute.Prefix(), p)
		}
	}
}

// ContainsPfxPath returns true if this prefix and path combination is
// present in this LocRIB.
func (a *LocRIB) ContainsPfxPath(pfx net.Prefix, p *route.Path) bool {
	a.mu.RLock()
	contains := false
	r := a.rt.Get(pfx)
	if r != nil {
		for _, path := range r.Paths() {
			if path.Compare(p) == 0 {
				contains = true
			}
		}
	} else {
		contains = false
	}
	a.mu.RUnlock()
	return contains
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
