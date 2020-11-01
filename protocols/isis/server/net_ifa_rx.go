package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
		return errors.Wrap(err, "Read failed")
	}

	return nifa.processPkt(src, pkt)
}

func (nifa *netIfa) processPkt(src ethernet.MACAddr, rawPkt []byte) error {
	buf := bytes.NewBuffer(rawPkt)
	pkt, err := packet.Decode(buf)
	if err != nil {
		return errors.Wrap(err, "Decode failed")
	}

	err = nifa.validatePkt(src, pkt)
	if err != nil {
		log.WithFields(nifa.fields()).WithError(err).Debug("Packet validation failed")
		return nil
	}

	switch pkt.Header.PDUType {
	case packet.P2P_HELLO:
		return nifa.processP2PHello(src, pkt.Body.(*packet.P2PHello))
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
	if hello.CircuitType == 1 || hello.CircuitType == 3 {
		if nifa.neighborManagerL1 != nil {
			err := nifa.neighborManagerL1.processP2PHello(src, hello)
			if err != nil {
				return errors.Wrap(err, "neighbor manager L1 failed processing the p2p hello")
			}
		}
	}

	if hello.CircuitType == 2 || hello.CircuitType == 3 {
		if nifa.neighborManagerL2 != nil {
			err := nifa.neighborManagerL2.processP2PHello(src, hello)
			if err != nil {
				return errors.Wrap(err, "neighbor manager L2 failed processing the p2p hello")
			}
		}
	}

	return nil
}
