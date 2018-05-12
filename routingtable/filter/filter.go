package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

type Filter struct {
	routingtable.ClientManager
	terms []*Term
}

func NewFilter(terms []*Term) *Filter {
	f := &Filter{
		terms: terms,
	}
	f.ClientManager = routingtable.NewClientManager(f)

	return f
}

func (f *Filter) AddPath(p net.Prefix, pa *route.Path) error {
	pa, rejected := f.processTerms(p, pa)
	if rejected {
		return nil
	}

	for _, c := range f.Clients() {
		c.AddPath(p, pa)
	}

	return nil
}

func (f *Filter) RemovePath(p net.Prefix, pa *route.Path) bool {
	pa, rejected := f.processTerms(p, pa)
	if rejected {
		return false
	}

	for _, c := range f.Clients() {
		c.RemovePath(p, pa)
	}

	return true
}

func (f *Filter) UpdateNewClient(c routingtable.RouteTableClient) error {
	return nil
}

func (f *Filter) processTerms(p net.Prefix, pa *route.Path) (modPath *route.Path, reject bool) {
	modPath = pa

	for _, t := range f.terms {
		modPath, reject = t.Process(p, modPath)
		if reject {
			return modPath, true
		}
	}

	return modPath, false
}
