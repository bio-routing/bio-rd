package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

// Action performs actions on a `route.Path`
type Action interface {
	Do(p *net.Prefix, pa *route.Path) Result
	Equal(x Action) bool
}

// Result is a filter result
type Result struct {
	Path      *route.Path
	Reject    bool
	Terminate bool
}
