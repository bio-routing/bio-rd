package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// RouteTableClient is the interface that every type of RIB must implement
type RouteTableClient interface {
	AddPath(net.Prefix, *route.Path) error
	RemovePath(net.Prefix, *route.Path) bool
	UpdateNewClient(RouteTableClient) error
	Register(RouteTableClient)
	RegisterWithOptions(RouteTableClient, ClientOptions)
	Unregister(RouteTableClient)
	RouteCount() int64
}
