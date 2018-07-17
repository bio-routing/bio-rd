package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type Filter struct {
	terms []*Term
}

func NewFilter(terms []*Term) *Filter {
	f := &Filter{
		terms: terms,
	}

	return f
}

func (f *Filter) ProcessTerms(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	modPath = pa

	for _, t := range f.terms {
		res := t.Process(p, modPath)
		if res.Terminate {
			return res.Path, res.Reject
		}
		modPath = res.Path
	}

	return modPath, false
}
