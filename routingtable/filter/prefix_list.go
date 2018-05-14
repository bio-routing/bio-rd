package filter

import "github.com/bio-routing/bio-rd/net"

type PrefixList struct {
	allowed []net.Prefix
}

func NewPrefixList(pfxs ...net.Prefix) *PrefixList {
	l := &PrefixList{
		allowed: pfxs,
	}
	return l
}

func (l *PrefixList) Matches(p net.Prefix) bool {
	for _, a := range l.allowed {
		if !a.Contains(p) {
			return false
		}
	}

	return true
}
