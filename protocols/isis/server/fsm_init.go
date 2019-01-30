package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	log "github.com/sirupsen/logrus"
)

type fsmInitializingState struct {
	fsm *fsm
}

func newFSMInitializingState(fsm *fsm) *fsmInitializingState {
	return &fsmInitializingState{
		fsm: fsm,
	}
}

func (s fsmInitializingState) run() (state, string) {
	for {
		pkt := <-s.fsm.pktCh
		if pkt.Header.PDUType != packet.P2P_HELLO {
			log.Warningf("Received unexpected packet type %v in initializing state", pkt.Header.PDUType)
			continue
		}

		if !s.fsm.neighbor.dev.p2p {
			return newFSMDownState(s.fsm), "Received P2PHello on non-P2P interface"
		}

		h := pkt.Body.(*packet.P2PHello)

		a := h.GetAreaAddressesTLV()
		if a == nil {
			return newFSMDownState(s.fsm), "Received P2PHello without area addresses TLV"
		}

		p := h.GetProtocolsSupportedTLV()
		if p == nil {
			return newFSMDownState(s.fsm), "Received P2PHello without protocols supported TLV"
		}

		/*if !s.fsm.neighbor.compareSupportedProtocols(p.NetworkLayerProtocolIDs) {
			return newFSMDownState(s.fsm), "Supported protocols mismatch"
		}*/

		p2pAdjTLV := h.GetP2PAdjTLV()
		if p2pAdjTLV == nil {
			return newFSMDownState(s.fsm), "Received P2PHello without P2P Adjacency TLV"
		}

		if p2pAdjTLV.Length() == packet.P2PAdjacencyStateTLVLenWithoutNeighbor {
			fmt.Printf("DEBUG: P2P Adjacency TLV has no neighbor\n")
			continue
		}

		fmt.Printf("p2pAdjTLV.NeighborSystemID: %s\n", p2pAdjTLV.NeighborSystemID.String())
		fmt.Printf("Local:                      %s\n", s.fsm.neighbor.dev.srv.config.NETs[0].SystemID)
		if p2pAdjTLV.NeighborSystemID == s.fsm.neighbor.dev.srv.config.NETs[0].SystemID {
			//return newFSMUpState(s.fsm), "Received P2P Hello with Adjacency TLV including us"
		}

		panic("foo")
	}
}

func (s fsmInitializingState) getState() uint8 {
	return packet.INITIALIZING_STATE
}
