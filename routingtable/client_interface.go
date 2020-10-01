package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

// RouteTableClient is the interface that every route table client must implement
type RouteTableClient interface {
	AddPath(pfx *net.Prefix, path *route.Path) error
	AddPathInitialDump(pfx *net.Prefix, path *route.Path) error
	RemovePath(*net.Prefix, *route.Path) bool
	ReplacePath(*net.Prefix, *route.Path, *route.Path)
	RefreshRoute(*net.Prefix, []*route.Path)
}

type AdjRIB interface {
	ReplaceFilterChain(filter.Chain)
	Dump() []*route.Route
	Register(client RouteTableClient)
	Unregister(client RouteTableClient)
	AddPath(pfx *net.Prefix, path *route.Path) error
	RemovePath(*net.Prefix, *route.Path) bool
	RouteCount() int64
	ClientCount() uint64
}

// AdjRIBIn is the interface any AdjRIBIn must implement
type AdjRIBIn interface {
	AdjRIB
}

// AdjRIBOut is the interface any AdjRIBOut must implement
type AdjRIBOut interface {
	AdjRIB
	AddPathInitialDump(pfx *net.Prefix, path *route.Path) error
	ReplacePath(*net.Prefix, *route.Path, *route.Path)
	RefreshRoute(*net.Prefix, []*route.Path)
}
