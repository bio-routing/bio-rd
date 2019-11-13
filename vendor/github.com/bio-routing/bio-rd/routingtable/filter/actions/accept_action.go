package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// AcceptAction accepts a path and terminates processing of filters
type AcceptAction struct {
}

// NewAcceptAction returns a new AcceptAction
func NewAcceptAction() *AcceptAction {
	return &AcceptAction{}
}

// Do applies the action
func (*AcceptAction) Do(p *net.Prefix, pa *route.Path) Result {
	return Result{
		Path:      pa,
		Terminate: true,
	}
}

// Equal compares actions
func (a *AcceptAction) Equal(b Action) bool {
	switch b.(type) {
	case *AcceptAction:
	default:
		return false
	}

	return true
}
