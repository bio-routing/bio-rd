package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// RouteTableClient is the interface that every type of RIB must implement
type RouteTableClient interface {
	AddPath(net.Prefix, *route.Path) error
	// TODO: Need some description what the returned bool should indicate
	RemovePath(net.Prefix, *route.Path) bool
	UpdateNewClient(RouteTableClient) error
	Register(RouteTableClient)
	Unregister(RouteTableClient)
	RouteCount() int64
	ClientCount() uint64
	Dump() []*route.Route
}
