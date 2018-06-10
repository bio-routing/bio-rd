package packet

import (
	"bytes"
	"fmt"
)

const AreaAddressTLVType = 1

type AreaAddressTLV struct {
	TLVType    uint8
	TLVLength  uint8
	AreaLength uint8
	AreaID     []byte
}

func readAreaAddressTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*AreaAddressTLV, uint8, error) {
	pdu := &AreaAddressTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
		AreaID:    make([]uint8, tlvLength),
	}

	fields := []interface{}{
		&pdu.AreaLength,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to decode fields: %v", err)
	}

	pdu.AreaID = make([]byte, pdu.AreaLength)
	n, err := buf.Read(pdu.AreaID)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to read: %v", err)
	}

	return pdu, uint8(n + 1), nil
}

func (i AreaAddressTLV) Type() uint8 {
	return i.TLVType
}

func (i AreaAddressTLV) Length() uint8 {
	return i.TLVLength
}
