package packet

import (
	"bytes"
	"fmt"
)

const (
	P2P_HELLO = 0x11
)

type isisPacket struct {
	header *isisHeader
	body   interface{}
}

func Decode(buf *bytes.Buffer) (*isisPacket, error) {
	pkt := &isisPacket{}

	hdr, err := decodeHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode header: %v", err)
	}
	pkt.header = hdr

	switch pkt.header.PDUType {
	case P2P_HELLO:
		p2pHello, err := decodeISISP2PHello(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode P2P hello: %v", err)
		}
		pkt.body = p2pHello
	}

	return pkt, nil
}
