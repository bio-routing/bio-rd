package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type setLocalPrefAction struct {
	pref uint32
}

func NewSetLocalPrefAction(pref uint32) filter.FilterAction {
	return &setLocalPrefAction{
		pref: pref,
	}
}

func (a *setLocalPrefAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	if pa.BGPPath == nil {
		return pa, false
	}

	modified := *pa
	modified.BGPPath.LocalPref = a.pref

	return &modified, false
}
