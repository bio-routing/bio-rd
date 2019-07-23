package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// SetMEDAction sets the BGP MED
type SetMEDAction struct {
	med uint32
}

// NewSetMEDAction creates new SetMEDAction
func NewSetMEDAction(med uint32) *SetMEDAction {
	return &SetMEDAction{
		med: med,
	}
}

// Do applies the action
func (a *SetMEDAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	modified := *pa
	modified.BGPPath.MED = a.med

	return Result{Path: &modified}
}
