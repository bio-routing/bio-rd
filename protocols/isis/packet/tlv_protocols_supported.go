package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
)

// ProtocolsSupportedTLVType is the type value of an protocols supported TLV
const ProtocolsSupportedTLVType = 129

// ProtocolsSupportedTLV represents a protocols supported TLV
type ProtocolsSupportedTLV struct {
	TLVType                 uint8
	TLVLength               uint8
	NetworkLayerProtocolIDs []uint8
}

func readProtocolsSupportedTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ProtocolsSupportedTLV, error) {
	pdu := &ProtocolsSupportedTLV{
		TLVType:                 tlvType,
		TLVLength:               tlvLength,
		NetworkLayerProtocolIDs: make([]uint8, tlvLength),
	}

	protoID := uint8(0)
	fields := []interface{}{
		&protoID,
	}

	for i := uint8(0); i < tlvLength; i++ {
		err := decode.Decode(buf, fields)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to decode fields")
		}
		pdu.NetworkLayerProtocolIDs[i] = protoID
	}

	return pdu, nil
}

func NewProtocolsSupportedTLV(protocols []uint8) ProtocolsSupportedTLV {
	return ProtocolsSupportedTLV{
		TLVType:                 ProtocolsSupportedTLVType,
		TLVLength:               uint8(len(protocols)),
		NetworkLayerProtocolIDs: protocols,
	}
}

// Type gets the type of the TLV
func (p ProtocolsSupportedTLV) Type() uint8 {
	return p.TLVType
}

// Length gets the length of the TLV
func (p ProtocolsSupportedTLV) Length() uint8 {
	return p.TLVLength
}

// Value gets the TLV itself
func (p ProtocolsSupportedTLV) Value() interface{} {
	return p
}

// Serialize serializes a protocols supported TLV
func (p ProtocolsSupportedTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(p.TLVType)
	buf.WriteByte(p.TLVLength)
	buf.Write(p.NetworkLayerProtocolIDs)
}
