package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"
)

// IPInterfaceAddressTLVType is the type value of an IP interface address TLV
const IPInterfaceAddressTLVType = 132

// IPInterfaceAddressTLV represents an IP interface TLV
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

// Type returns the type of the TLV
func (i IPInterfaceAddressTLV) Type() uint8 {
	return i.TLVType
}

// Length returns the length of the TLV
func (i IPInterfaceAddressTLV) Length() uint8 {
	return i.TLVLength
}

// Serialize serializes an IP interfaces address TLV
func (i IPInterfaceAddressTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	buf.Write(convert.Uint32Byte(i.IPv4Address))
}
