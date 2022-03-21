package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Decode decodes network packets
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
