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
	pathIDManager *pathIDManager
	mu            sync.RWMutex
	exportFilter  *filter.Filter
}

// New creates a new Adjacency RIB Out with BGP add path
func New(neighbor *routingtable.Neighbor, exportFilter *filter.Filter) *AdjRIBOut {
	a := &AdjRIBOut{
		rt:            routingtable.NewRoutingTable(),
		neighbor:      neighbor,
		pathIDManager: newPathIDManager(),
		exportFilter:  exportFilter,
	}
	a.ClientManager = routingtable.NewClientManager(a)
	return a
}

// UpdateNewClient sends current state to a new client
func (a *AdjRIBOut) UpdateNewClient(client routingtable.RouteTableClient) error {
	return nil
}

// AddPath adds path p to prefix `pfx`
func (a *AdjRIBOut) AddPath(pfx bnet.Prefix, p *route.Path) error {
	if !routingtable.ShouldPropagateUpdate(pfx, p, a.neighbor) {
		return nil
	}

	// Don't export routes learned via iBGP to an iBGP neighbor
	if !p.BGPPath.EBGP && a.neighbor.IBGP {
		return nil
	}

	p = p.Copy()
	if !a.neighbor.IBGP && !a.neighbor.RouteServerClient {
		p.BGPPath.ASPath = fmt.Sprintf("%d %s", a.neighbor.LocalASN, p.BGPPath.ASPath)
	}

	if !a.neighbor.IBGP && !a.neighbor.RouteServerClient {
		p.BGPPath.NextHop = a.neighbor.LocalAddress
	}

	p, reject := a.exportFilter.ProcessTerms(pfx, p)
	if reject {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.neighbor.CapAddPathRX {
		oldPaths := a.rt.ReplacePath(pfx, p)
		a.removePathsFromClients(pfx, oldPaths)
	}

	pathID, err := a.pathIDManager.addPath(p)
	if err != nil {
		return fmt.Errorf("Unable to get path ID: %v", err)
	}

	p.BGPPath.PathIdentifier = pathID
	a.rt.AddPath(pfx, p)

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
	pathID, err := a.pathIDManager.releasePath(p)
	if err != nil {
		log.Warningf("Unable to release path: %v", err)
		return true
	}

	p = p.Copy()
	p.BGPPath.PathIdentifier = pathID
	a.removePathFromClients(pfx, p)
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
