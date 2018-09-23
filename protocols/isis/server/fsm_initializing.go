package server

import "github.com/bio-routing/bio-rd/protocols/isis/packet"

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

			hello := pkt.Body.(packet.P2PHello)
			for _, tlv := range hello.TLVs {
				switch tlv.Type() {
				case packet.P2PAdjacencyStateTLVType:
					p2pAdjStateTLVFound = true
				case packet.ProtocolsSupportedTLVType:
					if !s.fsm.neighbor.ifa.compareSupportedProtocols(tlv.(packet.ProtocolsSupportedTLV).NerworkLayerProtocolIDs) {
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

			return newInitializingState(s.fsm), "Received valid P2PHello"
		default:
			return newDownState(s.fsm), "Received unexpected packet"
		}
	}
}

func (s initializingState) getState() uint8 {
	return packet.INITIALIZING_STATE
}
