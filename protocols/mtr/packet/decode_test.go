package packet

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDecodeHeader(t *testing.T) {
	testcases := []struct {
		name   string
		input  []byte
		output MTRRecord
		error  string
	}{
		{
			name: "peer table index",
			input: []byte{0, 1, 2, 3, // TimeStamp
				0, 13, 0, 1, // Type
				0, 0, 0, 0, // Length
			},
			output: MTRRecord{
				TimeStamp: time.Date(1970, 01, 01, 18, 20, 51, 0, time.UTC),
				Type: MessageType{
					TABLE_DUMP_V2, PEER_INDEX_TABLE,
				},
				Length:  0,
				Message: &PeerIndexTable{},
			},
		},
		{
			name: "short read peer table index",
			input: []byte{0, 1, 2, 3, // TimeStamp
				0, 13, 0, 1, // Type
				0, // Length
			},
			output: MTRRecord{
				Type: MessageType{
					TABLE_DUMP_V2, PEER_INDEX_TABLE,
				},
			},
			error: "failed to decode mtr header: Unable to read from buffer: unexpected EOF",
		},
		{
			name: "unknown type",
			input: []byte{0, 1, 2, 3, // TimeStamp
				0, 1, 0, 1, // Type
				0, 0, 0, 0, // Length
			},
			output: MTRRecord{
				TimeStamp: time.Date(1970, 01, 01, 18, 20, 51, 0, time.UTC),
				Type: MessageType{
					1, 1,
				},
			},
			error: "failed to set message for given type: given type 1 subtype 1 not implemented",
		},
		{
			name: "unknown subtype",
			input: []byte{0, 1, 2, 3, // TimeStamp
				0, 13, 0, 10, // Type
				0, 0, 0, 0, // Length
			},
			output: MTRRecord{
				TimeStamp: time.Date(1970, 01, 01, 18, 20, 51, 0, time.UTC),
				Type: MessageType{
					TABLE_DUMP_V2, 10,
				},
			},
			error: "failed to set message for given type: unknown subtype 10 for TABLE_DUMP_V2",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tc.input)
			out, err := decodeHeader(buf)
			assert.Equal(t, tc.output, out)
			if tc.error != "" {
				assert.EqualError(t, err, tc.error)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
