package actions

import (
	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type SetNextHopAction struct {
	ip bnet.IP
}

func NewSetNextHopAction(ip bnet.IP) *SetNextHopAction {
	return &SetNextHopAction{
		ip: ip,
	}
}

func (a *SetNextHopAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil {
		return pa, false
	}

	modified := pa.Copy()
	modified.BGPPath.NextHop = a.ip

	return modified, false
}
