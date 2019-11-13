package filter

import "github.com/bio-routing/bio-rd/net"

type PrefixList struct {
	allowed []*net.Prefix
	matcher PrefixMatcher
}

func NewPrefixList(pfxs ...*net.Prefix) *PrefixList {
	l := &PrefixList{
		allowed: pfxs,
		matcher: NewExactMatcher(),
	}
	return l
}

func NewPrefixListWithMatcher(matcher PrefixMatcher, pfxs ...*net.Prefix) *PrefixList {
	l := &PrefixList{
		allowed: pfxs,
		matcher: matcher,
	}
	return l
}

func (l *PrefixList) Matches(p *net.Prefix) bool {
	for _, a := range l.allowed {
		if a.Equal(p) {
			return true
		}
	}

	return false
}
