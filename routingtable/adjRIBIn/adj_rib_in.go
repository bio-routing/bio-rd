package adjRIBIn

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// AdjRIBIn represents an Adjacency RIB In as described in RFC4271
type AdjRIBIn struct {
	rt *routingtable.RoutingTable
	routingtable.ClientManager
}

// NewAdjRIBIn creates a new Adjacency RIB In
func NewAdjRIBIn() *AdjRIBIn {
	return &AdjRIBIn{
		rt: routingtable.NewRoutingTable(),
	}
}

// AddPath replaces the path for prefix `pfx`. If the prefix doesn't exist it is added.
func (a *AdjRIBIn) AddPath(pfx net.Prefix, p *route.Path) error {
	oldPaths := a.rt.ReplacePath(pfx, p)
	a.removePathsFromClients(pfx, oldPaths)
	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBIn) RemovePath(pfx net.Prefix, p *route.Path) error {
	r := a.rt.Get(pfx)
	if r == nil {
		return nil
	}

	oldPaths := r.Paths()
	for _, path := range oldPaths {
		a.rt.RemovePath(pfx, path)
	}

	a.removePathsFromClients(pfx, oldPaths)
	return nil
}

func (a *AdjRIBIn) removePathsFromClients(pfx net.Prefix, paths []*route.Path) {
	for _, path := range paths {
		for _, client := range a.ClientManager.Clients() {
			client.RemovePath(pfx, path)
		}
	}
}
