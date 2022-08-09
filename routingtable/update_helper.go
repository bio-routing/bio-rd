package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
)

// ShouldPropagateUpdate performs some default checks and returns if an route update should be propagated to a neighbor
func ShouldPropagateUpdate(pfx *net.Prefix, p *route.Path, sa *SessionAttrs) bool {
	return !isOwnPath(p, sa) && !isDisallowedByCommunity(p, sa)
}

func isOwnPath(p *route.Path, sa *SessionAttrs) bool {
	if p.Type != sa.Type {
		return false
	}

	switch p.Type {
	case route.BGPPathType:
		return p.BGPPath.BGPPathA.Source.Compare(sa.PeerIP) == 0
	}

	return false
}

func isDisallowedByCommunity(p *route.Path, sa *SessionAttrs) bool {
	if p.BGPPath == nil || (p.BGPPath.Communities != nil && len(*p.BGPPath.Communities) == 0) {
		return false
	}

	if p.BGPPath.Communities == nil {
		return false
	}

	for _, com := range *p.BGPPath.Communities {
		if (com == types.WellKnownCommunityNoExport && !sa.IBGP) || com == types.WellKnownCommunityNoAdvertise {
			return true
		}
	}

	return false
}
