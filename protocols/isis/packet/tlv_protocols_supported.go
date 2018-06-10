package packet

import (
	"bytes"
	"fmt"
)

const ProtocolsSupportedTLVType = 129

type ProtocolsSupportedTLV struct {
	TLVType                uint8
	TLVLength              uint8
	NerworkLayerProtocolID []uint8
}

func readProtocolsSupportedTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ProtocolsSupportedTLV, uint8, error) {
	pdu := &ProtocolsSupportedTLV{
		TLVType:                tlvType,
		TLVLength:              tlvLength,
		NerworkLayerProtocolID: make([]uint8, tlvLength),
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
		pdu.NerworkLayerProtocolID[i] = protoID
		read++
	}

	return pdu, read, nil
}

func (i ProtocolsSupportedTLV) Type() uint8 {
	return i.TLVType
}

func (i ProtocolsSupportedTLV) Length() uint8 {
	return i.TLVLength
}
