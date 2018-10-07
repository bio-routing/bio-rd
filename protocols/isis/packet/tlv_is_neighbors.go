package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

// ISNeighborsTLVType is the type value of an IS Neighbor TLV
const ISNeighborsTLVType = 6

// ISNeighborsTLV represents an IS Neighbor TLV
type ISNeighborsTLV struct {
	TLVType      uint8
	TLVLength    uint8
	NeighborSNPA [6]byte
}

// ISNeighborsTLVLength is the length of an IS Neighbor TLV
const ISNeighborsTLVLength = 8

func readISNeighborsTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ISNeighborsTLV, uint8, error) {
	pdu := &ISNeighborsTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []interface{}{
		&pdu.NeighborSNPA,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, ISNeighborsTLVLength, nil
}

// Type returns the type of the TLV
func (i ISNeighborsTLV) Type() uint8 {
	return i.TLVType
}

// Length returns the length of the TLV
func (i ISNeighborsTLV) Length() uint8 {
	return i.TLVLength
}

func (i ISNeighborsTLV) Value() interface{} {
	return i.NeighborSNPA
}

// Serialize serializes an WriteByte into a buffer
func (i ISNeighborsTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	buf.Write(i.NeighborSNPA[:])
}
