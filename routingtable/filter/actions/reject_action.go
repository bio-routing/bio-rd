package actions

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type RejectAction struct {
}

func (*RejectAction) Do(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	return pa, true
}
