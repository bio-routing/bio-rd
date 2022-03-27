package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
)

// ISNeighborsTLVType is the type value of an IS Neighbor TLV
const ISNeighborsTLVType = 6

// ISNeighborsTLV represents an IS Neighbor TLV
type ISNeighborsTLV struct {
	TLVType      uint8
	TLVLength    uint8
	NeighborSNPA types.SystemID
}

// ISNeighborsTLVLength is the length of an IS Neighbor TLV
const ISNeighborsTLVLength = 8

func readISNeighborsTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ISNeighborsTLV, error) {
	pdu := &ISNeighborsTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []interface{}{
		&pdu.NeighborSNPA,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	return pdu, nil
}

// Type returns the type of the TLV
func (i ISNeighborsTLV) Type() uint8 {
	return i.TLVType
}

// Length returns the length of the TLV
func (i ISNeighborsTLV) Length() uint8 {
	return i.TLVLength
}

// Value returns the TLV itself
func (i ISNeighborsTLV) Value() interface{} {
	return i
}

// Serialize serializes an WriteByte into a buffer
func (i ISNeighborsTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	buf.Write(i.NeighborSNPA[:])
}
