package server

import (
	"bytes"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

func (nifa *netIfa) sendLSPDU(lsp *packet.LSPDU, level int) error {
	if level == 1 {
		panic("L1 is not supported yet")
		return nil
	}

	return nifa.sendPDU(lsp, packet.L2_LS_PDU_TYPE)
}

func (nifa *netIfa) sendPSNP(psnp *packet.PSNP, level int) error {
	if level == 1 {
		panic("L1 is not supported yet")
		return nil
	}

	return nifa.sendPDU(psnp, packet.L2_PSNP_TYPE)
}

func (nifa *netIfa) sendCSNP(csnp *packet.CSNP, level int) error {
	if level == 1 {
		panic("L1 is not supported yet")
		return nil
	}

	return nifa.sendPDU(csnp, packet.L2_CSNP_TYPE)
}

func (nifa *netIfa) sendPDU(pkt packet.Serializable, pduType uint8) error {
	buf := bytes.NewBuffer(nil)
	pkt.Serialize(buf)

	hdr := getHeader(pduType)
	hdrBuf := bytes.NewBuffer(nil)
	hdr.Serialize(hdrBuf)
	hdrBuf.Write(buf.Bytes())

	_, err := nifa.isP2PHelloCon.Write(hdrBuf.Bytes())
	return err
}

func getHeader(pduType uint8) packet.ISISHeader {
	h := packet.ISISHeader{
		ProtoDiscriminator:  0x83,
		LengthIndicator:     lengthIndicatorByPDUType(pduType),
		ProtocolIDExtension: 1,
		IDLength:            0,
		PDUType:             pduType,
		Version:             1,
		MaxAreaAddresses:    0,
	}

	return h
}

func lengthIndicatorByPDUType(pduType uint8) uint8 {
	switch pduType {
	case packet.L2_CSNP_TYPE:
		return packet.CSNPMinLen
	case packet.L1_CSNP_TYPE:
		return packet.CSNPMinLen
	case packet.L2_PSNP_TYPE:
		return packet.PSNPMinLen
	case packet.L1_PSNP_TYPE:
		return packet.PSNPMinLen
	case packet.P2P_HELLO:
		return packet.P2PHelloMinLen
	case packet.L2_LS_PDU_TYPE:
		return packet.LSPDUMinLen
	case packet.L1_LS_PDU_TYPE:
		return packet.LSPDUMinLen
	}

	return 0
}
