package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/log"
)

func (nifa *netIfa) receiver() {
	for {
		select {
		case <-nifa.done:
			return
		default:
			err := nifa.receive()
			if err != nil {
				log.WithFields(nifa.fields()).WithError(err).Error("Error")
			}
		}
	}
}

func (nifa *netIfa) receive() error {
	pkt, src, err := nifa.ethHandler.RecvPacket()
	if err != nil {
		return fmt.Errorf("Read failed: %w", err)
	}

	return nifa.processPkt(src, pkt)
}

func (nifa *netIfa) processPkt(src ethernet.MACAddr, rawPkt []byte) error {
	buf := bytes.NewBuffer(rawPkt)
	pkt, err := packet.Decode(buf)
	if err != nil {
		return fmt.Errorf("Decode failed: %w", err)
	}

	err = nifa.validatePkt(src, pkt)
	if err != nil {
		log.WithFields(nifa.fields()).WithError(err).Debug("Packet validation failed")
		return nil
	}

	switch pkt.Header.PDUType {
	case packet.P2P_HELLO:
		return nifa.processP2PHello(src, pkt.Body.(*packet.P2PHello))
	case packet.L2_LS_PDU_TYPE:
		nifa.srv.lsdbL2.processLSP(nifa, pkt.Body.(*packet.LSPDU))
		return nil
	case packet.L2_CSNP_TYPE:
		nifa.srv.lsdbL2.processCSNP(nifa, pkt.Body.(*packet.CSNP))
		return nil
	case packet.L2_PSNP_TYPE:
		nifa.srv.lsdbL2.processPSNP(nifa, pkt.Body.(*packet.PSNP))
		return nil
	}

	return fmt.Errorf("Unknown PDU type %d", pkt.Header.PDUType)
}

func (nifa *netIfa) validatePkt(src ethernet.MACAddr, pkt *packet.ISISPacket) error {
	if pkt.Header.PDUType == packet.L2_LS_PDU_TYPE || pkt.Header.PDUType == packet.L2_CSNP_TYPE || pkt.Header.PDUType == packet.L2_PSNP_TYPE {
		if nifa.neighborManagerL2 == nil {
			return fmt.Errorf("Received L2 PDU on L2 disabled interface")
		}

		if !nifa.neighborManagerL2.neighborUp(src) {
			return fmt.Errorf("Received L2 PDU without neighbor up (src=%v)", src)
		}
	}

	if pkt.Header.PDUType == packet.L1_LS_PDU_TYPE || pkt.Header.PDUType == packet.L1_CSNP_TYPE || pkt.Header.PDUType == packet.L1_PSNP_TYPE {
		if nifa.neighborManagerL1 == nil {
			return fmt.Errorf("Received L1 PDU on L1 disabled interface")
		}

		if !nifa.neighborManagerL1.neighborUp(src) {
			return fmt.Errorf("Received L1 PDU without neighbor up (src=%v)", src)
		}
	}

	return nil
}

func (nifa *netIfa) processP2PHello(src ethernet.MACAddr, hello *packet.P2PHello) error {
	if hello.CircuitType == types.CircuitTypeL1 || hello.CircuitType == types.CircuitTypeL1L2 {
		if nifa.neighborManagerL1 != nil {
			err := nifa.neighborManagerL1.processP2PHello(src, hello)
			if err != nil {
				return fmt.Errorf("neighbor manager L1 failed processing the p2p hello: %w", err)
			}
		}
	}

	if hello.CircuitType == types.CircuitTypeL2 || hello.CircuitType == types.CircuitTypeL1L2 {
		if nifa.neighborManagerL2 != nil {
			err := nifa.neighborManagerL2.processP2PHello(src, hello)
			if err != nil {
				return fmt.Errorf("neighbor manager L2 failed processing the p2p hello: %w", err)
			}
		}
	}

	return nil
}
