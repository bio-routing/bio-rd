package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type setNextHopAction struct {
	addr uint32
}

func NewSetNextHopAction(addr uint32) filter.FilterAction {
	return &setNextHopAction{
		addr: addr,
	}
}

func (a *setNextHopAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil {
		return pa, false
	}

	modified := *pa
	modified.BGPPath.NextHop = a.addr

	return &modified, false
}
