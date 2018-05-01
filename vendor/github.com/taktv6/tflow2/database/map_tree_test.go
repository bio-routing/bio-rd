package database

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKey(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("test", createKey("test"))
	assert.Equal("\x17", createKey(byte(23)))
	assert.Equal("\x00\x17", createKey(uint16(23)))
	assert.Equal("\x17\x2a", createKey([]uint8{23, 42}))
	assert.Equal("\x00\x00\x00\x00\x00\x00\x00\x17", createKey(int64(23)))
	assert.Equal("\x00\x00\x00\x17", createKey(uint32(23)))
	assert.Equal("\x01\x02\x03\x04", createKey(net.IP{1, 2, 3, 4}))
	assert.Equal("\xfe\x80\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01", createKey(net.ParseIP("fe80::1")))

	assert.PanicsWithValuef("unsupported key type: struct {}", func() { createKey(struct{}{}) }, "")
}

func TestMapTree(t *testing.T) {
	assert := assert.New(t)

	mapTree := newMapTree()
	assert.Nil(mapTree.Get("foo"))

	mapTree.Insert("foo", "bar")
	assert.NotNil(mapTree.Get("foo"))
}
