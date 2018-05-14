package filter

import "github.com/bio-routing/bio-rd/net"

type PrefixMatcher func(pattern, prefix net.Prefix) bool

func InRange(min, max uint8) PrefixMatcher {
	return func(pattern, prefix net.Prefix) bool {
		contains := pattern.Equal(prefix) || pattern.Contains(prefix)
		return contains && prefix.Pfxlen() >= min && prefix.Pfxlen() <= max
	}
}

func Exact() PrefixMatcher {
	return func(pattern, prefix net.Prefix) bool {
		return pattern.Equal(prefix)
	}
}

func OrLonger() PrefixMatcher {
	return func(pattern, prefix net.Prefix) bool {
		return pattern.Equal(prefix) || pattern.Contains(prefix)
	}
}

func Longer() PrefixMatcher {
	return func(pattern, prefix net.Prefix) bool {
		return pattern.Contains(prefix) && prefix.Pfxlen() > pattern.Pfxlen()
	}
}
