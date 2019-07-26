package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

// RouteTableClient is the interface that every type of RIB must implement
type RouteTableClient interface {
	AddPath(net.Prefix, *route.Path) error
	RemovePath(net.Prefix, *route.Path) bool
	ReplacePath(net.Prefix, *route.Path, *route.Path)
	UpdateNewClient(RouteTableClient) error
	Register(RouteTableClient)
	Unregister(RouteTableClient)
	RouteCount() int64
	ClientCount() uint64
	Dump() []*route.Route
	ReplaceFilterChain(filter.Chain)
	RefreshRoute(net.Prefix, []*route.Path)
}
