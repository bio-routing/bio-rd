package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

// AreaAddressesTLVType is the type value of an area address TLV
const AreaAddressesTLVType = 1

// AreaAddressesTLV represents an area address TLV
type AreaAddressesTLV struct {
	TLVType   uint8
	TLVLength uint8
	AreaIDs   []types.AreaID
}

func readAreaAddressesTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*AreaAddressesTLV, error) {
	pdu := &AreaAddressesTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
		AreaIDs:   make([]types.AreaID, 0),
	}

	areaNum := 0
	read := uint8(0)
	for read < tlvLength {
		areaLen, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("Unable to read: %v", err)
		}
		read++

		newArea := make(types.AreaID, areaLen)
		_, err = buf.Read(newArea)
		if err != nil {
			return nil, fmt.Errorf("Unable to read: %v", err)
		}
		read += areaLen

		pdu.AreaIDs = append(pdu.AreaIDs, newArea)
		areaNum++
	}

	return pdu, nil
}

// NewAreaAddressesTLV creates a new area addresses TLV
func NewAreaAddressesTLV(areas []types.AreaID) *AreaAddressesTLV {
	a := &AreaAddressesTLV{
		TLVType:   AreaAddressesTLVType,
		TLVLength: 1,
		AreaIDs:   make([]types.AreaID, len(areas)),
	}

	for i, area := range areas {
		a.TLVLength += uint8(len(areas[i]))
		a.AreaIDs[i] = area
	}

	return a
}

// Type gets the type of the TLV
func (a AreaAddressesTLV) Type() uint8 {
	return a.TLVType
}

// Length gets the length of the TLV
func (a AreaAddressesTLV) Length() uint8 {
	return a.TLVLength
}

// Value gets the TLV itself
func (a AreaAddressesTLV) Value() interface{} {
	return a
}

// Serialize serializes an area address TLV
func (a AreaAddressesTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(a.TLVType)
	buf.WriteByte(a.TLVLength)

	for _, area := range a.AreaIDs {
		buf.WriteByte(uint8(len(area)))
		buf.Write(area)
	}
}
