package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

type upState struct {
	fsm *FSM
}

func newUpState(fsm *FSM) *upState {
	return &upState{
		fsm: fsm,
	}
}

func (s upState) run() (state, string) {
	for {
		pkt := <-s.fsm.pktCh
		switch pkt.Header.PDUType {
		case packet.L2_LS_PDU_TYPE:
			fmt.Printf("LSPDU Received\n")
		case packet.L2_CSNP_TYPE:
			fmt.Printf("CSNP received\n")
		case packet.L2_PSNP_TYPE:
			fmt.Printf("PSNP received\n")
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
					if tlv.(packet.P2PAdjacencyStateTLV).NeighborSystemID == s.fsm.neighbor.ifa.isisServer.systemID() {
						foundSelf = true
					}
				case packet.ProtocolsSupportedTLVType:
					if !s.fsm.neighbor.ifa.compareSupportedProtocols(tlv.(*packet.ProtocolsSupportedTLV).NerworkLayerProtocolIDs) {
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
				return newUpState(s.fsm), "Found myself in P2PAdjacencyTLV"
			}

			return newInitializingState(s.fsm), "Received valid P2PHello"
		default:
			return newDownState(s.fsm), "Received unexpected packet"
		}
	}
}

func (s upState) getState() uint8 {
	return packet.UP_STATE
}
