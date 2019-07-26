package filter

import (
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
)

// NewAcceptAllFilter returns a filter accepting any paths/prefixes
func NewAcceptAllFilter() *Filter {
	return NewFilter(
		"ACCEPT_ALL",
		[]*Term{
			NewTerm(
				"ACCEPT_ALL",
				[]*TermCondition{},
				[]actions.Action{
					&actions.AcceptAction{},
				}),
		})
}

// NewAcceptAllFilterChain returns a filter chain that accepts any paths/prefixes
func NewAcceptAllFilterChain() Chain {
	return Chain{
		NewAcceptAllFilter(),
	}
}

// NewDrainFilter returns a filter rejecting any paths/prefixes
func NewDrainFilter() *Filter {
	return NewFilter(
		"REJECT_ALL",
		[]*Term{
			NewTerm(
				"REJECT_ALL",
				[]*TermCondition{},
				[]actions.Action{
					&actions.RejectAction{},
				}),
		})
}

// NewDrainFilterChain creates a filter chain that rejects any paths/prefixes
func NewDrainFilterChain() Chain {
	return Chain{
		NewDrainFilter(),
	}
}
