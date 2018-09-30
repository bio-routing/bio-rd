package packet

import(
	"bytes"
)

const PaddingType = 8

type PaddingTLV struct {
	TLVType     uint8
	TLVLength   uint8
	PaddingData []byte
}

func NewPaddingTLV(length uint8) *PaddingTLV {
	return &PaddingTLV{
		TLVType: PaddingType,
		TLVLength: length,
		PaddingData: make([]byte, length),
	}
}

func (p *PaddingTLV) Type() uint8 {
	return p.TLVType
}

func (p *PaddingTLV) Length() uint8 {
	return p.TLVLength
}

func (p *PaddingTLV) Value() interface{} {
	return p.PaddingData
}

func (p *PaddingTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(p.TLVType)
	buf.WriteByte(p.TLVLength)
	buf.Write(p.PaddingData)
}