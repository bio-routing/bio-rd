package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type Chain []*Filter

// Process processes a filter chain
func (c Chain) Process(p *net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
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

// Equal compares twp filter chains
func (c Chain) Equal(d Chain) bool {
	if len(c) != len(d) {
		return false
	}

	for i := range c {
		if !c[i].equal(d[i]) {
			return false
		}
	}

	return true
}
