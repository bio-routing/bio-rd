package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type SetNextHopAction struct {
	addr uint32
}

func NewSetNextHopAction(addr uint32) *SetNextHopAction {
	return &SetNextHopAction{
		addr: addr,
	}
}

func (a *SetNextHopAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil {
		return pa, false
	}

	modified := pa.Copy()
	modified.BGPPath.NextHop = a.addr

	return modified, false
}
