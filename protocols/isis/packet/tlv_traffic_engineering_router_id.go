package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// TrafficEngineeringRouterIDTLVType is the type value of an Traffic Engineering Router ID TLV
	TrafficEngineeringRouterIDTLVType = 134
)

// TrafficEngineeringRouterIDTLV is a Traffic Engineering Router ID TLV
type TrafficEngineeringRouterIDTLV struct {
	TLVType   uint8
	TLVLength uint8
	Address   uint32
}

// NewTrafficEngineeringRouterIDTLV creates a new TrafficEngineeringRouterIDTLV
func NewTrafficEngineeringRouterIDTLV(addr uint32) *TrafficEngineeringRouterIDTLV {
	return &TrafficEngineeringRouterIDTLV{
		TLVType:   TrafficEngineeringRouterIDTLVType,
		TLVLength: 4,
		Address:   addr,
	}
}

// Serialize serializes a TrafficEngineeringRouterIDTLV
func (t *TrafficEngineeringRouterIDTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(t.TLVType)
	buf.WriteByte(t.TLVLength)
	buf.Write(convert.Uint32Byte(t.Address))
}

func readTrafficEngineeringRouterIDTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*TrafficEngineeringRouterIDTLV, error) {
	pdu := &TrafficEngineeringRouterIDTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	fields := []interface{}{
		&pdu.Address,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	return pdu, nil
}

func (t TrafficEngineeringRouterIDTLV) Copy() TLV {
	x := t
	return &x
}

// Type gets the type of the TLV
func (t TrafficEngineeringRouterIDTLV) Type() uint8 {
	return t.TLVType
}

// Length gets the length of the TLV
func (t TrafficEngineeringRouterIDTLV) Length() uint8 {
	return t.TLVLength
}

// Value gets the TLV itself
func (t TrafficEngineeringRouterIDTLV) Value() interface{} {
	return t
}
