package actions

import (
	"strings"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
)

type AddLargeCommunityAction struct {
	communities []*packet.LargeCommunity
}

func NewAddLargeCommunityAction(coms []*packet.LargeCommunity) *AddLargeCommunityAction {
	return &AddLargeCommunityAction{
		communities: coms,
	}
}

func (a *AddLargeCommunityAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil || len(a.communities) == 0 {
		return pa, false
	}

	modified := pa.Copy()

	for _, com := range a.communities {
		modified.BGPPath.LargeCommunities = modified.BGPPath.LargeCommunities + " " + com.String()
	}
	modified.BGPPath.LargeCommunities = strings.TrimLeft(modified.BGPPath.LargeCommunities, " ")

	return modified, false
}
