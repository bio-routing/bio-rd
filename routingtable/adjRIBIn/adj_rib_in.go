package adjRIBIn

import (
	"sync"

	"github.com/bio-routing/bio-rd/routingtable/filter"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	log "github.com/sirupsen/logrus"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	routingtable.ClientManager
	rt               *routingtable.RoutingTable
	mu               sync.RWMutex
	exportFilter     *filter.Filter
	contributingASNs *routingtable.ContributingASNs
	routerID         uint32
	clusterID        uint32
}

// New creates a new Adjacency RIB In
func New(exportFilter *filter.Filter, contributingASNs *routingtable.ContributingASNs, routerID uint32, clusterID uint32) *AdjRIBIn {
	a := &AdjRIBIn{
		rt:               routingtable.NewRoutingTable(),
		exportFilter:     exportFilter,
		contributingASNs: contributingASNs,
		routerID:         routerID,
		clusterID:        clusterID,
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
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

	// RFC4456 Sect. 8: Ignore routes which contian our ClusterID in their ClusterList
	if len(p.BGPPath.ClusterList) > 0 {
		for _, cid := range p.BGPPath.ClusterList {
			if cid == a.clusterID {
				return nil
			}
		}
	}

	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	p, reject := a.exportFilter.ProcessTerms(pfx, p)
	if reject {
		return nil
	}

	// Bail out - for all clients for now - if any of our ASNs is within the path
	if a.ourASNsInPath(p) {
		return nil
	}

	for _, client := range a.ClientManager.Clients() {
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

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	oldPaths := r.Paths()
	for _, path := range oldPaths {
		a.rt.RemovePath(pfx, path)
	}

	a.removePathsFromClients(pfx, oldPaths)
	return true
}

func (a *AdjRIBIn) removePathsFromClients(pfx net.Prefix, paths []*route.Path) {
	for _, path := range paths {
		path, reject := a.exportFilter.ProcessTerms(pfx, path)
		if reject {
			continue
		}
		for _, client := range a.ClientManager.Clients() {
			client.RemovePath(pfx, path)
		}
	}
}

func (a *AdjRIBIn) RT() *routingtable.RoutingTable {
	return a.rt
}
