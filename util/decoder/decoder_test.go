package decoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	input := []byte{
		3, 0, 0, 0, 100, 200,
	}

	type testData struct {
		a uint8
		b uint32
		c []byte
	}

	s := testData{
		c: make([]byte, 1),
	}

	fields := []interface{}{
		&s.a,
		&s.b,
		&s.c,
	}

	buf := bytes.NewBuffer(input)
	Decode(buf, fields)

	expected := testData{
		a: 3,
		b: 100,
		c: []byte{200},
	}

	assert.Equal(t, expected, s)
}
