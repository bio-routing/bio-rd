package actions

import (
	"fmt"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type SetMEDAction struct {
	med uint32
}

func NewSetMEDAction(med uint32) *SetMEDAction {
	return &SetMEDAction{
		med: med,
	}
}

func (a *SetMEDAction) Do(p net.Prefix, pa *route.Path) Result {
	if pa.BGPPath == nil {
		return Result{Path: pa}
	}

	fmt.Printf("Setting MED %d\n", a.med)
	modified := *pa
	modified.BGPPath.MED = a.med

	return Result{Path: &modified}
}
