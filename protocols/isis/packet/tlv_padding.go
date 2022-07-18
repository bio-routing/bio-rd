package packet

import (
	"bytes"
)

// PaddingType is the type value of a padding TLV
const PaddingType = 8

// PaddingTLV represents a padding TLV
type PaddingTLV struct {
	TLVType     uint8
	TLVLength   uint8
	PaddingData []byte
}

// NewPaddingTLV creates a new padding TLV
func NewPaddingTLV(length uint8) *PaddingTLV {
	return &PaddingTLV{
		TLVType:     PaddingType,
		TLVLength:   length,
		PaddingData: make([]byte, length),
	}
}

func (p *PaddingTLV) Copy() TLV {
	ret := *p
	ret.PaddingData = make([]byte, 0, len(p.PaddingData))
	copy(ret.PaddingData, p.PaddingData)
	return &ret
}

// Type gets the type of the TLV
func (p *PaddingTLV) Type() uint8 {
	return p.TLVType
}

// Length gets the length of the TLV
func (p *PaddingTLV) Length() uint8 {
	return p.TLVLength
}

// Value gets the TLV itself
func (p *PaddingTLV) Value() interface{} {
	return p
}

// Serialize serializes a padding TLV
func (p *PaddingTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(p.TLVType)
	buf.WriteByte(p.TLVLength)
	buf.Write(p.PaddingData)
}
