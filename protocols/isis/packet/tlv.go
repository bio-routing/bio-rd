package packet

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type TLV interface {
	Type() uint8
	Length() uint8
}

func readTLVs(buf *bytes.Buffer, length uint16) ([]TLV, error) {
	var err error
	tlvType := uint8(0)
	tlvLength := uint8(0)

	headFields := []interface{}{
		&tlvType,
		&tlvLength,
	}

	TLVs := make([]TLV, 0)

	read := uint16(0)
	for read < length {
		err = decode(buf, headFields)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode fields: %v", err)
		}

		read += 2

		var tlv TLV
		l := uint8(0)

		switch tlvType {
		case ISNeighborsTLVType:
			tlv, l, err = readISNeighborsTLV(buf, tlvType, tlvLength)
		case ProtocolsSupportedTLVType:
			tlv, l, err = readProtocolsSupportedTLV(buf, tlvType, tlvLength)
		case IPInterfaceAddressTLVType:
			tlv, l, err = readIPInterfaceAddressTLV(buf, tlvType, tlvLength)
		case AreaAddressTLVType:
			tlv, l, err = readAreaAddressTLV(buf, tlvType, tlvLength)
		default:
			log.Warningf("Unknown type: %d", tlvType)
			for i := uint8(0); i < tlvLength; i++ {
				_, err = buf.ReadByte()
				if err != nil {
					return nil, fmt.Errorf("Unable to read: %v", err)
				}
				read++
			}
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("Unable to read IS neighbors TLV: %v", err)
		}
		read += uint16(l)
		TLVs = append(TLVs, tlv)
	}

	return TLVs, nil
}
