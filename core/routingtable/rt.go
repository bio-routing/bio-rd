package routingtable

import (
	"github.com/bio-routing/bio-rd/core/route"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type routeSource interface {
	administrativeDistance() uint8
}

type RoutingTable struct {
	IndirectNextHops   []*IndirectNextHops
	NHResolutionPolicy *filter.Filter
}

type IdentifyingAttrs uint8

const (
	FlagAddPathID          = 0x01
	FlagLabelStack         = 0x02
	FlagRouteDistinguisher = 0x04
)

func (rt *RoutingTable) AddPath(pfx *net.Prefix, rs routeSource, p *route.Path, ident IdentifyingAttrs) {
	// If BGP nexthop, then recursive route lookup necessary => create IndirectNextHops object

	// After changing routing table we have to re-resolve all IndirectNextHops

	res := rt.NHResolutionPolicy.Process()
	res.Reject
}

type IndirectNextHops struct {
	addr *net.IP
}
