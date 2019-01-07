package server

import (
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type neighbor struct {
	systemID               types.SystemID
	ifa                    *netIf
	holdingTime            uint16
	localCircuitID         uint8
	extendedLocalCircuitID uint32
	fsm                    *FSM
	ipInterfaceAddresses   []uint32
}

func newNeighbor(sysID types.SystemID, ifa *netIf, extendedLocalCircuitID uint32) *neighbor {
	return &neighbor{
		systemID:               sysID,
		ifa:                    ifa,
		extendedLocalCircuitID: extendedLocalCircuitID,
		ipInterfaceAddresses:   make([]uint32, 0),
	}
}

func (n *neighbor) getExtendedISReachabilityNeighbor() *packet.ExtendedISReachabilityNeighbor {
	eirn := packet.NewExtendedISReachabilityNeighbor(
		types.NewSourceID(
			n.systemID,
			n.localCircuitID,
		),
		n.ifa.l2.get3ByteMetric(),
	)

	for i := range n.ifa.device.Addrs {
		if !n.ifa.device.Addrs[i].Addr().IsIPv4() {
			continue
		}

		eirn.AddSubTLV(packet.NewIPv4InterfaceAddressSubTLV(n.ifa.device.Addrs[i].Addr().ToUint32()))
	}

	for i := range n.ipInterfaceAddresses {
		eirn.AddSubTLV(packet.NewIPv4NeighborAddressSubTLV(n.ipInterfaceAddresses[i]))
	}

	eirn.AddSubTLV(packet.NewLinkLocalRemoteIdentifiersSubTLV(uint32(n.ifa.device.Index), n.extendedLocalCircuitID))
	return eirn
}
