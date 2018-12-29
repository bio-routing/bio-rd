package server

import (
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type neighbor struct {
	systemID               types.SystemID
	ifa                    *netIf
	holdingTime            uint16
	localCircuitID         uint8
	extendedLocalCircuitID uint32
	fsm                    *FSM
}

func newNeighbor(sysID types.SystemID, ifa *netIf, extendedLocalCircuitID uint32) *neighbor {
	return &neighbor{
		systemID:               sysID,
		ifa:                    ifa,
		extendedLocalCircuitID: extendedLocalCircuitID,
	}
}
