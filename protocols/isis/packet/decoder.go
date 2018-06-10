package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func decode(buf *bytes.Buffer, fields []interface{}) error {
	var err error
	for _, field := range fields {
		switch field := field.(type) {
		case *[6]byte:
			_, err := buf.Read(field[:])
			if err != nil {
				return fmt.Errorf("Unable to read from buffer: %v", err)
			}
		default:
			err = binary.Read(buf, binary.BigEndian, field)
			if err != nil {
				return fmt.Errorf("Unable to read from buffer: %v", err)
			}
		}
	}
	return nil
}
