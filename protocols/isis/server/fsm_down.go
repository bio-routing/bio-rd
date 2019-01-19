package server

import "github.com/bio-routing/bio-rd/protocols/isis/packet"

type fsmDownState struct {
	fsm *fsm
}

func newFSMDownState(fsm *fsm) *fsmDownState {
	return &fsmDownState{
		fsm: fsm,
	}
}

func (s *fsmDownState) getState() uint8 {
	return packet.DOWN_STATE
}

func (s *fsmDownState) run() (state, string) {
	for {
		pkt := <-s.fsm.pktCh

		if pkt.Header.PDUType != packet.P2P_HELLO {
			continue
		}

		if !s.fsm.neighbor.dev.p2p {
			return newFSMDownState(s.fsm), "Received P2PHello on non-P2P interface"
		}

		h := pkt.Body.(*packet.P2PHello)
		p2pAdjTLV := h.GetP2PAdjTLV()
		if p2pAdjTLV != nil {
			if p2pAdjTLV.Length() == packet.P2PAdjacencyStateTLVLenWithoutNeighbor {
				return newFSMInitializingState(s.fsm), "Received P2P Hello with Adjacency TLV"
			}

			if p2pAdjTLV.NeighborSystemID == s.fsm.neighbor.dev.srv.config.NETs[0].SystemID {
				//return newFSMUpState(s.fsm), "Received P2P Hello with Adjacency TLV including us"
			}

			return newFSMDownState(s.fsm), "Received P2P Hello with Adjacency TLV including someone else"
		}
	}
}
