package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
)

const ChecksumTLVType = 12

// ChecksumTLV represents a checksum TLV
type ChecksumTLV struct {
	TLVType   uint8
	TLVLength uint8
	Checksum  uint16
}

// Type gets the type of the TLV
func (c *ChecksumTLV) Type() uint8 {
	return c.TLVType
}

// Length gets the length of the TLV
func (c *ChecksumTLV) Length() uint8 {
	return c.TLVLength
}

// Value returns the checksum
func (c *ChecksumTLV) Value() interface{} {
	return c.Checksum
}

func readChecksumTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ChecksumTLV, error) {
	pdu := &ChecksumTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	fields := []interface{}{
		&pdu.Checksum,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, nil
}

// Serialize serializes a checksum TLV
func (c *ChecksumTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(c.TLVType)
	buf.WriteByte(c.TLVLength)
	buf.Write(convert.Uint16Byte(c.Checksum))
}
