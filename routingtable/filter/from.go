package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type From struct {
	prefixList *PrefixList
}

func (f *From) Matches(p net.Prefix, pa *route.Path) bool {
	return f.prefixList.Matches(p)
}
