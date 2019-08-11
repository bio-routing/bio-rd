package packet

import (
	"bytes"
	"fmt"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
	"io"
	"time"
)

const minimalHeaderLength = 12

// The Decode function reads from the input and calls the target with every successfully
// decoded MTR record.
// If there is an error while decoding, all decoding stops and the error is returned
func Decode(input io.Reader, target func(MTRRecord)) error {
	minimalHeader := bytes.NewBuffer(make([]byte, minimalHeaderLength))
	minimalHeader.Reset()
	payload := bytes.NewBuffer(nil)
	for {
		// copy header to check if next header exits.
		read, err := io.CopyN(minimalHeader, input, minimalHeaderLength)
		if err == io.EOF {
			if read == 0 {
				return nil
			}
			return fmt.Errorf("read %d bytes, expected %d for next mtr header", read, minimalHeaderLength)
		}
		if err != nil {
			return errors.Wrap(err, "failed to read mtr header from input")
		}
		record, err := decodeHeader(minimalHeader)
		if err != nil {
			return errors.Wrap(err, "failed to decode header")
		}
		_, err = io.CopyN(payload, input, int64(record.Length))
		if err != nil {
			return errors.Wrap(err, "failed to copy expected message length from input stream")
		}
		err = record.Message.Decode(payload)
		if err != nil {
			return errors.Wrap(err, "failed to decode mtr message")
		}
		target(record)
	}
}

// decodeHeader decodes
func decodeHeader(data *bytes.Buffer) (MTRRecord, error) {
	var seconds uint32
	var ret MTRRecord
	err := decode.Decode(data, []interface{}{&seconds, &ret.Type.Type, &ret.Type.SubType, &ret.Length})
	if err != nil {
		return ret, errors.Wrap(err, "failed to decode mtr header")
	}
	ret.Message, err = messageForType(ret.Type)
	ret.TimeStamp = time.Unix(int64(seconds), 0).UTC()
	return ret, errors.Wrap(err, "failed to set message for given type")
}
