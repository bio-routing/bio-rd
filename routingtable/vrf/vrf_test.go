package vrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithDuplicate(t *testing.T) {
	_, err := New("master")
	assert.Nil(t, err, "no error on first invocation")

	_, err = New("master")
	assert.NotNil(t, err, "ambigious VRF name")
}

func TestIPv4UnicastRIBWith(t *testing.T) {
	v := newUntrackedVRF("master")
	rib, err := v.CreateIPv4UnicastLocRIB("inet.0")

	assert.Equal(t, rib, v.IPv4UnicastRIB())
	assert.Nil(t, err, "error must be nil")
}

func TestIPv6UnicastRIB(t *testing.T) {
	v := newUntrackedVRF("master")
	rib, err := v.CreateIPv6UnicastLocRIB("inet6.0")

	assert.Equal(t, rib, v.IPv6UnicastRIB())
	assert.Nil(t, err, "error must be nil")
}

func TestCreateLocRIBTwice(t *testing.T) {
	v := newUntrackedVRF("master")
	_, err := v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.Nil(t, err, "error must be nil on first invokation")

	_, err = v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.NotNil(t, err, "error must not be nil on second invokation")
}

func TestRIBByName(t *testing.T) {
	v := newUntrackedVRF("master")
	rib, _ := v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.NotNil(t, rib, "rib must not be nil after creation")

	foundRIB, found := v.RIBByName("inet6.0")
	assert.True(t, found)
	assert.Exactly(t, rib, foundRIB)
}
