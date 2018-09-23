package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

const P2PAdjacencyStateTLVType = 2

type P2PAdjacencyStateTLV struct {
	TLVType                        uint8
	TLVLength                      uint8
	AdjacencyState                 uint8
	ExtendedLocalCircuitID         uint32
	NeighborSystemID               types.SystemID
	NeighborExtendedLocalCircuitID uint32
}

func readP2PAdjacencyStateTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*P2PAdjacencyStateTLV, uint8, error) {
	pdu := &P2PAdjacencyStateTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	read := uint8(0)
	fields := make([]interface{}, 0)
	switch pdu.TLVLength {
	case 5:
		read = 5
		fields = []interface{}{
			&pdu.AdjacencyState,
			&pdu.ExtendedLocalCircuitID,
		}
	case 15:
		read = 15
		fields = []interface{}{
			&pdu.AdjacencyState,
			&pdu.ExtendedLocalCircuitID,
			&pdu.NeighborSystemID,
			&pdu.NeighborExtendedLocalCircuitID,
		}
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, read, nil
}

func NewP2PAdjacencyStateTLV(adjacencyState uint8, extendedLocalCircuitID uint32) P2PAdjacencyStateTLV {
	return P2PAdjacencyStateTLV{
		TLVType:                P2PAdjacencyStateTLVType,
		TLVLength:              5,
		AdjacencyState:         adjacencyState,
		ExtendedLocalCircuitID: extendedLocalCircuitID,
	}
}

// Type gets the type of the TLV
func (p P2PAdjacencyStateTLV) Type() uint8 {
	return p.TLVType
}

// Length gets the length of the TLV
func (p P2PAdjacencyStateTLV) Length() uint8 {
	return p.TLVLength
}

// Value gets the TLV itself
func (p P2PAdjacencyStateTLV) Value() interface{} {
	return p
}

// Serialize serializes a protocols supported TLV
func (p P2PAdjacencyStateTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(p.TLVType)
	buf.WriteByte(p.TLVLength)

}
