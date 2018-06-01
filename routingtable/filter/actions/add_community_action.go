package actions

import (
	"strings"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
)

type AddCommunityAction struct {
	communities []uint32
}

func NewAddCommunityAction(coms []uint32) *AddCommunityAction {
	return &AddCommunityAction{
		communities: coms,
	}
}

func (a *AddCommunityAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil || len(a.communities) == 0 {
		return pa, false
	}

	modified := pa.Copy()

	for _, com := range a.communities {
		modified.BGPPath.Communities = modified.BGPPath.Communities + " " + packet.CommunityStringForUint32(com)
	}
	modified.BGPPath.Communities = strings.TrimLeft(modified.BGPPath.Communities, " ")

	return modified, false
}
