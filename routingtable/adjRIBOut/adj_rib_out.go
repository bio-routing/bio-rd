package adjRIBOut

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"

	bnet "github.com/bio-routing/bio-rd/net"
	log "github.com/sirupsen/logrus"
)

// AdjRIBOut represents an Adjacency RIB Out with BGP add path
type AdjRIBOut struct {
	routingtable.ClientManager
	rt            *routingtable.RoutingTable
	neighbor      *routingtable.Neighbor
	addPathTX     bool
	pathIDManager *pathIDManager
	exportFilter  *filter.Filter
	mu            sync.RWMutex
}

// New creates a new Adjacency RIB Out with BGP add path
func New(neighbor *routingtable.Neighbor, exportFilter *filter.Filter, addPathTX bool) *AdjRIBOut {
	a := &AdjRIBOut{
		rt:            routingtable.NewRoutingTable(),
		neighbor:      neighbor,
		pathIDManager: newPathIDManager(),
		exportFilter:  exportFilter,
		addPathTX:     addPathTX,
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBOut) UpdateNewClient(client routingtable.RouteTableClient) error {
	return nil
}

// RouteCount returns the number of stored routes
func (a *AdjRIBOut) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

// AddPath adds path p to prefix `pfx`
func (a *AdjRIBOut) AddPath(pfx bnet.Prefix, p *route.Path) error {
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		if a.addPathTX {
			a.removePathsForPrefix(pfx)
		}
		return nil
	}

	// Don't export routes learned via iBGP to an iBGP neighbor which is NOT a route reflection client
	if !p.BGPPath.EBGP && a.neighbor.IBGP && !a.neighbor.RouteReflectorClient {
		return nil
	}

	// If the neighbor is an eBGP peer and not a Route Server client modify ASPath and Next Hop
	p = p.Copy()
	if !a.neighbor.IBGP && !a.neighbor.RouteServerClient {
		p.BGPPath.Prepend(a.neighbor.LocalASN, 1)
		p.BGPPath.NextHop = a.neighbor.LocalAddress
	}

	// If the iBGP neighbor is a route reflection client...
	if a.neighbor.IBGP && a.neighbor.RouteReflectorClient {
		/*
		 * RFC4456 Section 8:
		 * This attribute will carry the BGP Identifier of the originator of the route in the local AS.
		 * A BGP speaker SHOULD NOT create an ORIGINATOR_ID attribute if one already exists.
		 */
		if p.BGPPath.OriginatorID == 0 {
			p.BGPPath.OriginatorID = p.BGPPath.Source.ToUint32()
		}

		/*
		 * When an RR reflects a route, it MUST prepend the local CLUSTER_ID to the CLUSTER_LIST.
		 * If the CLUSTER_LIST is empty, it MUST create a new one.
		 */
		cList := make([]uint32, len(p.BGPPath.ClusterList)+1)
		copy(cList[1:], p.BGPPath.ClusterList)
		cList[0] = a.neighbor.ClusterID
		p.BGPPath.ClusterList = cList
	}

	p, reject := a.exportFilter.ProcessTerms(pfx, p)
	if reject {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.addPathTX {
		pathID, err := a.pathIDManager.addPath(p)
		if err != nil {
			return fmt.Errorf("Unable to get path ID: %v", err)
		}

		p.BGPPath.PathIdentifier = pathID
		a.rt.AddPath(pfx, p)
	} else {
		// rt.ReplacePath will add this path to the rt in any case, so no rt.AddPath here!
		oldPaths := a.rt.ReplacePath(pfx, p)
		a.removePathsFromClients(pfx, oldPaths)
	}

	for _, client := range a.ClientManager.Clients() {
		err := client.AddPath(pfx, p)
		if err != nil {
			log.WithField("Sender", "AdjRIBOutAddPath").WithError(err).Error("Could not send update to client")
		}
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBOut) RemovePath(pfx bnet.Prefix, p *route.Path) bool {
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		return false
	}

	p, reject := a.exportFilter.ProcessTerms(pfx, p)
	if reject {
		return false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	a.rt.RemovePath(pfx, p)

	// If the neighbar has AddPath capabilities, try to find the PathID
	if a.addPathTX {
		pathID, err := a.pathIDManager.releasePath(p)
		if err != nil {
			log.Warningf("Unable to release path for prefix %s: %v", pfx.String(), err)
			return true
		}

		p = p.Copy()
		p.BGPPath.PathIdentifier = pathID
	}

	a.removePathFromClients(pfx, p)
	return true
}

func (a *AdjRIBOut) removePathsForPrefix(pfx bnet.Prefix) bool {
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
		return p.BGPPath.Source == a.neighbor.Address
	}

	return false
}

func (a *AdjRIBOut) removePathsFromClients(pfx bnet.Prefix, paths []*route.Path) {
	for _, p := range paths {
		a.removePathFromClients(pfx, p)
	}
}

func (a *AdjRIBOut) removePathFromClients(pfx bnet.Prefix, path *route.Path) {
	for _, client := range a.ClientManager.Clients() {
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

// Dump all routes present in this AdjRIBOut
func (a *AdjRIBOut) Dump() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ret := fmt.Sprintf("DUMPING ADJ-RIB-OUT:\n")
	routes := a.rt.Dump()
	for _, r := range routes {
		ret += fmt.Sprintf("%s\n", r.Print())
	}

	return ret
}
