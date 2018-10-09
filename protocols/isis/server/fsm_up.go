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
			lspdu := pkt.Body.(packet.LSPDU)
			ack := packet.CSNP{}
		case packet.L2_CSNP_TYPE:
			fmt.Printf("CSNP received\n")
			return s.processCSNP(pkt.Body.(*packet.CSNP))
		case packet.L2_PSNP_TYPE:
			fmt.Printf("PSNP received\n")
			return s.processPSNP(pkt.Body.(*packet.PSNP))
		case packet.P2P_HELLO:
			return s.processP2PHello(pkt.Body.(*packet.P2PHello))
		default:
			return newDownState(s.fsm), "Received unexpected packet"
		}
	}
}

func (s upState) processCSNP(csnp *packet.CSNP) (state, string) {
	s.fsm.neighbor.ifa.isisServer.lsdb.processCSNP(s.fsm.neighbor.ifa, csnp)
	return newUpState(s.fsm), "Received CSNP"
}

func (s upState) processPSNP(psnp *packet.PSNP) (state, string) {
	s.fsm.neighbor.ifa.isisServer.lsdb.processPSNP(s.fsm.neighbor.ifa, psnp)
	return newUpState(s.fsm), "Received PSNP"
}

func (s upState) processL2LSPDU(ifa *netIf, pdu *packet.LSPDU) (state, string) {
	s.fsm.neighbor.ifa.isisServer.lsdb.processLSPDU(ifa, pdu)

	return newUpState(s.fsm), "Found myself in P2PAdjacencyTLV"
}

func (s upState) processP2PHello(hello *packet.P2PHello) (state, string) {
	if !s.fsm.neighbor.ifa.p2p {
		return newDownState(s.fsm), "Received P2PHello on non-P2P interface"
	}

	p2pAdjStateTLVFound := false
	protocolsSupportedTLVfound := false
	areaAddressesTLVFound := false
	foundSelf := false

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
		return newUpState(s.fsm), "Found myself in P2PAdjacencyTLV"
	}

	return newInitializingState(s.fsm), "Received valid P2PHello"
}

func (s upState) getState() uint8 {
	return packet.UP_STATE
}
