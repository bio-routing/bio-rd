package rismirror

import (
	"github.com/bio-routing/bio-rd/risclient"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/mergedlocrib"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

type vrfWithMergedLocRIBs struct {
	vrf         *vrf.VRF
	ipv4Unicast *mergedlocrib.MergedLocRIB
	ipv6Unicast *mergedlocrib.MergedLocRIB
	Clients     []*risclient.RISClient
}

func newVRFWithMergedLocRIBs(locRIBIPv4Unicast, locRIBIPv6Unicast *locRIB.LocRIB, v *vrf.VRF) *vrfWithMergedLocRIBs {
	return &vrfWithMergedLocRIBs{
		vrf:         v,
		ipv4Unicast: mergedlocrib.New(locRIBIPv4Unicast),
		ipv6Unicast: mergedlocrib.New(locRIBIPv6Unicast),
		Clients:     make([]*risclient.RISClient, 0, 2),
	}
}

func (v *vrfWithMergedLocRIBs) getRIB(afi uint8) *mergedlocrib.MergedLocRIB {
	if afi == 6 {
		return v.ipv6Unicast
	}

	return v.ipv4Unicast
}
