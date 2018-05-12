package filter

import "github.com/bio-routing/bio-rd/net"

type PrefixList struct {
	allowed []net.Prefix
}

func (f *PrefixList) Matches(p net.Prefix) bool {
	for _, a := range f.allowed {
		if !a.Contains(p) {
			return false
		}
	}

	return true
}
