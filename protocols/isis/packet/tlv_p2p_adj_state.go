package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
)

const (
	// P2PAdjacencyStateTLVType is the type value of an P2P adjacency state TLV
	P2PAdjacencyStateTLVType = 240
	// withoutNeighbor is the length of this TLV without a neighbor
	withoutNeighbor = 5
	//withNeighbor is the length of this TLV including a neighbor
	withNeighbor = 15
)

// P2PAdjacencyStateTLV represents an P2P adjacency state TLV
type P2PAdjacencyStateTLV struct {
	TLVType                        uint8
	TLVLength                      uint8
	AdjacencyState                 uint8
	ExtendedLocalCircuitID         uint32
	NeighborSystemID               types.SystemID
	NeighborExtendedLocalCircuitID uint32
}

func readP2PAdjacencyStateTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*P2PAdjacencyStateTLV, error) {
	pdu := &P2PAdjacencyStateTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	fields := make([]interface{}, 0)
	switch pdu.TLVLength {
	case withoutNeighbor:
		fields = []interface{}{
			&pdu.AdjacencyState,
			&pdu.ExtendedLocalCircuitID,
		}
	case withNeighbor:
		fields = []interface{}{
			&pdu.AdjacencyState,
			&pdu.ExtendedLocalCircuitID,
			&pdu.NeighborSystemID,
			&pdu.NeighborExtendedLocalCircuitID,
		}
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, nil
}

// NewP2PAdjacencyStateTLV creates a new P2PAdjacencyStateTLV
func NewP2PAdjacencyStateTLV(adjacencyState uint8, extendedLocalCircuitID uint32) *P2PAdjacencyStateTLV {
	return &P2PAdjacencyStateTLV{
		TLVType:                P2PAdjacencyStateTLVType,
		TLVLength:              withoutNeighbor,
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
	buf.WriteByte(p.AdjacencyState)
	buf.Write(convert.Uint32Byte(p.ExtendedLocalCircuitID))

	if p.TLVLength == withNeighbor {
		buf.Write(p.NeighborSystemID[:])
		buf.Write(convert.Uint32Byte(p.NeighborExtendedLocalCircuitID))
	}
}
