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

func (a *SetNextHopAction) Do(p *net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	modified := pa.Copy()
	modified.BGPPath.BGPPathA.NextHop = a.ip

	return Result{Path: modified}
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
