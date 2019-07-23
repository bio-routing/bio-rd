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

	pa.BGPPath.LocalPref = a.pref
	return Result{Path: pa}
}

// Equal compares actions
func (a *SetLocalPrefAction) Equal(b Action) bool {
	switch b.(type) {
	case *SetLocalPrefAction:
	default:
		return false
	}

	return true
}
