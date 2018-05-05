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

	for _, oldPath := range oldPaths {
		for _, client := range a.ClientManager.Clients() {
			client.RemovePath(pfx, oldPath)
		}
	}

	return nil
}

// RemovePath removes the path for prefix `pfx`
func (a *AdjRIBIn) RemovePath(pfx net.Prefix, p *route.Path) error {
	if !a.rt.RemovePath(pfx, p) {
		return nil
	}

	for _, client := range a.ClientManager.Clients() {
		client.RemovePath(pfx, p)
	}

	return nil
}
