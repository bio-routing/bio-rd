package decoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	type testData struct {
		a uint8
		b uint32
		c []byte
	}

	tests := []struct {
		name     string
		input    []byte
		expected testData
		wantFail bool
	}{
		{
			name: "valid input",
			input: []byte{
				3, 0, 0, 0, 100, 200,
			},
			expected: testData{
				a: 3,
				b: 100,
				c: []byte{200},
			},
		},
		{
			name: "input too short",
			input: []byte{
				3, 0, 0, 0, 100,
			},
			wantFail: true,
		},
		{
			name:     "input null",
			wantFail: true,
		},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := testData{
				c: make([]byte, 1),
			}

			fields := []interface{}{
				&s.a,
				&s.b,
				&s.c,
			}

			buf := bytes.NewBuffer(test.input)

			err := Decode(buf, fields)
			if err != nil {
				if !test.wantFail {
					t.Fatalf("Unexpected error: %s", err)
				}

				return
			}

			assert.Equal(t, test.expected, s)
		})
	}
}
