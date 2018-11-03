package testing

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	m := &MockConn{
		buf: bytes.NewBuffer(nil),
	}

	payload := []byte{1, 2, 3}
	m.Write(payload)

	assert.Equal(t, payload, m.buf.Bytes())
}

func TestRead(t *testing.T) {
	m := &MockConn{
		buf: bytes.NewBuffer(nil),
	}

	payload := []byte{1, 2, 3}
	m.buf.Write(payload)

	buffer := make([]byte, 4)
	n, _ := m.Read(buffer)

	assert.Equal(t, payload, buffer[:n])
}
