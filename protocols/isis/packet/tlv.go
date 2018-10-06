package packet

import (
	"bytes"
	"fmt"
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
	var err error
	tlvType := uint8(0)
	tlvLength := uint8(0)

	headFields := []interface{}{
		&tlvType,
		&tlvLength,
	}

	TLVs := make([]TLV, 0)

	length := buf.Len()
	read := uint16(0)
	for read < uint16(length) {
		err = decode(buf, headFields)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode fields: %v", err)
		}

		read += 2
		read += uint16(tlvLength)

		var tlv TLV

		fmt.Printf("Decode: TLV Type = %d\n", tlvType)
		fmt.Printf("Length: %d\n", tlvLength)

		switch tlvType {
		case DynamicHostNameTLVType:
			tlv, err = readDynamicHostnameTLV(buf, tlvType, tlvLength)
		case ChecksumTLVType:
			tlv, err = readChecksumTLV(buf, tlvType, tlvLength)
		case ProtocolsSupportedTLVType:
			tlv, _, err = readProtocolsSupportedTLV(buf, tlvType, tlvLength)
		case IPInterfaceAddressTLVType:
			tlv, _, err = readIPInterfaceAddressTLV(buf, tlvType, tlvLength)
		case AreaAddressesTLVType:
			tlv, err = readAreaAddressesTLV(buf, tlvType, tlvLength)
		case P2PAdjacencyStateTLVType:
			tlv, _, err = readP2PAdjacencyStateTLV(buf, tlvType, tlvLength)
		default:
			tlv, err = readUnknownTLV(buf, tlvType, tlvLength)
		}

		if err != nil {
			return nil, fmt.Errorf("Unable to read TLV: %v", err)
		}
		TLVs = append(TLVs, tlv)
	}

	return TLVs, nil
}
