package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
)

type AddCommunityAction struct {
	communities *types.Communities
}

func NewAddCommunityAction(coms *types.Communities) *AddCommunityAction {
	return &AddCommunityAction{
		communities: coms,
	}
}

func (a *AddCommunityAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil || len(*a.communities) == 0 {
		return Result{Path: pa}
	}

	modified := pa.Copy()
	for _, com := range *a.communities {
		if modified.BGPPath.Communities == nil {
			modified.BGPPath.Communities = &types.Communities{}
		}

		*modified.BGPPath.Communities = append(*modified.BGPPath.Communities, com)
	}

	return Result{Path: modified}
}
