package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

const NodeMaximumSIDDepthTLVType = 23
const NodeMaximumSIDDepthTLVLen = 0

type NodeMaximumSIDDepthTLV struct {
	TLVType   uint8
	TLVLength uint8
	MSDs      []MSD
}

type MSD struct {
	Type  uint8
	Value uint8
}

func readNodeMaximumSIDDepthTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*NodeMaximumSIDDepthTLV, error) {
	pdu := &NodeMaximumSIDDepthTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	for i := 0; i < int(tlvLength/2); i++ {
		msd := MSD{}
		fields := []any{
			&msd.Type,
			&msd.Value,
		}

		err := decode.Decode(buf, fields)
		if err != nil {
			return nil, fmt.Errorf("unable to decode fields: %v", err)
		}
		pdu.MSDs = append(pdu.MSDs, msd)
	}

	return pdu, nil
}

func (n NodeMaximumSIDDepthTLV) Copy() TLV {
	ret := n
	ret.MSDs = make([]MSD, len(n.MSDs))
	copy(ret.MSDs, n.MSDs)
	return &ret
}

// Type gets the type of the TLV
func (n NodeMaximumSIDDepthTLV) Type() uint8 {
	return n.TLVType
}

// Length gets the length of the TLV
func (n NodeMaximumSIDDepthTLV) Length() uint8 {
	return n.TLVLength
}

// Value returns the TLV itself
func (n *NodeMaximumSIDDepthTLV) Value() any {
	return n
}

// Serialize serializes an NodeMaximumSIDDepthTLV
func (n NodeMaximumSIDDepthTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(n.TLVType)
	buf.WriteByte(n.TLVLength)
	for _, msd := range n.MSDs {
		buf.WriteByte(msd.Type)
		buf.WriteByte(msd.Value)
	}
}
