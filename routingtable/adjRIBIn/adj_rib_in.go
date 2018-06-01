package adjRIBIn

import (
	"sync"

	"fmt"

	"github.com/bio-routing/bio-rd/metrics"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	log "github.com/sirupsen/logrus"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	routingtable.ClientManager
	rt       *routingtable.RoutingTable
	neighbor *routingtable.Neighbor
	mu       sync.RWMutex
}

// New creates a new Adjacency RIB In
func New(neighbor *routingtable.Neighbor) *AdjRIBIn {
	a := &AdjRIBIn{
		rt:       routingtable.NewRoutingTable(),
		neighbor: neighbor,
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

			err := client.AddPath(route.Prefix(), path)
			if err != nil {
				log.WithField("Sender", "AdjRIBOutAddPath").WithError(err).Error("Could not send update to client")
			}
		}
	}
	return nil
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBIn) AddPath(pfx net.Prefix, p *route.Path) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	defer metrics.PathUpdates.WithLabelValues(fmt.Sprintf("adjRIBIn-%s", a.neighbor.String()), metrics.AddPathAction).Inc()

	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	for _, client := range a.ClientManager.Clients() {
		client.AddPath(pfx, p)
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBIn) RemovePath(pfx net.Prefix, p *route.Path) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	defer metrics.PathUpdates.WithLabelValues(fmt.Sprintf("adjRIBIn-%s", a.neighbor.String()), metrics.RemovePathAction).Inc()

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
		for _, client := range a.ClientManager.Clients() {
			client.RemovePath(pfx, path)
		}
	}
}

func (a *AdjRIBIn) RT() *routingtable.RoutingTable {
	return a.rt
}
