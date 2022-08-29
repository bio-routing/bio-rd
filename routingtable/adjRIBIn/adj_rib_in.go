package adjRIBIn

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/util/log"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	clientManager     *routingtable.ClientManager
	rt                *routingtable.RoutingTable
	mu                sync.RWMutex
	exportFilterChain filter.Chain
	contributingASNs  *routingtable.ContributingASNs
	sessionAttrs      routingtable.SessionAttrs
}

// New creates a new Adjacency RIB In
func New(exportFilterChain filter.Chain, contributingASNs *routingtable.ContributingASNs, sessionAttrs routingtable.SessionAttrs) *AdjRIBIn {
	a := &AdjRIBIn{
		rt:                routingtable.NewRoutingTable(),
		exportFilterChain: exportFilterChain,
		contributingASNs:  contributingASNs,
		sessionAttrs:      sessionAttrs,
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
	var oldPaths []*route.Path
	if a.sessionAttrs.AddPathRX {
		oldPaths = make([]*route.Path, 0)
		r := a.rt.Get(pfx)
		if r != nil {
			for _, path := range r.Paths() {
				// RFC7911 sec 5 par 1 states (pfx, PathIdentifier) should be unique
				if path.BGPPath.PathIdentifier == p.BGPPath.PathIdentifier {
					a.rt.RemovePath(pfx, path)
					oldPaths = append(oldPaths, path)
				}
			}
		}
		a.rt.AddPath(pfx, p)
	} else {
		oldPaths = a.rt.ReplacePath(pfx, p)
	}
	a.removePathsFromClients(pfx, oldPaths)

	// Bail out if this path is considered ineligible
	p.HiddenReason = a.validatePath(p)
	if p.HiddenReason != route.HiddenReasonNone {
		return nil
	}

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
		if a.sessionAttrs.AddPathRX {
			if p != nil && path.BGPPath.PathIdentifier != p.BGPPath.PathIdentifier {
				continue
			}
		}

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

// Validate path information
func (a *AdjRIBIn) validatePath(p *route.Path) uint8 {
	// Bail out - for all clients for now - if any of our ASNs is within the path
	if a.ourASNsInPath(p) {
		return route.HiddenReasonASLoop
	}

	// RFC4456 Sect. 8: Ignore route with our RouterID as OriginatorID
	if p.BGPPath.BGPPathA.OriginatorID == a.sessionAttrs.RouterID {
		return route.HiddenReasonOurOriginatorID
	}

	// RFC4456 Sect. 8: Ignore routes which contain our ClusterID in their ClusterList
	if p.BGPPath.ClusterList != nil && len(*p.BGPPath.ClusterList) > 0 {
		for _, cid := range *p.BGPPath.ClusterList {
			if cid == a.sessionAttrs.ClusterID {
				return route.HiddenReasonClusterLoop
			}
		}
	}

	// RFC9234 Sect. 5: Validate OTC attribute
	if !a.validatePathOnlyToCustomer(p) {
		return route.HiddenReasonOTCMismatch
	}

	return route.HiddenReasonNone
}

func (a *AdjRIBIn) ourASNsInPath(p *route.Path) bool {
	if p.BGPPath.ASPath == nil {
		return false
	}

	for _, pathSegment := range *p.BGPPath.ASPath {
		for _, asn := range pathSegment.ASNs {
			if a.contributingASNs.IsContributingASN(asn) {
				return true
			}
		}
	}

	return false
}

// RFC9234 Sect 5. - BGP Peer Roles & BGO OTC path attribute
func (a *AdjRIBIn) validatePathOnlyToCustomer(path *route.Path) bool {
	// If either we or the peer don't have a peer role configured, we can't check OTC and consider the path valid.
	if !a.sessionAttrs.PeerRoleEnabled || !a.sessionAttrs.PeerRoleAdvByPeer {
		return true
	}

	pr := a.sessionAttrs.PeerRoleRemote

	// RFC9234 Sect. 5 - Validate OTC attribute if present
	if path.BGPPath.BGPPathA.OnlyToCustomer != 0 {
		// 1. Route with OTC attr from Customer or RS-Client are ineligible
		if pr == packet.PeerRoleRoleCustomer || pr == packet.PeerRoleRoleRSClient {
			return false
		}

		// 2. Route with OTC attr from peer but peer AS != OTC AS
		if pr == packet.PeerRoleRolePeer && path.BGPPath.BGPPathA.OnlyToCustomer != a.sessionAttrs.PeerASN {
			return false
		}
	}

	// 3. Route rcvd from Provider, Peer or RS without OTC set
	if path.BGPPath.BGPPathA.OnlyToCustomer == 0 &&
		(pr == packet.PeerRoleRoleProvider || pr == packet.PeerRoleRolePeer || pr == packet.PeerRoleRoleRS) {
		path.BGPPath.BGPPathA.OnlyToCustomer = a.sessionAttrs.PeerASN
		return true
	}

	return true
}
