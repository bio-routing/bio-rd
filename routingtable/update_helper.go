package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
)

// ShouldPropagateUpdate performs some default checks and returns if an route update should be propagated to a neighbor
func ShouldPropagateUpdate(pfx net.Prefix, p *route.Path, n *Neighbor) bool {
	return !isOwnPath(p, n) && !isDisallowedByCommunity(p, n)
}

func isOwnPath(p *route.Path, n *Neighbor) bool {
	if p.Type != n.Type {
		return false
	}

	switch p.Type {
	case route.BGPPathType:
		return p.BGPPath.Source == n.Address
	}

	return false
}

func isDisallowedByCommunity(p *route.Path, n *Neighbor) bool {
	if p.BGPPath == nil || len(p.BGPPath.Communities) == 0 {
		return false
	}

	for _, com := range p.BGPPath.Communities {
		if (com == types.WellKnownCommunityNoExport && !n.IBGP) || com == types.WellKnownCommunityNoAdvertise {
			return true
		}
	}

	return false
}
