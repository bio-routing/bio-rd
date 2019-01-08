package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

// IPInterfaceAddressesTLVType is the type value of an IP interface address TLV
const IPInterfaceAddressesTLVType = 132

// IPInterfaceAddressesTLV represents an IP interface TLV
type IPInterfaceAddressesTLV struct {
	TLVType       uint8
	TLVLength     uint8
	IPv4Addresses []uint32
}

// NewIPInterfaceAddressesTLV creates a new IPInterfaceAddressesTLV
func NewIPInterfaceAddressesTLV(addrs []uint32) *IPInterfaceAddressesTLV {
	return &IPInterfaceAddressesTLV{
		TLVType:       IPInterfaceAddressesTLVType,
		TLVLength:     4,
		IPv4Addresses: addrs,
	}
}

func readIPInterfaceAddressesTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*IPInterfaceAddressesTLV, error) {
	pdu := &IPInterfaceAddressesTLV{
		TLVType:       tlvType,
		TLVLength:     tlvLength,
		IPv4Addresses: make([]uint32, tlvLength/4),
	}

	fields := make([]interface{}, len(pdu.IPv4Addresses))
	for i := range pdu.IPv4Addresses {
		fields[i] = &pdu.IPv4Addresses[i]
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, nil
}

// Type returns the type of the TLV
func (i IPInterfaceAddressesTLV) Type() uint8 {
	return i.TLVType
}

// Length returns the length of the TLV
func (i IPInterfaceAddressesTLV) Length() uint8 {
	return i.TLVLength
}

// Value gets the TLV itself
func (i IPInterfaceAddressesTLV) Value() interface{} {
	return i
}

// Serialize serializes an IP interfaces address TLV
func (i IPInterfaceAddressesTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	for j := range i.IPv4Addresses {
		buf.Write(convert.Uint32Byte(i.IPv4Addresses[j]))
	}
}
