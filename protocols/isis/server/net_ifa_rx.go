package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (nifa *netIfa) receiver() {
	for {
		err := nifa.receive()
		if err != nil {
			log.WithFields(nifa.fields()).Errorf("Error: %v", err)
		}
	}
}

func (nifa *netIfa) receive() error {
	pkt, _, err := nifa.ethHandler.RecvPacket()
	if err != nil {
		return errors.Wrap(err, "Read failed")
	}

	return nifa.processPkt(pkt)
}

func (nifa *netIfa) processPkt(rawPkt []byte) error {
	buf := bytes.NewBuffer(rawPkt)
	pkt, err := packet.Decode(buf)
	if err != nil {
		return errors.Wrap(err, "Decode failed")
	}

	switch pkt.Header.PDUType {
	case packet.P2P_HELLO:
		return nifa.processP2PHello(pkt)
	case packet.L2_LS_PDU_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 LSPDU")

		return nil
	case packet.L2_CSNP_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 CSNP")

		return nil
	case packet.L2_PSNP_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 PSNP")

		return nil
	}

	return fmt.Errorf("Unknown PDU type %d", pkt.Header.PDUType)
}

func (nifa *netIfa) processP2PHello(pkt *packet.ISISPacket) error {
	hello := pkt.Body.(*packet.P2PHello)

	if hello.CircuitType == 1 || hello.CircuitType == 3 {
		if nifa.neighborManagerL1 != nil {
			err := nifa.neighborManagerL1.processP2PHello(hello)
			if err != nil {
				return errors.Wrap(err, "neighbor manager L1 failed processing the p2p hello")
			}
		}
	}

	if hello.CircuitType == 2 || hello.CircuitType == 3 {
		if nifa.neighborManagerL2 != nil {
			err := nifa.neighborManagerL2.processP2PHello(hello)
			if err != nil {
				return errors.Wrap(err, "neighbor manager L2 failed processing the p2p hello")
			}
		}
	}

	return nil
}
