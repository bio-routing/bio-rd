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

func (a *SetLocalPrefAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil {
		return pa, false
	}

	modified := *pa
	modified.BGPPath.LocalPref = a.pref

	return &modified, false
}
