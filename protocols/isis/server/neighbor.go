package server

import "github.com/bio-routing/bio-rd/protocols/isis/types"

type neighbor struct {
	systemID               types.SystemID
	dev                    *dev
	holdingTime            uint16
	localCircuitID         uint8
	extendedLocalCircuitID uint32
	ipInterfaceAddresses   []uint32
	//fsm                    *FSM
}
