package locRIB

import (
	"fmt"
	"math"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/sirupsen/logrus"
)

// LocRIB represents a routing information base
type LocRIB struct {
	routingtable.ClientManager
	rt               *routingtable.RoutingTable
	mu               sync.RWMutex
	contributingASNs *routingtable.ContributingASNs
}

// New creates a new routing information base
func New() *LocRIB {
	a := &LocRIB{
		rt:               routingtable.NewRoutingTable(),
		contributingASNs: routingtable.NewContributingASNs(),
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// GetContributingASNs returns a pointer to the list of contributing ASNs
func (a *LocRIB) GetContributingASNs() *routingtable.ContributingASNs {
	return a.contributingASNs
}

//Count routes from the LocRIP
func (a *LocRIB) Count() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return uint64(len(a.rt.Dump()))
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

// RouteCount returns the number of stored routes
func (a *LocRIB) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *LocRIB) AddPath(pfx net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	logrus.WithFields(map[string]interface{}{
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
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *LocRIB) RemovePath(pfx net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	logrus.WithFields(map[string]interface{}{
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
	for _, client := range a.ClientManager.Clients() {
		opts := a.ClientManager.GetOptions(client)
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
	for _, client := range a.ClientManager.Clients() {
		opts := a.ClientManager.GetOptions(client)
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
