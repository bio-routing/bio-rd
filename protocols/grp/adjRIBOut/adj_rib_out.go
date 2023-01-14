package adjRIBOut

import (
	"fmt"
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/util/log"
)

// AdjRIBOut represents an Adjacency RIB Out
type AdjRIBOut struct {
	clientManager            *routingtable.ClientManager
	rib                      *locRIB.LocRIB
	rt                       *routingtable.RoutingTable
	metaData                 map[string]string
	exportFilterChain        filter.Chain
	exportFilterChainPending filter.Chain
	mu                       sync.RWMutex
}

// New creates a new Adjacency RIB Out
func New(rib *locRIB.LocRIB, exportFilterChain filter.Chain, metaData map[string]string) *AdjRIBOut {
	a := &AdjRIBOut{
		rib:               rib,
		rt:                routingtable.NewRoutingTable(),
		exportFilterChain: exportFilterChain,
	}

	a.clientManager = routingtable.NewClientManager(a)

	if metaData != nil {
		a.metaData = metaData
	} else {
		a.metaData = make(map[string]string)
	}

	return a
}

// ClientCount gets the number of registered clients
func (a *AdjRIBOut) ClientCount() uint64 {
	return a.clientManager.ClientCount()
}

// Dump dumps the RIB
func (a *AdjRIBOut) Dump() []*route.Route {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.rt.Dump()
}

func (a *AdjRIBOut) EndOfRIB() {
	for _, client := range a.clientManager.Clients() {
		client.EndOfRIB()
	}
}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBOut) UpdateNewClient(client routingtable.RouteTableClient) error {
	return nil
}

// RouteCount returns the number of stored routes
func (a *AdjRIBOut) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

func (a *AdjRIBOut) AddPathInitialDump(pfx *bnet.Prefix, p *route.Path) error {
	return a.AddPath(pfx, p)
}

// AddPath adds path p to prefix `pfx`
func (a *AdjRIBOut) AddPath(pfx *bnet.Prefix, p *route.Path) error {
	p, err := p.RedistributeTo(route.GRPPathType)
	if err != nil {
		return err
	}

	p, reject := a.exportFilterChain.Process(pfx, p)
	if reject {
		return nil
	}

	p.GRPPath.AddMetaData(a.metaData)

	a.mu.Lock()
	defer a.mu.Unlock()

	return a.addPath(pfx, p)
}

func (a *AdjRIBOut) addPath(pfx *bnet.Prefix, p *route.Path) error {
	// rt.ReplacePath will add this path to the rt in any case, so no rt.AddPath here!
	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	for _, client := range a.clientManager.Clients() {
		err := client.AddPath(pfx, p)
		if err != nil {
			log.WithFields(log.Fields{
				"sender": "AdjRIBOutAddPath",
			}).WithError(err).Error("Could not send update to client")
		}
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBOut) RemovePath(pfx *bnet.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.removePath(pfx, p)
}

func (a *AdjRIBOut) removePath(pfx *bnet.Prefix, p *route.Path) bool {
	p, reject := a.exportFilterChain.Process(pfx, p)
	if reject {
		return false
	}

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	a.rt.RemovePath(pfx, p)
	a.removePathFromClients(pfx, p)

	return true
}

func (a *AdjRIBOut) removePathsFromClients(pfx *bnet.Prefix, paths []*route.Path) {
	for _, p := range paths {
		a.removePathFromClients(pfx, p)
	}
}

func (a *AdjRIBOut) removePathFromClients(pfx *bnet.Prefix, path *route.Path) {
	for _, client := range a.clientManager.Clients() {
		client.RemovePath(pfx, path)
	}
}

// Print dumps all prefixes in the Adj-RIB
func (a *AdjRIBOut) Print() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ret := "DUMPING ADJ-RIB-OUT:\n"
	routes := a.rt.Dump()
	for _, r := range routes {
		ret += fmt.Sprintf("%s\n", r.Prefix().String())
	}

	return ret
}

// Register registers a client for updates
func (a *AdjRIBOut) Register(client routingtable.RouteTableClient) {
	a.clientManager.RegisterWithOptions(client, routingtable.ClientOptions{BestOnly: true})
}

// RegisterWithOptions registers a client for updates
func (a *AdjRIBOut) RegisterWithOptions(client routingtable.RouteTableClient, options routingtable.ClientOptions) {
	a.clientManager.RegisterWithOptions(client, options)
}

// Unregister unregisters a client
func (a *AdjRIBOut) Unregister(client routingtable.RouteTableClient) {
	a.clientManager.Unregister(client)
}

// ReplaceFilterChain replaces the export filter chain
func (a *AdjRIBOut) ReplaceFilterChain(c filter.Chain) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.exportFilterChainPending = c
	a.rib.RefreshClient(a)
	a.exportFilterChain = c
}

// ReplacePath is here to fulfill an interface
func (a *AdjRIBOut) ReplacePath(pfx *bnet.Prefix, old *route.Path, new *route.Path) {

}

// RefreshRoute refreshes a route
func (a *AdjRIBOut) RefreshRoute(pfx *bnet.Prefix, ribPaths []*route.Path) {
	for _, path := range ribPaths {
		p := path.Copy()

		currentPath, currentReject := a.exportFilterChain.Process(pfx, p)
		newPath, newReject := a.exportFilterChainPending.Process(pfx, p)

		if currentReject && newReject {
			continue
		}

		if !currentReject && newReject {
			a.removePath(pfx, currentPath)
			continue
		}

		if currentReject && !newReject {
			a.addPath(pfx, newPath)
			continue
		}

		if !currentReject && !newReject {
			if !currentPath.Equal(newPath) {
				a.removePath(pfx, currentPath)
				a.addPath(pfx, newPath)
			}
			continue
		}

	}
}

// LPM performs a longest prefix match on the routing table
func (a *AdjRIBOut) LPM(pfx *bnet.Prefix) (res []*route.Route) {
	return a.rt.LPM(pfx)
}

// Get gets a route
func (a *AdjRIBOut) Get(pfx *bnet.Prefix) *route.Route {
	return a.rt.Get(pfx)
}

// GetLonger gets all more specifics
func (a *AdjRIBOut) GetLonger(pfx *bnet.Prefix) (res []*route.Route) {
	return a.rt.GetLonger(pfx)
}

// Dispose is here to fulfill an interface. We don't care if the RIB we're registred to
// as a client is gone as this only happens in the BMP use case where AdjRIBOut is not used.
func (a *AdjRIBOut) Dispose() {}
