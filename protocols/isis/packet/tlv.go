package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

const (
	tlvBaseLen = 2 // Type + Length field
)

// TLV is an interface that all TLVs must fulfill
type TLV interface {
	Type() uint8
	Length() uint8
	Serialize(*bytes.Buffer)
	Value() interface{}
}

func serializeTLVs(tlvs []TLV) []byte {
	buf := bytes.NewBuffer(nil)

	for _, tlv := range tlvs {
		tlv.Serialize(buf)
	}

	return buf.Bytes()
}

func readTLVs(buf *bytes.Buffer) ([]TLV, error) {
	TLVs := make([]TLV, 0)
	for buf.Len() > 0 {
		tlv, err := readTLV(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to read TLV: %w", err)
		}

		TLVs = append(TLVs, tlv)
	}

	return TLVs, nil
}

func readTLV(buf *bytes.Buffer) (TLV, error) {
	var err error
	tlvType := uint8(0)
	tlvLength := uint8(0)

	headFields := []interface{}{
		&tlvType,
		&tlvLength,
	}

	err = decode.Decode(buf, headFields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	var tlv TLV
	switch tlvType {
	case DynamicHostNameTLVType:
		tlv, err = readDynamicHostnameTLV(buf, tlvType, tlvLength)
	case ChecksumTLVType:
		tlv, err = readChecksumTLV(buf, tlvType, tlvLength)
	case ProtocolsSupportedTLVType:
		tlv, err = readProtocolsSupportedTLV(buf, tlvType, tlvLength)
	case IPInterfaceAddressesTLVType:
		tlv, err = readIPInterfaceAddressesTLV(buf, tlvType, tlvLength)
	case AreaAddressesTLVType:
		tlv, err = readAreaAddressesTLV(buf, tlvType, tlvLength)
	case P2PAdjacencyStateTLVType:
		tlv, err = readP2PAdjacencyStateTLV(buf, tlvType, tlvLength)
	case ISNeighborsTLVType:
		tlv, err = readISNeighborsTLV(buf, tlvType, tlvLength)
	case LSPEntriesTLVType:
		tlv, err = readLSPEntriesTLV(buf, tlvType, tlvLength)
	default:
		tlv, err = readUnknownTLV(buf, tlvType, tlvLength)
	}

	if err != nil {
		return nil, fmt.Errorf("Unable to read TLV: %v", err)
	}

	return tlv, nil
}
