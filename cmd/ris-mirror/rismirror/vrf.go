package rismirror

import (
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/mergedlocrib"
)

type _vrf struct {
	ipv4Unicast *mergedlocrib.MergedLocRIB
	ipv6Unicast *mergedlocrib.MergedLocRIB
}

func newVRF(locRIBIPv4Unicast, locRIBIPv6Unicast *locRIB.LocRIB) *_vrf {
	return &_vrf{
		ipv4Unicast: mergedlocrib.New(locRIBIPv4Unicast),
		ipv6Unicast: mergedlocrib.New(locRIBIPv6Unicast),
	}
}

func (v *_vrf) getRIB(afi uint8) *mergedlocrib.MergedLocRIB {
	if afi == 6 {
		return v.ipv6Unicast
	}

	return v.ipv4Unicast
}
