package filter

import "github.com/bio-routing/bio-rd/net"

type PrefixMatcher interface {
	Match(pattern, prefix *net.Prefix) bool
	equal(PrefixMatcher) bool
}

type InRangeMatcher struct {
	min uint8
	max uint8
}

func NewInRangeMatcher(min, max uint8) *InRangeMatcher {
	return &InRangeMatcher{
		min: min,
		max: max,
	}
}

func (i *InRangeMatcher) Match(pattern, prefix *net.Prefix) bool {
	contains := pattern.Equal(prefix) || pattern.Contains(prefix)
	return contains && prefix.Len() >= i.min && prefix.Len() <= i.max
}

func (i *InRangeMatcher) equal(x PrefixMatcher) bool {
	switch x.(type) {
	case *InRangeMatcher:
	default:
		return false
	}

	y := x.(*InRangeMatcher)
	if i.min != y.min || i.max != y.max {
		return false
	}

	return true
}

type ExactMatcher struct{}

func NewExactMatcher() *ExactMatcher {
	return &ExactMatcher{}
}

func (e *ExactMatcher) Match(pattern, prefix *net.Prefix) bool {
	return pattern.Equal(prefix)
}

func (e *ExactMatcher) equal(x PrefixMatcher) bool {
	switch x.(type) {
	case *ExactMatcher:
	default:
		return false
	}

	return true
}

type OrLongerMatcher struct{}

func NewOrLongerMatcher() *OrLongerMatcher {
	return &OrLongerMatcher{}
}

func (e *OrLongerMatcher) Match(pattern, prefix *net.Prefix) bool {
	return pattern.Equal(prefix) || pattern.Contains(prefix)
}

func (e *OrLongerMatcher) equal(x PrefixMatcher) bool {
	switch x.(type) {
	case *OrLongerMatcher:
	default:
		return false
	}

	return true
}

type LongerMatcher struct{}

func NewLongerMatcher() *LongerMatcher {
	return &LongerMatcher{}
}

func (e *LongerMatcher) Match(pattern, prefix *net.Prefix) bool {
	return pattern.Contains(prefix) && prefix.Len() > pattern.Len()
}

func (e *LongerMatcher) equal(x PrefixMatcher) bool {
	switch x.(type) {
	case *LongerMatcher:
	default:
		return false
	}

	return true
}
