package packet

import (
	"bytes"
	"fmt"
)

const ISNeighborsTLVType = 6

type ISNeighborsTLV struct {
	TLVType      uint8
	TLVLength    uint8
	NeighborSNPA [6]byte
}

const ISNeighborsTLVLength = 8

func readISNeighborsTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ISNeighborsTLV, uint8, error) {
	pdu := &ISNeighborsTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []interface{}{
		&pdu.NeighborSNPA,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, 0, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, ISNeighborsTLVLength, nil
}

func (i ISNeighborsTLV) Type() uint8 {
	return i.TLVType
}

func (i ISNeighborsTLV) Length() uint8 {
	return i.TLVLength
}
