package packet

import (
	"bytes"
	"fmt"
)

// UnknownTLV represents an unknown TLV
type UnknownTLV struct {
	TLVType   uint8
	TLVLength uint8
	TLVValue  []byte
}

func readUnknownTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*UnknownTLV, error) {
	pdu := &UnknownTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
		TLVValue:  make([]byte, tlvLength),
	}

	n, err := buf.Read(pdu.TLVValue)
	if err != nil {
		return nil, fmt.Errorf("unable to read: %v", err)
	}

	if n != int(tlvLength) {
		return nil, fmt.Errorf("read of TLVType %d incomplete", pdu.TLVType)
	}

	return pdu, nil
}

func (u UnknownTLV) Copy() TLV {
	ret := &UnknownTLV{}
	ret.TLVValue = make([]byte, 0, len(u.TLVValue))
	copy(ret.TLVValue, u.TLVValue)
	return ret
}

// Type gets the type of the TLV
func (u UnknownTLV) Type() uint8 {
	return u.TLVType
}

// Length gets the length of the TLV
func (u UnknownTLV) Length() uint8 {
	return u.TLVLength
}

// Value gets the TLV itself
func (u *UnknownTLV) Value() interface{} {
	return u
}

// Serialize serializes an unknown TLV
func (u UnknownTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(u.TLVType)
	buf.WriteByte(u.TLVLength)
	buf.Write(u.TLVValue)
}
