package adjRIBOutAddPath

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/taktv6/tflow2/convert"
)

// AdjRIBOutAddPath represents an Adjacency RIB Out with BGP add path
type AdjRIBOutAddPath struct {
	routingtable.ClientManager
	rt       *routingtable.RoutingTable
	neighbor *routingtable.Neighbor
	mu       sync.RWMutex
}

// New creates a new Adjacency RIB Out with BGP add path
func New(neighbor *routingtable.Neighbor) *AdjRIBOutAddPath {
	a := &AdjRIBOutAddPath{
		rt:       routingtable.NewRoutingTable(),
		neighbor: neighbor,
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBOutAddPath) UpdateNewClient(client routingtable.RouteTableClient) error {
	return nil
}

// AddPath adds path p to prefix `pfx`
func (a *AdjRIBOutAddPath) AddPath(pfx net.Prefix, p *route.Path) error {
	fmt.Printf("THIS IS ADD PATH ON ADJ RIB OUT FOR %v", convert.Uint32Byte(a.neighbor.Address))

	if a.isOwnPath(p) {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	p.BGPPath.PathIdentifier = 7

	a.rt.AddPath(pfx, p)

	for _, client := range a.ClientManager.Clients() {
		client.AddPath(pfx, p)
	}
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBOutAddPath) RemovePath(pfx net.Prefix, p *route.Path) bool {
	if a.isOwnPath(p) {
		return false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	r := a.rt.Get(pfx)
	if r == nil {
		return false
	}

	a.rt.RemovePath(pfx, p)
	a.removePathFromClients(pfx, p)
	return true
}

func (a *AdjRIBOutAddPath) isOwnPath(p *route.Path) bool {
	if p.Type != a.neighbor.Type {
		return false
	}

	switch p.Type {
	case route.BGPPathType:
		return p.BGPPath.Source == a.neighbor.Address
	}

	return false
}

func (a *AdjRIBOutAddPath) removePathFromClients(pfx net.Prefix, path *route.Path) {
	for _, client := range a.ClientManager.Clients() {
		client.RemovePath(pfx, path)
	}
}

// Print dumps all prefixes in the Adj-RIB
func (a *AdjRIBOutAddPath) Print() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ret := fmt.Sprintf("DUMPING ADJ-RIB-OUT:\n")
	routes := a.rt.Dump()
	for _, r := range routes {
		ret += fmt.Sprintf("%s\n", r.Prefix().String())
	}

	return ret
}
