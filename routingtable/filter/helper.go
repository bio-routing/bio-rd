package filter

import (
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
)

// NewAcceptAllFilter returns a filter accepting any paths/prefixes
func NewAcceptAllFilter() *Filter {
	return NewFilter(
		[]*Term{
			NewTerm(
				[]*TermCondition{},
				[]Action{
					&actions.AcceptAction{},
				}),
		})
}

// NewDrainFilter returns a filter rejecting any paths/prefixes
func NewDrainFilter() *Filter {
	return NewFilter(
		[]*Term{
			NewTerm(
				[]*TermCondition{},
				[]Action{
					&actions.RejectAction{},
				}),
		})
}
