package decode

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Decode reads fields from a buffer
func Decode(buf *bytes.Buffer, fields []interface{}) error {
	var err error
	for _, field := range fields {
		err = binary.Read(buf, binary.BigEndian, field)
		if err != nil {
			return fmt.Errorf("unable to read from buffer: %w", err)
		}
	}
	return nil
}

// DecodeUint8 decodes an uint8
func DecodeUint8(buf *bytes.Buffer, x *uint8) error {
	y, err := buf.ReadByte()
	if err != nil {
		return err
	}

	*x = y
	return nil
}

// DecodeUint16 decodes an uint16
func DecodeUint16(buf *bytes.Buffer, x *uint16) error {
	a, err := buf.ReadByte()
	if err != nil {
		return err
	}

	b, err := buf.ReadByte()
	if err != nil {
		return err
	}

	*x = uint16(a)<<8 + uint16(b)
	return nil
}

// DecodeUint32 decodes an uint32
func DecodeUint32(buf *bytes.Buffer, x *uint32) error {
	a, err := buf.ReadByte()
	if err != nil {
		return err
	}

	b, err := buf.ReadByte()
	if err != nil {
		return err
	}

	c, err := buf.ReadByte()
	if err != nil {
		return err
	}

	d, err := buf.ReadByte()
	if err != nil {
		return err
	}

	*x = uint32(a)<<24 + uint32(b)<<16 + uint32(c)<<8 + uint32(d)
	return nil
}
