package decoder

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

// Decode decodes network packets
func Decode(buf *bytes.Buffer, fields []interface{}) error {
	var err error
	for _, field := range fields {
		err = binary.Read(buf, binary.BigEndian, field)
		if err != nil {
			return errors.Wrap(err, "Unable to read from buffer")
		}
	}
	return nil
}
