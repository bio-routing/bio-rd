package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// RouteTable is the interface that every route table must implement
type RouteTable interface {
	AddPath(pfx *net.Prefix, path *route.Path) error
	RemovePath(*net.Prefix, *route.Path) bool
	UpdateNewClient(RouteTableClient) error
}

// RouteTableClient is the interface that every route table client must implement
type RouteTableClient interface {
	AddPath(pfx *net.Prefix, path *route.Path) error
	RemovePath(*net.Prefix, *route.Path) bool
	ReplacePath(*net.Prefix, *route.Path, *route.Path)
	RefreshRoute(*net.Prefix, []*route.Path)
}
