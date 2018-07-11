package packet

import (
	"bytes"
	"fmt"
)

// ProtocolsSupportedTLVType is the type value of an protocols supported TLV
const ProtocolsSupportedTLVType = 129

// ProtocolsSupportedTLV represents a protocols supported TLV
type ProtocolsSupportedTLV struct {
	TLVType                 uint8
	TLVLength               uint8
	NerworkLayerProtocolIDs []uint8
}

func readProtocolsSupportedTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ProtocolsSupportedTLV, uint8, error) {
	pdu := &ProtocolsSupportedTLV{
		TLVType:                 tlvType,
		TLVLength:               tlvLength,
		NerworkLayerProtocolIDs: make([]uint8, tlvLength),
	}

	protoID := uint8(0)
	fields := []interface{}{
		&protoID,
	}

	read := uint8(2)
	for i := uint8(0); i < tlvLength; i++ {
		err := decode(buf, fields)
		if err != nil {
			return nil, 0, fmt.Errorf("Unable to decode fields: %v", err)
		}
		pdu.NerworkLayerProtocolIDs[i] = protoID
		read++
	}

	return pdu, read, nil
}

// Type gets the type of the TLV
func (p ProtocolsSupportedTLV) Type() uint8 {
	return p.TLVType
}

// Length gets the length of the TLV
func (p ProtocolsSupportedTLV) Length() uint8 {
	return p.TLVLength
}

// Serialize serializes a protocols supported TLV
func (p ProtocolsSupportedTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(p.TLVType)
	buf.WriteByte(p.TLVLength)
	buf.Write(p.NerworkLayerProtocolIDs)
}
