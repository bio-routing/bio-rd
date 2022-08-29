package adjRIBOut

import (
	"fmt"
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/util/log"
)

// AdjRIBOut represents an Adjacency RIB Out with BGP add path
type AdjRIBOut struct {
	clientManager            *routingtable.ClientManager
	rib                      *locRIB.LocRIB
	rt                       *routingtable.RoutingTable
	sessionAttrs             routingtable.SessionAttrs
	pathIDManager            *pathIDManager
	exportFilterChain        filter.Chain
	exportFilterChainPending filter.Chain
	mu                       sync.RWMutex
}

// New creates a new Adjacency RIB Out with BGP add path
func New(rib *locRIB.LocRIB, sessionAttrs routingtable.SessionAttrs, exportFilterChain filter.Chain) *AdjRIBOut {
	a := &AdjRIBOut{
		rib:               rib,
		rt:                routingtable.NewRoutingTable(),
		sessionAttrs:      sessionAttrs,
		pathIDManager:     newPathIDManager(),
		exportFilterChain: exportFilterChain,
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

func (a *AdjRIBOut) checkPropagateUpdate(pfx *bnet.Prefix, p *route.Path) (retPath *route.Path, propagate bool) {
	if !routingtable.ShouldPropagateUpdate(pfx, p, &a.sessionAttrs) {
		if a.sessionAttrs.AddPathTX {
			a.removePathsForPrefix(pfx)
		}
		return nil, false
	}

	p = p.Copy()

	if a.sessionAttrs.IBGP {
		return a.checkPropagateUpdateIBGP(pfx, p)
	}

	return a.checkPropagateUpdateEBGP(pfx, p)
}

func (a *AdjRIBOut) checkPropagateUpdateIBGP(pfx *bnet.Prefix, p *route.Path) (retPath *route.Path, propagate bool) {
	// Don't export routes learned via iBGP to an iBGP neighbor which is NOT a route reflection client
	if !p.BGPPath.BGPPathA.EBGP && a.sessionAttrs.IBGP && !a.sessionAttrs.RouteReflectorClient {
		return nil, false
	}

	// If the iBGP neighbor is a route reflection client...
	if a.sessionAttrs.IBGP && a.sessionAttrs.RouteReflectorClient {
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
		cList[0] = a.sessionAttrs.ClusterID
		p.BGPPath.ClusterList = &cList
	}

	return p, true
}

func (a *AdjRIBOut) checkPropagateUpdateEBGP(pfx *bnet.Prefix, p *route.Path) (retPath *route.Path, propagate bool) {
	// If the neighbor is an eBGP peer and not a Route Server client modify ASPath and Next Hop
	if !a.sessionAttrs.RouteServerClient {
		p.BGPPath.Prepend(a.sessionAttrs.LocalASN, 1)
		p.BGPPath.BGPPathA.NextHop = a.sessionAttrs.LocalIP
	}

	// RFC9234 Sect 5. Egress par. - Check OTC attribute
	if a.sessionAttrs.PeerRoleEnabled && a.sessionAttrs.PeerRoleAdvByPeer {
		pr := a.sessionAttrs.PeerRoleRemote

		// 2. If a route already contains the OTC Attribute, it MUST NOT be propagated to Providers, Peers, or RSes
		if p.BGPPath.BGPPathA.OnlyToCustomer != 0 &&
			(pr == packet.PeerRoleRoleProvider || pr == packet.PeerRoleRolePeer || pr == packet.PeerRoleRoleRS) {
			return nil, false
		}

		// 1. If peer is Customer, Peer, or RSClient and OTC is not present we MUST add it with our ASN
		if p.BGPPath.BGPPathA.OnlyToCustomer == 0 &&
			(pr == packet.PeerRoleRoleCustomer || pr == packet.PeerRoleRolePeer || pr == packet.PeerRoleRoleRSClient) {
			p.BGPPath.BGPPathA.OnlyToCustomer = a.sessionAttrs.PeerASN
		}

	}

	return p, true
}

func (a *AdjRIBOut) AddPathInitialDump(pfx *bnet.Prefix, p *route.Path) error {
	return a.AddPath(pfx, p)
}

// AddPath adds path p to prefix `pfx`
func (a *AdjRIBOut) AddPath(pfx *bnet.Prefix, p *route.Path) error {
	p, propagate := a.checkPropagateUpdate(pfx, p)
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
	if a.sessionAttrs.AddPathTX {
		pathID, err := a.pathIDManager.addPath(p)
		if err != nil {
			return fmt.Errorf("unable to get path ID: %w", err)
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
	if !routingtable.ShouldPropagateUpdate(pfx, p, &a.sessionAttrs) {
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
	if a.sessionAttrs.AddPathTX {
		for _, sp := range r.Paths() {
			if sp.Select(p) == 0 {
				a.rt.RemovePath(pfx, sp)

				_, err := a.pathIDManager.releasePath(p)
				if err != nil {
					log.WithError(err).
						Errorf("Unable to release path for prefix %s: %v", pfx.String(), err)
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
	if p.Type != a.sessionAttrs.Type {
		return false
	}

	switch p.Type {
	case route.BGPPathType:
		return p.BGPPath.BGPPathA.Source == a.sessionAttrs.PeerIP
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
	for _, p := range ribPaths {
		p, propagate := a.checkPropagateUpdate(pfx, p)
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
