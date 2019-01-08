package packet

import (
	"bytes"
	"fmt"
)

const (
	L1_LAN_HELLO_TYPE = 0x0f
	L1_LS_PDU_TYPE    = 0x18
	L1_CSNP_TYPE      = 0x24
	L1_PSNP_TYPE      = 0x26
	L2_LAN_HELLO_TYPE = 0x10
	L2_LS_PDU_TYPE    = 0x14
	L2_CSNP_TYPE      = 0x19
	L2_PSNP_TYPE      = 0x1b
	P2P_HELLO         = 0x11

	DOWN_STATE         = 2
	INITIALIZING_STATE = 1
	UP_STATE           = 0
)

// ISISPacket represents an ISIS packet
type ISISPacket struct {
	Header *ISISHeader
	Body   interface{}
}

// Decode decodes ISIS packets
func Decode(buf *bytes.Buffer) (*ISISPacket, error) {
	pkt := &ISISPacket{}

	hdr, err := DecodeHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode header: %v", err)
	}
	pkt.Header = hdr

	switch pkt.Header.PDUType {
	case P2P_HELLO:
		p2pHello, err := DecodeP2PHello(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode P2P hello: %v", err)
		}
		pkt.Body = p2pHello
	case L2_LS_PDU_TYPE:
		lspdu, err := DecodeLSPDU(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode LSPDU: %v", err)
		}
		pkt.Body = lspdu
	case L2_CSNP_TYPE:
		csnp, err := DecodeCSNP(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode CSNP: %v", err)
		}
		pkt.Body = csnp
	case L2_PSNP_TYPE:
		psnp, err := DecodeCSNP(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode PSNP: %v", err)
		}
		pkt.Body = psnp
	}



	return pkt, nil
}
