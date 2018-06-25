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
		modPath, reject = t.Process(p, modPath)
		if reject {
			return modPath, true
		}
	}

	return modPath, false
}
