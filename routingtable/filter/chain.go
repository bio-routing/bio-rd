package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type Chain []*Filter

// Process processes a filter chain
func (c Chain) Process(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	mp := pa.Copy()
	for _, f := range c {
		res := f.Process(p, mp)
		if res.Terminate {
			return res.Path, res.Reject
		}

		mp = res.Path
	}

	return mp, false
}
