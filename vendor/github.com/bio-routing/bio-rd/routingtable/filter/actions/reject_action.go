package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// RejectAction rejects a prefix and terminates filter processing
type RejectAction struct {
}

func NewRejectAction() *RejectAction {
	return &RejectAction{}
}

func (*RejectAction) Do(p *net.Prefix, pa *route.Path) Result {
	return Result{
		Path:      pa,
		Reject:    true,
		Terminate: true,
	}
}

// Equal compares actions
func (a *RejectAction) Equal(b Action) bool {
	switch b.(type) {
	case *RejectAction:
	default:
		return false
	}

	return true
}
