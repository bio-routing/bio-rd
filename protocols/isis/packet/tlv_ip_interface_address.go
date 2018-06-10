package packet

import (
	"bytes"
	"fmt"
)

const IPInterfaceAddressTLVType = 132

type IPInterfaceAddressTLV struct {
	TLVType     uint8
	TLVLength   uint8
	IPv4Address uint32
}

const ipv4AddressLength = 4

func readIPInterfaceAddressTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*IPInterfaceAddressTLV, uint8, error) {
	pdu := &IPInterfaceAddressTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	fields := []interface{}{
		&pdu.IPv4Address,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, ipv4AddressLength, nil
}

func (i IPInterfaceAddressTLV) Type() uint8 {
	return i.TLVType
}

func (i IPInterfaceAddressTLV) Length() uint8 {
	return i.TLVLength
}
