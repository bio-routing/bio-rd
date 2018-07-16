package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type SetLocalPrefAction struct {
	pref uint32
}

func NewSetLocalPrefAction(pref uint32) *SetLocalPrefAction {
	return &SetLocalPrefAction{
		pref: pref,
	}
}

func (a *SetLocalPrefAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	modified := *pa
	modified.BGPPath.LocalPref = a.pref

	return Result{Path: &modified}
}
