package adjRIBOut

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	log "github.com/sirupsen/logrus"
)

// AdjRIBOut represents an Adjacency RIB In as described in RFC4271
type AdjRIBOut struct {
	routingtable.ClientManager
	rt       *routingtable.RoutingTable
	neighbor *routingtable.Neighbor
	mu       sync.RWMutex
}

// New creates a new Adjacency RIB In
func New(neighbor *routingtable.Neighbor) *AdjRIBOut {
	a := &AdjRIBOut{
		rt:       routingtable.NewRoutingTable(),
		neighbor: neighbor,
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBOut) UpdateNewClient(client routingtable.RouteTableClient) error {
	return nil
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBOut) AddPath(pfx net.Prefix, p *route.Path) error {
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	for _, client := range a.ClientManager.Clients() {
		err := client.AddPath(pfx, p)
		if err != nil {
			log.WithField("Sender", "AdjRIBOut").WithError(err).Error("Could not send update to client")
		}
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBOut) RemovePath(pfx net.Prefix, p *route.Path) bool {
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		return false
	}

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

func (a *AdjRIBOut) removePathsFromClients(pfx net.Prefix, paths []*route.Path) {
	for _, path := range paths {
		for _, client := range a.ClientManager.Clients() {
			client.RemovePath(pfx, path)
		}
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
