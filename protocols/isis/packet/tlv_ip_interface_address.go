package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
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

// NewIPInterfaceAddressTLV creates a new IPInterfaceAddressTLV
func NewIPInterfaceAddressTLV(addr uint32) *IPInterfaceAddressTLV {
	return &IPInterfaceAddressTLV{
		TLVType:     IPInterfaceAddressTLVType,
		TLVLength:   4,
		IPv4Address: addr,
	}
}

func readIPInterfaceAddressTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*IPInterfaceAddressTLV, error) {
	pdu := &IPInterfaceAddressTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	fields := []interface{}{
		&pdu.IPv4Address,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode fields")
	}

	return pdu, nil
}

// Type returns the type of the TLV
func (i IPInterfaceAddressTLV) Type() uint8 {
	return i.TLVType
}

// Length returns the length of the TLV
func (i IPInterfaceAddressTLV) Length() uint8 {
	return i.TLVLength
}

// Value gets the TLV itself
func (i IPInterfaceAddressTLV) Value() interface{} {
	return i
}

// Serialize serializes an IP interfaces address TLV
func (i IPInterfaceAddressTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	buf.Write(convert.Uint32Byte(i.IPv4Address))
}
