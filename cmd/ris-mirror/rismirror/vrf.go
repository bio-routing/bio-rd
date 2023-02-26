package rismirror

import (
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/mergedlocrib"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

type vrfWithMergedLocRIBs struct {
	vrf         *vrf.VRF
	ipv4Unicast *mergedlocrib.MergedLocRIB
	ipv6Unicast *mergedlocrib.MergedLocRIB
}

func newVRFWithMergedLocRIBs(locRIBIPv4Unicast, locRIBIPv6Unicast *locRIB.LocRIB) *vrfWithMergedLocRIBs {
	return &vrfWithMergedLocRIBs{
		ipv4Unicast: mergedlocrib.New(locRIBIPv4Unicast),
		ipv6Unicast: mergedlocrib.New(locRIBIPv6Unicast),
	}
}

func (v *vrfWithMergedLocRIBs) getRIB(afi uint8) *mergedlocrib.MergedLocRIB {
	if afi == 6 {
		return v.ipv6Unicast
	}

	return v.ipv4Unicast
}
