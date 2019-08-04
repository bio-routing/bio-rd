package server

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

const ospfVersion = 3

// OSPFServer handles the OSPFv3 routing protocol
type OSPFServer interface{}

type RoutingTable interface {
	AddPath(net.Prefix, *route.Path) error
	ReplacePath(net.Prefix, *route.Path) []*route.Path
	RemovePath(net.Prefix, *route.Path)
	RemovePfx(net.Prefix) []*route.Path
	LPM(net.Prefix) []*route.Route
	Get(net.Prefix) *route.Route
	GetLonger(net.Prefix) []*route.Route
	Dump() []*route.Route
}
