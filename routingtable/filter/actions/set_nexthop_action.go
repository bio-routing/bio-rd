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

func (a *SetNextHopAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	pa.BGPPath.NextHop = a.ip

	return Result{Path: pa}
}
