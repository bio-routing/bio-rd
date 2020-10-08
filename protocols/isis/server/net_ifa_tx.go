package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
)

func (nifa *netIfa) sendLSPDU(lsp *packet.LSPDU) {
	lspBuf := bytes.NewBuffer(nil)
	lsp.Serialize(lspBuf)

	hdr := getHeader(packet.L2_LS_PDU_TYPE, uint8(len(lspBuf.Bytes())))

	hdrBuf := bytes.NewBuffer(nil)
	hdr.Serialize(hdrBuf)
	hdrBuf.Write(lspBuf.Bytes())

	// TODO: Send the PDU
	fmt.Printf("Sending PDU: %v\n", hdrBuf.Bytes())
}

func (nifa *netIfa) sendPSNP(psnp *packet.PSNP) {
	// TODO: Implement sending PSNPs
}

func (nifa *netIfa) sendCSNP(csnp *packet.CSNP) {
	// TODO: Implement sending CSNPs
}

func getHeader(pduType uint8, lengthIndicator uint8) packet.ISISHeader {
	return packet.ISISHeader{
		ProtoDiscriminator:  0x83,
		LengthIndicator:     lengthIndicator,
		ProtocolIDExtension: 1,
		IDLength:            0,
		PDUType:             pduType,
		Version:             1,
		MaxAreaAddresses:    0,
	}
}
