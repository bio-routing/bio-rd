package server

import "github.com/bio-routing/bio-rd/protocols/isis/packet"

type downState struct {
	fsm *FSM
}

func newDownState(fsm *FSM) *downState {
	return &downState{
		fsm: fsm,
	}
}

func (s *downState) getState() uint8 {
	return packet.DOWN_STATE
}

func (s *downState) run() (state, string) {
	pkt := <-s.fsm.pktCh

	for {
		if pkt.Header.PDUType != packet.P2P_HELLO {
			continue
		}

		h := pkt.Body.(packet.P2PHello)
		p2pAdjTLV := h.GetP2PAdjTLV()
		if p2pAdjTLV != nil {
			if p2pAdjTLV.Length() == packet.P2PAdjacencyStateTLVLenWithoutNeighbor {
				return newInitializingState(s.fsm), "Received P2P Hello with Adjacency TLV"
			}

			if p2pAdjTLV.NeighborSystemID == s.fsm.isisServer.config.NETs[0].SystemID {
				return newUpState(s.fsm), "Received P2P Hello with Adjacency TLV including us"
			}

			return newDownState(s.fsm), "Received P2P Hello with Adjacency TLV including someone else"
		}

	}
}
