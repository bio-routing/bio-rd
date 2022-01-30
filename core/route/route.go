package route

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
)

/*
Route Identifier:
1: Prefix
2: Source (BGP Neighbor / Protocol (IS-IS, OSPF, Kernel, etc))
3: Add-Path ID (optional)
4: MPLS Label Stack (optional, BGP-LU)
*/

const (
	asPathAttr = 1

	flagHidden       = 0x01
	flagValidNextHop = 0x02
	flagDiscard      = 0x04
	flagUnreachable  = 0x08
)

type RouteSource interface {
	AdministrativeDistance() uint8
}

type Route struct {
	pfx   *net.Prefix
	paths []*Path
}

type RouteReceiver interface {
	AdvertiseRoute(pfx *net.Prefix, p *Path) (changedAttrs []PathAttribute, hidden bool)
}

type Export struct {
	routeReceiver RouteReceiver
	changedAttrs  []PathAttribute
}

type Path struct {
	source     RouteSource
	flags      uint64
	attributes []PathAttribute
	exportedTo []Export
}

func (p *Path) isHidden() bool {
	return p.flags&flagHidden == flagHidden
}

func (p *Path) hasValidNextHop() bool {
	return p.flags&flagValidNextHop == flagValidNextHop
}

type PathAttribute interface {
	Type() uint16
	Value() interface{}
}

func (p *Path) getFirst(typ uint16) interface{} {
	for _, attr := range p.attributes {
		if attr.Type() == typ {
			return attr.Value()
		}
	}

	return nil
}

func (p *Path) getLast(typ uint16) interface{} {
	for i := len(p.attributes) - 1; i >= 0; i-- {
		if p.attributes[i].Type() == typ {
			return p.attributes[i].Value()
		}
	}

	return nil
}

func (p *Path) GetASPathPrePolicy() types.ASPath {
	asp := p.getFirst(asPathAttr)
	if asp != nil {
		return asp.(types.ASPath)
	}

	return nil
}

func (p *Path) GetASPathPostPolicy() types.ASPath {
	asp := p.getLast(asPathAttr)
	if asp != nil {
		return asp.(types.ASPath)
	}

	return nil
}
