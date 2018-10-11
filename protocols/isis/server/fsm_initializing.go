package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

type initializingState struct {
	fsm *FSM
}

func newInitializingState(fsm *FSM) *initializingState {
	return &initializingState{
		fsm: fsm,
	}
}

func (s initializingState) run() (state, string) {
	for {
		pkt := <-s.fsm.pktCh
		switch pkt.Header.PDUType {
		case packet.P2P_HELLO:
			if !s.fsm.neighbor.ifa.p2p {
				return newDownState(s.fsm), "Received P2PHello on non-P2P interface"
			}

			p2pAdjStateTLVFound := false
			protocolsSupportedTLVfound := false
			areaAddressesTLVFound := false
			foundSelf := false

			hello := pkt.Body.(*packet.P2PHello)
			for _, tlv := range hello.TLVs {
				fmt.Printf("### TLV Type: %d\n", tlv.Type())
				switch tlv.Type() {
				case packet.P2PAdjacencyStateTLVType:
					p2pAdjStateTLVFound = true
					if tlv.(*packet.P2PAdjacencyStateTLV).NeighborSystemID == s.fsm.neighbor.ifa.isisServer.systemID() {
						foundSelf = true
					}
				case packet.ProtocolsSupportedTLVType:
					if !s.fsm.neighbor.ifa.compareSupportedProtocols(tlv.(*packet.ProtocolsSupportedTLV).NetworkLayerProtocolIDs) {
						return newDownState(s.fsm), "Supported protocols mismatch"
					}
					protocolsSupportedTLVfound = true
				case packet.AreaAddressesTLVType:
					areaAddressesTLVFound = true
				}
			}

			if !p2pAdjStateTLVFound {
				return newDownState(s.fsm), "Received P2PHello without P2P adjacency state TLV"
			}
			if !protocolsSupportedTLVfound {
				return newDownState(s.fsm), "Received P2PHello without protocols supported TLV"
			}
			if !areaAddressesTLVFound {
				return newDownState(s.fsm), "Received P2PHello without area address TLV"
			}

			if foundSelf {
				s.fsm.neighbor.ifa.isisServer.lsdb.setSRMAny(s.fsm.neighbor.ifa)
				newState := newUpState(s.fsm)
				go newState.sender()
				return newState, "Found myself in P2PAdjacencyTLV"
			}

			return newInitializingState(s.fsm), "Received valid P2PHello"
		default:
			return newDownState(s.fsm), "Received unexpected packet"
		}
	}
}

func (s initializingState) getState() uint8 {
	return packet.INITIALIZING_STATE
}
