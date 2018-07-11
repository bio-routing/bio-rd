package packet

import (
	"bytes"
	"fmt"
)

// AreaAddressTLVType is the type value of an area address TLV
const AreaAddressTLVType = 1

// AreaAddressTLV represents an area address TLV
type AreaAddressTLV struct {
	TLVType   uint8
	TLVLength uint8
	AreaID    [6]byte
}

func readAreaAddressTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*AreaAddressTLV, uint8, error) {
	pdu := &AreaAddressTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	n, err := buf.Read(pdu.AreaID[:])
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to read: %v", err)
	}

	return pdu, uint8(n + 1), nil
}

// Type gets the type of the TLV
func (i AreaAddressTLV) Type() uint8 {
	return i.TLVType
}

// Length gets the length of the TLV
func (i AreaAddressTLV) Length() uint8 {
	return i.TLVLength
}

// Serialize serializes an area address TLV
func (i AreaAddressTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	buf.Write(i.AreaID[:])
}
