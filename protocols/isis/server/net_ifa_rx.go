package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
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
	pkt, src, err := nifa.ethHandler.RecvPacket()
	if err != nil {
		return errors.Wrap(err, "Read failed")
	}

	return nifa.processPkt(src, pkt)
}

func (nifa *netIfa) processPkt(src types.MACAddress, rawPkt []byte) error {
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
	case packet.L2_LS_PDU_TYPE:
		log.WithFields(nifa.fields()).Debug("Received L2 LSPDU")

		nifa.processL2LSPDU(pkt.Body.(*packet.LSPDU))
		return nil
	case packet.L2_CSNP_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 CSNP")

		nifa.processL2CSNPDU(pkt.Body.(*packet.CSNP))
		return nil
	case packet.L2_PSNP_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 PSNP")

		return nil
	}

	return fmt.Errorf("Unknown PDU type %d", pkt.Header.PDUType)
}

func (nifa *netIfa) validatePkt(src types.MACAddress, pkt *packet.ISISPacket) error {
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

func (nifa *netIfa) processP2PHello(src types.MACAddress, hello *packet.P2PHello) error {
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

func (nifa *netIfa) processL2CSNPDU(pkt *packet.CSNP) {
	nifa.srv.lsdbL2.processCSNP(pkt, nifa)
}

func (nifa *netIfa) processL2LSPDU(pkt *packet.LSPDU) error {
	nifa.srv.lsdbL2.lspsMu.Lock()
	defer nifa.srv.lsdbL2.lspsMu.Unlock()

	if nifa.srv.lsdbL2._exists(pkt) {
		return nifa._processL2LSPDULSPExists(pkt)
	}

	return nifa._processL2LSPDUNewOrNewerLSP(pkt)
}

func (nifa *netIfa) _processL2LSPDULSPExists(pkt *packet.LSPDU) error {
	if nifa.srv.lsdbL2._isNewer(pkt) {
		return nifa._processL2LSPDUNewOrNewerLSP(pkt)
	}

	return nifa._processL2LSPDULSPIsOlder(pkt)
}

func (nifa *netIfa) _processL2LSPDULSPIsOlder(pkt *packet.LSPDU) error {
	nifa.srv.lsdbL2.lsps[pkt.LSPID].setSRM(nifa)
	nifa.srv.lsdbL2.lsps[pkt.LSPID].clearSSNFlag(nifa)
	return nil
}

func (nifa *netIfa) _processL2LSPDUNewOrNewerLSP(pkt *packet.LSPDU) error {
	nifa.srv.lsdbL2.lsps[pkt.LSPID].lspdu = pkt

	for _, ifa := range nifa.srv.netIfaManager.getAllInterfacesExcept(nifa) {
		nifa.srv.lsdbL2.lsps[pkt.LSPID].setSRM(ifa)
	}

	nifa.srv.lsdbL2.lsps[pkt.LSPID].setSSN(nifa)
	return nil
}
