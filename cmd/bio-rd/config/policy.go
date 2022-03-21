package config

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
)

type PolicyOptions struct {
	PolicyStatements       []*PolicyStatement `yaml:"policy_statements"`
	PolicyStatementsFilter []*filter.Filter
	PrefixLists            []PrefixList `yaml:"prefix_lists"`
}

type PrefixList struct {
	Prefixes []string `yaml:"prefixes"`
}

type PolicyStatement struct {
	Name  string                 `yaml:"name"`
	Terms []*PolicyStatementTerm `yaml:"terms"`
}

type PolicyStatementTerm struct {
	Name string                  `yaml:"name"`
	From PolicyStatementTermFrom `yaml:"from"`
	Then PolicyStatementTermThen `yaml:"then"`
}

type PolicyStatementTermFrom struct {
	RouteFilters []*RouteFilter `yaml:"route_filters"`
}

type RouteFilter struct {
	Prefix  string `yaml:"prefix"`
	Matcher string `yaml:"matcher"`
	LenMin  uint8  `yaml:"len_min"`
	LenMax  uint8  `yaml:"len_max"`
}

type PolicyStatementTermThen struct {
	Accept        bool           `yaml:"accept"`
	Reject        bool           `yaml:"reject"`
	MED           *uint32        `yaml:"med"`
	LocalPref     *uint32        `yaml:"local_pref"`
	ASPathPrepend *ASPathPrepend `yaml:"as_path_prepend"`
	NextHop       *NextHop       `yaml:"next_hop"`
}

type ASPathPrepend struct {
	ASN   uint32 `yaml:"asn"`
	Count uint16 `yaml:"count"`
}

type NextHop struct {
	Address string `yaml:"address"`
}

func (rf *RouteFilter) toFilterRouteFilter() (*filter.RouteFilter, error) {
	pfx, err := bnet.PrefixFromString(rf.Prefix)
	if err != nil {
		return nil, fmt.Errorf("Invalid prefix: %w", err)
	}

	var m filter.PrefixMatcher
	switch rf.Matcher {
	case "exact":
		m = filter.NewExactMatcher()
	case "orlonger":
		m = filter.NewOrLongerMatcher()
	case "longer":
		m = filter.NewLongerMatcher()
	case "range":
		m = filter.NewInRangeMatcher(rf.LenMin, rf.LenMax)
	default:
		return nil, fmt.Errorf("Invalid matcher: %q", rf.Matcher)
	}

	return filter.NewRouteFilter(pfx, m), nil
}

func (po *PolicyOptions) getPolicyStatementFilter(name string) *filter.Filter {
	for _, f := range po.PolicyStatementsFilter {
		if f.Name() == name {
			return f
		}

	}

	return nil
}

func (po *PolicyOptions) load() error {
	for _, ps := range po.PolicyStatements {
		f, err := ps.toFilter()
		if err != nil {
			return fmt.Errorf("Failed to convert policy_statement: %w", err)
		}

		po.PolicyStatementsFilter = append(po.PolicyStatementsFilter, f)
	}

	return nil
}

func (ps *PolicyStatement) toFilter() (*filter.Filter, error) {
	terms := make([]*filter.Term, 0)

	for _, t := range ps.Terms {
		ft, err := t.toFilterTerm()
		if err != nil {
			return nil, fmt.Errorf("unable to process filter term: %w", err)
		}

		terms = append(terms, ft)
	}

	return filter.NewFilter(ps.Name, terms), nil
}

func (pst *PolicyStatementTerm) toFilterTerm() (*filter.Term, error) {
	conditions := make([]*filter.TermCondition, 0)
	a := make([]actions.Action, 0)

	routeFilters := make([]*filter.RouteFilter, 0)
	for i := range pst.From.RouteFilters {
		rf, err := pst.From.RouteFilters[i].toFilterRouteFilter()
		if err != nil {
			return nil, fmt.Errorf("unable to parse route filter: %w", err)
		}

		routeFilters = append(routeFilters, rf)
	}

	if len(routeFilters) > 0 {
		conditions = append(conditions, filter.NewTermConditionWithRouteFilters(routeFilters...))
	}

	if pst.Then.Reject {
		a = append(a, actions.NewRejectAction())
	}

	if pst.Then.LocalPref != nil {
		a = append(a, actions.NewSetLocalPrefAction(*pst.Then.LocalPref))
	}

	if pst.Then.MED != nil {
		a = append(a, actions.NewSetMEDAction(*pst.Then.MED))
	}

	if pst.Then.ASPathPrepend != nil {
		a = append(a, actions.NewASPathPrependAction(pst.Then.ASPathPrepend.ASN, pst.Then.ASPathPrepend.Count))
	}

	if pst.Then.NextHop != nil {
		addr, err := bnet.IPFromString(pst.Then.NextHop.Address)
		if err != nil {
			return nil, fmt.Errorf("Invalid next_hop address: %w", err)
		}

		a = append(a, actions.NewSetNextHopAction(addr.Dedup()))
	}

	if pst.Then.Accept {
		a = append(a, actions.NewAcceptAction())
	}

	return filter.NewTerm(pst.Name, conditions, a), nil
}
