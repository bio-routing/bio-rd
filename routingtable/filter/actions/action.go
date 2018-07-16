package actions

import (
	"github.com/bio-routing/bio-rd/route"
)

type Result struct {
	Path      *route.Path
	Reject    bool
	Terminate bool
}
