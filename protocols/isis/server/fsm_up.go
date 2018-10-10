package server

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"

	log "github.com/sirupsen/logrus"
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
		select {
		case <-s.fsm.holdTimer.C:
			return s.holdTimerExpired()
		case pkt := <-s.fsm.pktCh:
			switch pkt.Header.PDUType {
			case packet.L2_LS_PDU_TYPE:
				fmt.Printf("LSPDU Received\n")
				lspdu := pkt.Body.(*packet.LSPDU)
				return s.processL2LSPDU(s.fsm.neighbor.ifa, lspdu)
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
}

func (s upState) holdTimerExpired() (state, string) {
	return newDownState(s.fsm), "Hold timer expired"
}

func (s *upState) sender() {
	t := time.NewTicker(time.Millisecond * 100)
	for {
		<-t.C
		lspdus, psnpEntries := s.fsm.neighbor.ifa.isisServer.lsdb.scanSRMSSN(s.fsm.neighbor.ifa)

		for _, lspdu := range lspdus {
			lspBuf := bytes.NewBuffer(nil)
			lspdu.Serialize(lspBuf)

			hdrBuf := bytes.NewBuffer(nil)
			hdr := &packet.ISISHeader{
				ProtoDiscriminator:  0x83,
				LengthIndicator:     20,
				ProtocolIDExtension: 1,
				IDLength:            0,
				PDUType:             packet.L2_LS_PDU_TYPE,
				Version:             1,
				MaxAreaAddresses:    0,
			}
			hdr.Serialize(hdrBuf)

			hdrBuf.Write(lspBuf.Bytes())
			fmt.Printf("Sending LSP: %v\n", hdrBuf.Bytes())

			err := s.fsm.neighbor.ifa.sendPacket(hdrBuf.Bytes(), AllP2PISS)
			if err != nil {
				log.Fatalf("failed to send packet: %v", err)
			}
		}

		psnps := packet.NewPSNPs(s.fsm.neighbor.ifa.isisServer.systemID(), psnpEntries, 1492)
		for _, psnp := range psnps {
			psnpBuf := bytes.NewBuffer(nil)
			psnp.Serialize(psnpBuf)

			hdrBuf := bytes.NewBuffer(nil)
			hdr := &packet.ISISHeader{
				ProtoDiscriminator:  0x83,
				LengthIndicator:     20,
				ProtocolIDExtension: 1,
				IDLength:            0,
				PDUType:             packet.L2_PSNP_TYPE,
				Version:             1,
				MaxAreaAddresses:    0,
			}
			hdr.Serialize(hdrBuf)

			hdrBuf.Write(psnpBuf.Bytes())
			fmt.Printf("Sending PSNP: %v\n", hdrBuf.Bytes())

			err := s.fsm.neighbor.ifa.sendPacket(hdrBuf.Bytes(), AllP2PISS)
			if err != nil {
				log.Fatalf("failed to send packet: %v", err)
			}
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
		if !s.fsm.holdTimer.Stop() {
			<-s.fsm.holdTimer.C
		}
		s.fsm.holdTimer.Reset(time.Second * time.Duration(s.fsm.neighbor.holdingTime))
		return newUpState(s.fsm), "Found myself in P2PAdjacencyTLV"
	}

	return newInitializingState(s.fsm), "Received valid P2PHello"
}

func (s upState) getState() uint8 {
	return packet.UP_STATE
}
