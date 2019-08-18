package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
)

// Term matches a path against a list of conditions and performs actions if it matches
type Term struct {
	name string
	from []*TermCondition
	then []actions.Action
}

type TermResult struct {
	Path      *route.Path
	Terminate bool
	Reject    bool
}

// NewTerm creates a new term
func NewTerm(name string, from []*TermCondition, then []actions.Action) *Term {
	t := &Term{
		name: name,
		from: from,
		then: then,
	}

	return t
}

// Process processes a path returning if the path should be rejected and returns a possible modified version of the path
func (t *Term) Process(p *net.Prefix, pa *route.Path) TermResult {
	if len(t.from) == 0 {
		return t.processActions(p, pa)
	}

	for _, f := range t.from {
		if f.Matches(p, pa) {
			return t.processActions(p, pa)
		}
	}

	return TermResult{Path: pa}
}

func (t *Term) processActions(p *net.Prefix, pa *route.Path) TermResult {
	for _, action := range t.then {
		res := action.Do(p, pa)
		if res.Terminate {
			return TermResult{
				Path:      pa,
				Terminate: true,
				Reject:    res.Reject,
			}
		}
		pa = res.Path
	}

	return TermResult{Path: pa}
}

func (t *Term) equal(x *Term) bool {
	if len(t.from) != len(x.from) {
		return false
	}

	if len(t.then) != len(x.then) {
		return false
	}

	for i := range t.from {
		if !t.from[i].equal(x.from[i]) {
			return false
		}
	}

	for i := range t.then {
		if !t.then[i].Equal(x.then[i]) {
			return false
		}
	}

	return true
}
