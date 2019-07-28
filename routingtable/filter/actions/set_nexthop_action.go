package actions

import (
	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type SetNextHopAction struct {
	ip *bnet.IP
}

func NewSetNextHopAction(ip *bnet.IP) *SetNextHopAction {
	return &SetNextHopAction{
		ip: ip.Dedup(),
	}
}

func (a *SetNextHopAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	pa.BGPPath.NextHop = a.ip

	return Result{Path: pa}
}

// Equal compares actions
func (a *SetNextHopAction) Equal(b Action) bool {
	switch b.(type) {
	case *SetNextHopAction:
	default:
		return false
	}

	if a.ip != b.(*SetNextHopAction).ip {
		return false
	}

	return true
}
