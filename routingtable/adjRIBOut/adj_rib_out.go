package adjRIBOut

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// AdjRIBOut represents an Adjacency RIB Out with BGP add path
type AdjRIBOut struct {
	clientManager            *routingtable.ClientManager
	rib                      *locRIB.LocRIB
	rt                       *routingtable.RoutingTable
	neighbor                 *routingtable.Neighbor
	addPathTX                bool
	pathIDManager            *pathIDManager
	exportFilterChain        filter.Chain
	exportFilterChainPending filter.Chain
	mu                       sync.RWMutex
}

// New creates a new Adjacency RIB Out with BGP add path
func New(rib *locRIB.LocRIB, neighbor *routingtable.Neighbor, exportFilterChain filter.Chain, addPathTX bool) *AdjRIBOut {
	a := &AdjRIBOut{
		rib:               rib,
		rt:                routingtable.NewRoutingTable(),
		neighbor:          neighbor,
		pathIDManager:     newPathIDManager(),
		exportFilterChain: exportFilterChain,
		addPathTX:         addPathTX,
	}
	a.clientManager = routingtable.NewClientManager(a)
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

// UpdateNewClient sends current state to a new client
func (a *AdjRIBOut) UpdateNewClient(client routingtable.RouteTableClient) error {
	return nil
}

// RouteCount returns the number of stored routes
func (a *AdjRIBOut) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

func (a *AdjRIBOut) bgpChecks(pfx *bnet.Prefix, p *route.Path) (retPath *route.Path, propagate bool) {
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		if a.addPathTX {
			a.removePathsForPrefix(pfx)
		}
		return nil, false
	}

	// Don't export routes learned via iBGP to an iBGP neighbor which is NOT a route reflection client
	if !p.BGPPath.BGPPathA.EBGP && a.neighbor.IBGP && !a.neighbor.RouteReflectorClient {
		return nil, false
	}

	// If the neighbor is an eBGP peer and not a Route Server client modify ASPath and Next Hop
	p = p.Copy()
	if !a.neighbor.IBGP && !a.neighbor.RouteServerClient {
		p.BGPPath.Prepend(a.neighbor.LocalASN, 1)
		p.BGPPath.BGPPathA.NextHop = a.neighbor.LocalAddress
	}

	// If the iBGP neighbor is a route reflection client...
	if a.neighbor.IBGP && a.neighbor.RouteReflectorClient {
		/*
		 * RFC4456 Section 8:
		 * This attribute will carry the BGP Identifier of the originator of the route in the local AS.
		 * A BGP speaker SHOULD NOT create an ORIGINATOR_ID attribute if one already exists.
		 */
		if p.BGPPath.BGPPathA.OriginatorID == 0 {
			p.BGPPath.BGPPathA.OriginatorID = p.BGPPath.BGPPathA.Source.ToUint32()
		}

		/*
		 * When an RR reflects a route, it MUST prepend the local CLUSTER_ID to the CLUSTER_LIST.
		 * If the CLUSTER_LIST is empty, it MUST create a new one.
		 */

		x := 1
		if p.BGPPath.ClusterList != nil {
			x += len(*p.BGPPath.ClusterList)
		}
		cList := make(types.ClusterList, x)
		if p.BGPPath.ClusterList != nil {
			copy(cList[1:], *p.BGPPath.ClusterList)
		}
		cList[0] = a.neighbor.ClusterID
		p.BGPPath.ClusterList = &cList
	}

	return p, true
}

func (a *AdjRIBOut) AddPathInitialDump(pfx *bnet.Prefix, p *route.Path) error {
	return a.AddPath(pfx, p)
}

// AddPath adds path p to prefix `pfx`
func (a *AdjRIBOut) AddPath(pfx *bnet.Prefix, p *route.Path) error {
	p, propagate := a.bgpChecks(pfx, p)
	if !propagate {
		return nil
	}

	p, reject := a.exportFilterChain.Process(pfx, p)
	if reject {
		return nil
	}

	p.BGPPath = p.BGPPath.Dedup()

	a.mu.Lock()
	defer a.mu.Unlock()

	return a.addPath(pfx, p)
}

func (a *AdjRIBOut) addPath(pfx *bnet.Prefix, p *route.Path) error {
	if a.addPathTX {
		pathID, err := a.pathIDManager.addPath(p)
		if err != nil {
			return errors.Wrap(err, "Unable to get path ID")
		}

		p.BGPPath.PathIdentifier = pathID
		a.rt.AddPath(pfx, p)
	} else {
		// rt.ReplacePath will add this path to the rt in any case, so no rt.AddPath here!
		oldPaths := a.rt.ReplacePath(pfx, p)
		a.removePathsFromClients(pfx, oldPaths)
	}

	for _, client := range a.clientManager.Clients() {
		err := client.AddPath(pfx, p)
		if err != nil {
			log.WithField("Sender", "AdjRIBOutAddPath").WithError(err).Error("Could not send update to client")
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
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		return false
	}

	p, reject := a.exportFilterChain.Process(pfx, p)
	if reject {
		return false
	}

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	sentPath := p
	if a.addPathTX {
		for _, sp := range r.Paths() {
			if sp.Select(p) == 0 {
				a.rt.RemovePath(pfx, sp)

				_, err := a.pathIDManager.releasePath(p)
				if err != nil {
					log.Warningf("Unable to release path for prefix %s: %v", pfx.String(), err)
					return true
				}

				sentPath = sp
				break
			}

		}

		if sentPath == p {
			return false
		}
	} else {
		a.rt.RemovePath(pfx, p)
	}

	a.removePathFromClients(pfx, sentPath)
	return true
}

func (a *AdjRIBOut) removePathsForPrefix(pfx *bnet.Prefix) bool {
	// We were called before a.AddPath() had a lock, so we need to lock here and release it
	// after the get to prevent a dead lock as RemovePath() will acquire a lock itself!
	a.mu.Lock()
	r := a.rt.Get(pfx)
	a.mu.Unlock()

	// If no path with this prefix is present, we're done
	if r == nil {
		return false
	}

	for _, path := range r.Paths() {
		a.RemovePath(pfx, path)
	}

	return true
}

func (a *AdjRIBOut) isOwnPath(p *route.Path) bool {
	if p.Type != a.neighbor.Type {
		return false
	}

	switch p.Type {
	case route.BGPPathType:
		return p.BGPPath.BGPPathA.Source == a.neighbor.Address
	}

	return false
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

	ret := fmt.Sprintf("DUMPING ADJ-RIB-OUT:\n")
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
func (a *AdjRIBOut) ReplacePath(pfx *net.Prefix, old *route.Path, new *route.Path) {

}

// RefreshRoute refreshes a route
func (a *AdjRIBOut) RefreshRoute(pfx *net.Prefix, ribPaths []*route.Path) {
	for _, p := range ribPaths {
		p, propagate := a.bgpChecks(pfx, p)
		if !propagate {
			continue
		}

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
func (a *AdjRIBOut) LPM(pfx *net.Prefix) (res []*route.Route) {
	return a.rt.LPM(pfx)
}

// Get gets a route
func (a *AdjRIBOut) Get(pfx *net.Prefix) *route.Route {
	return a.rt.Get(pfx)
}

// GetLonger gets all more specifics
func (a *AdjRIBOut) GetLonger(pfx *net.Prefix) (res []*route.Route) {
	return a.rt.GetLonger(pfx)
}

// Destroy is here to fulfill an interface
func (a *AdjRIBOut) Destroy() {

}
