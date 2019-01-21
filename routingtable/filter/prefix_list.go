package filter

import "github.com/bio-routing/bio-rd/net"

type PrefixList struct {
	allowed []net.Prefix
	matcher PrefixMatcher
}

func NewPrefixList(pfxs ...net.Prefix) *PrefixList {
	return &PrefixList{
		allowed: pfxs,
		matcher: Exact(),
	}
}

func NewPrefixListWithMatcher(matcher PrefixMatcher, pfxs ...net.Prefix) *PrefixList {
	return &PrefixList{
		allowed: pfxs,
		matcher: matcher,
	}
}

func (l *PrefixList) Matches(p net.Prefix) bool {
	for _, a := range l.allowed {
		if a.Equal(p) {
			return true
		}
	}

	return false
}
