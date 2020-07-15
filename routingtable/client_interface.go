package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
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

// AdjRIBIn is the interface any AdjRIBIn must implement
type AdjRIBIn interface {
	ReplaceFilterChain(filter.Chain)
	Dump() []*route.Route
	Register(client RouteTableClient)
	Unregister(client RouteTableClient)
	AddPath(pfx *net.Prefix, path *route.Path) error
	RemovePath(*net.Prefix, *route.Path) bool
	RouteCount() int64
	ClientCount() uint64
}

// AdjRIBOut is the interface any AdjRIBOut must implement
type AdjRIBOut interface {
	ReplaceFilterChain(filter.Chain)
	Dump() []*route.Route
	Register(client RouteTableClient)
	Unregister(client RouteTableClient)
	AddPath(pfx *net.Prefix, path *route.Path) error
	RemovePath(*net.Prefix, *route.Path) bool
	ReplacePath(*net.Prefix, *route.Path, *route.Path)
	RefreshRoute(*net.Prefix, []*route.Path)
	RouteCount() int64
	ClientCount() uint64
}
