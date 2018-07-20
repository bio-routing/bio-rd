package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

// AreaAddressTLVType is the type value of an area address TLV
const AreaAddressTLVType = 1

// AreaAddressTLV represents an area address TLV
type AreaAddressTLV struct {
	TLVType   uint8
	TLVLength uint8
	AreaIDs   []types.AreaID
}

func readAreaAddressTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*AreaAddressTLV, uint8, error) {
	count := tlvLength / 6
	pdu := &AreaAddressTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
		AreaIDs:   make([]types.AreaID, count),
	}

	n := 0
	for i := uint8(0); i < count; i++ {
		nread, err := buf.Read(pdu.AreaIDs[i][:])
		if err != nil {
			return nil, 0, fmt.Errorf("Unable to read: %v", err)
		}
		n += nread
	}

	return pdu, uint8(n + 1), nil
}

// Type gets the type of the TLV
func (a AreaAddressTLV) Type() uint8 {
	return a.TLVType
}

// Length gets the length of the TLV
func (a AreaAddressTLV) Length() uint8 {
	return a.TLVLength
}

// Serialize serializes an area address TLV
func (a AreaAddressTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(a.TLVType)
	buf.WriteByte(a.TLVLength)

	count := a.TLVLength / 6
	for i := uint8(0); i < count; i++ {
		buf.Write(a.AreaIDs[i][:])
	}
}
