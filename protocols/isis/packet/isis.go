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
)

type ISISPacket struct {
	Header *ISISHeader
	Body   interface{}
}

func Decode(buf *bytes.Buffer) (*ISISPacket, error) {
	pkt := &ISISPacket{}

	hdr, err := decodeHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode header: %v", err)
	}
	pkt.Header = hdr

	switch pkt.Header.PDUType {
	case P2P_HELLO:
		p2pHello, err := decodeISISP2PHello(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode P2P hello: %v", err)
		}
		pkt.Body = p2pHello
	}

	return pkt, nil
}
