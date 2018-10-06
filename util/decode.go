package util

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
			return fmt.Errorf("Unable to read from buffer: %v", err)
		}
	}
	return nil
}
