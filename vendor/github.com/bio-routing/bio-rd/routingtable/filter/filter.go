package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type FilterResult struct {
	Path      *route.Path
	Terminate bool
	Reject    bool
}

type Filter struct {
	name  string
	terms []*Term
}

func NewFilter(name string, terms []*Term) *Filter {
	f := &Filter{
		name:  name,
		terms: terms,
	}

	return f
}

// Name returns the name of the filter
func (f *Filter) Name() string {
	return f.name
}

// Process processes a filter
func (f *Filter) Process(p *net.Prefix, pa *route.Path) FilterResult {
	for _, t := range f.terms {
		res := t.Process(p, pa)
		if res.Terminate {
			return FilterResult{
				Path:      pa,
				Terminate: res.Terminate,
				Reject:    res.Reject,
			}
		}
	}

	return FilterResult{
		Path: pa,
	}
}

func (f *Filter) equal(x *Filter) bool {
	if len(f.terms) != len(x.terms) {
		return false
	}

	for i := range f.terms {
		if !f.terms[i].equal(x.terms[i]) {
			return false
		}
	}

	return true
}
