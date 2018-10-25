package testing

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	m := &MockConn{
		Buf: bytes.NewBuffer(nil),
	}

	payload := []byte{1, 2, 3}
	m.Write(payload)

	assert.Equal(t, payload, m.Buf.Bytes())
}

func TestRead(t *testing.T) {
	m := &MockConn{
		Buf: bytes.NewBuffer(nil),
	}

	payload := []byte{1, 2, 3}
	m.Buf.Write(payload)

	buffer := make([]byte, 4)
	n, _ := m.Read(buffer)

	assert.Equal(t, payload, buffer[:n])
}
