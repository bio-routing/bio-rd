package config

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
)

type PolicyOptions struct {
	// description: |
	//   Policy statements to filter route imports and exports
	//   Example:
	//   - name: "PeerA-In"
	//       terms:
	//         - name: "Reject_certain_stuff"
	//           from:
	//             route_filters:
	//                - prefix: "198.51.100.0/24"
	//                  matcher: "orlonger"
	//                - prefix: "203.0.113.0/25"
	//                  matcher: "exact"
	//                - prefix: "203.0.113.128/25"
	//                  matcher: "exact"
	PolicyStatements []*PolicyStatement `yaml:"policy_statements"`
	// docgen:nodoc
	PolicyStatementsFilter []*filter.Filter
	// description: |
	//   prefix lists to be used in the policy statements.
	//   Example:
	//     prefix-lists:
	//       prefixes:
	//         - 2001:db8:0:1::/64
	PrefixLists []PrefixList `yaml:"prefix_lists"`
}

type PrefixList struct {
	// description: |
	//   List of prefixes
	Prefixes []string `yaml:"prefixes"`
}

type PolicyStatement struct {
	// description: |
	//   Name of the policy statement
	Name string `yaml:"name"`
	// description: |
	//   List of terms defining the policy (see example above)
	Terms []*PolicyStatementTerm `yaml:"terms"`
}

type PolicyStatementTerm struct {
	// description: |
	//   Name of the term
	Name string `yaml:"name"`
	// description: |
	//   Filter to match the term
	From PolicyStatementTermFrom `yaml:"from"`
	// description: |
	//   Action to execute if the filter matches
	//   Available actions are:
	//     - Accept: accepts the route without modifications
	//     - Reject: rejects the route
	//     - MED: sets the MED to the specified value (max 4294967295)
	//     - LocalPref sets the local preference to the specified value (max 4294967295)
	//     - AsPathPrepend: prepends AS numbers to the route. Details bellow
	//     - NextHop: modify the next-hop to the specified address
	Then PolicyStatementTermThen `yaml:"then"`
}

type PolicyStatementTermFrom struct {
	// description: |
	//   List of route filters to match incoming packets
	//   Example:
	//     route_filters:
	//        - prefix: "198.51.100.0/24"
	//          matcher: "orlonger"
	RouteFilters []*RouteFilter `yaml:"route_filters"`
}

type RouteFilter struct {
	// description: |
	//   Prefix to match. Defined in CIDR notation
	Prefix string `yaml:"prefix"`
	// description: |
	//   Qualifier for the filter.
	//   Available options:
	//     - exact: matches only the exact prefix
	//     - orlonger: matches if the prefix equals the filter, or has a longer prefix length
	//     - longer: matches if the route has a longer prefix length than the one defined in the filter
	//     - range: the route falls between the values defined in len_min and len_max
	Matcher string `yaml:"matcher"`
	// description: |
	//   minimum length of the range
	LenMin uint8 `yaml:"len_min"`
	// description: |
	//   maximum lange of the range
	LenMax uint8 `yaml:"len_max"`
}

type PolicyStatementTermThen struct {
	// description: |
	//   accept the route
	Accept bool `yaml:"accept"`
	// description: |
	//   reject the route
	Reject bool `yaml:"reject"`
	// description: |
	//   Multi-exit discriminator
	MED *uint32 `yaml:"med"`
	// description: |
	//   Local preference
	LocalPref *uint32 `yaml:"local_pref"`
	// description: |
	//   ASN to prepend.
	//     Values:
	//       - ASN: asn number to prepend
	//       - count: amount of times that the number will be prepended
	ASPathPrepend *ASPathPrepend `yaml:"as_path_prepend"`
	// description: |
	//   IP address to be used as a next-hop for the route
	NextHop *NextHop `yaml:"next_hop"`
}

type ASPathPrepend struct {
	// description: |
	//   AS number
	ASN uint32 `yaml:"asn"`
	// description: |
	//   times to prepend
	Count uint16 `yaml:"count"`
}

type NextHop struct {
	// description: |
	//   IP address to be used as next hop
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
