package adjRIBIn

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	log "github.com/sirupsen/logrus"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	clientManager    *routingtable.ClientManager
	rt               *routingtable.RoutingTable
	mu               sync.RWMutex
	exportFilter     *filter.Filter
	contributingASNs *routingtable.ContributingASNs
	routerID         uint32
	clusterID        uint32
	addPathRX        bool
}

// New creates a new Adjacency RIB In
func New(exportFilter *filter.Filter, contributingASNs *routingtable.ContributingASNs, routerID uint32, clusterID uint32, addPathRX bool) *AdjRIBIn {
	a := &AdjRIBIn{
		rt:               routingtable.NewRoutingTable(),
		exportFilter:     exportFilter,
		contributingASNs: contributingASNs,
		routerID:         routerID,
		clusterID:        clusterID,
		addPathRX:        addPathRX,
	}
	a.clientManager = routingtable.NewClientManager(a)
	return a
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

// UpdateNewClient sends current state to a new client
func (a *AdjRIBIn) UpdateNewClient(client routingtable.RouteTableClient) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	routes := a.rt.Dump()
	for _, route := range routes {
		paths := route.Paths()
		for _, path := range paths {
			path, reject := a.exportFilter.ProcessTerms(route.Prefix(), path)
			if reject {
				continue
			}

			err := client.AddPath(route.Prefix(), path)
			if err != nil {
				log.WithField("Sender", "AdjRIBOutAddPath").WithError(err).Error("Could not send update to client")
			}
		}
	}
	return nil
}

// RouteCount returns the number of stored routes
func (a *AdjRIBIn) RouteCount() int64 {
	return a.rt.GetRouteCount()
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBIn) AddPath(pfx net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// RFC4456 Sect. 8: Ignore route with our RouterID as OriginatorID
	if p.BGPPath.OriginatorID == a.routerID {
		return nil
	}

	// RFC4456 Sect. 8: Ignore routes which contain our ClusterID in their ClusterList
	if len(p.BGPPath.ClusterList) > 0 {
		for _, cid := range p.BGPPath.ClusterList {
			if cid == a.clusterID {
				return nil
			}
		}
	}

	if a.addPathRX {
		a.rt.AddPath(pfx, p)
	} else {
		oldPaths := a.rt.ReplacePath(pfx, p)
		a.removePathsFromClients(pfx, oldPaths)
	}

	p, reject := a.exportFilter.ProcessTerms(pfx, p)
	if reject {
		return nil
	}

	// Bail out - for all clients for now - if any of our ASNs is within the path
	if a.ourASNsInPath(p) {
		return nil
	}

	for _, client := range a.clientManager.Clients() {
		client.AddPath(pfx, p)
	}
	return nil
}

func (a *AdjRIBIn) ourASNsInPath(p *route.Path) bool {
	for _, pathSegment := range p.BGPPath.ASPath {
		for _, asn := range pathSegment.ASNs {
			if a.contributingASNs.IsContributingASN(asn) {
				return true
			}
		}
	}

	return false
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBIn) RemovePath(pfx net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.removePath(pfx, p)
}

// removePath removes the path for prefix `pfx`
func (a *AdjRIBIn) removePath(pfx net.Prefix, p *route.Path) bool {
	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	removed := make([]*route.Path, 0)
	oldPaths := r.Paths()
	for _, path := range oldPaths {
		if a.addPathRX {
			if path.BGPPath.PathIdentifier != p.BGPPath.PathIdentifier {
				continue
			}
		}

		a.rt.RemovePath(pfx, path)
		removed = append(removed, path)
	}

	a.removePathsFromClients(pfx, removed)
	return true
}

func (a *AdjRIBIn) removePathsFromClients(pfx net.Prefix, paths []*route.Path) {
	for _, path := range paths {
		path, reject := a.exportFilter.ProcessTerms(pfx, path)
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
