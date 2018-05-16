package adjRIBOut

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/taktv6/tflow2/convert"
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
	fmt.Printf("THIS IS ADJ RIB OUT NON ADD PATH FOR %v\n", convert.Uint32Byte(a.neighbor.Address))

	if a.isOwnPath(p) {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)

	for _, client := range a.ClientManager.Clients() {
		client.AddPath(pfx, p)
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBOut) RemovePath(pfx net.Prefix, p *route.Path) bool {
	if a.isOwnPath(p) {
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
