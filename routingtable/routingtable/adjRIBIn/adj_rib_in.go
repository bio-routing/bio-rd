package adjRIBIn

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable"
)

// AdjRIBIn implements an Adjacency RIB In (for use with BGP)
type AdjRIBIn struct {
	rt *routingtable.RoutingTable
}

// New creates a new Adjacency RIB In
func New() *AdjRIBIn {
	return &AdjRIBIn{
		rt: routingtable.New(),
	}
}

// AddPath adds a route to the AdjRIBIn. If prefix exists already it is replaced.
func (a *AdjRIBIn) AddPath(pfx *net.Prefix, p *routingtable.Path) {

}

// RemovePath removes prefix pfx from the AdjRIBIn
func (a *AdjRIBIn) RemovePath(pfx *net.Prefix, p *routingtable.Path) {

}
