package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	m := &MockConn{}

	payload := []byte{1, 2, 3}
	m.Write(payload)

	assert.Equal(t, payload, m.Bytes)
}

func TestRead(t *testing.T) {
	m := &MockConn{}

	payload := []byte{1, 2, 3}
	m.Bytes = payload

	buffer := make([]byte, 4)
	m.Read(buffer)

	assert.Equal(t, payload, buffer[:3])
}
