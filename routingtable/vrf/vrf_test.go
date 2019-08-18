package vrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithDuplicate(t *testing.T) {
	_, err := New("master", 123)
	assert.Nil(t, err, "no error on first invocation")

	_, err = New("master", 123)
	assert.NotNil(t, err, "ambigious VRF name")
}

func TestIPv4UnicastRIBWith(t *testing.T) {
	v := newUntrackedVRF("master", 0)
	rib, err := v.CreateIPv4UnicastLocRIB("inet.0")

	assert.Equal(t, rib, v.IPv4UnicastRIB())
	assert.Nil(t, err, "error must be nil")
}

func TestIPv6UnicastRIB(t *testing.T) {
	v := newUntrackedVRF("master", 0)
	rib, err := v.CreateIPv6UnicastLocRIB("inet6.0")

	assert.Equal(t, rib, v.IPv6UnicastRIB())
	assert.Nil(t, err, "error must be nil")
}

func TestCreateLocRIBTwice(t *testing.T) {
	v := newUntrackedVRF("master", 0)
	_, err := v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.Nil(t, err, "error must be nil on first invokation")

	_, err = v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.NotNil(t, err, "error must not be nil on second invokation")
}

func TestRIBByName(t *testing.T) {
	v := newUntrackedVRF("master", 0)
	rib, _ := v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.NotNil(t, rib, "rib must not be nil after creation")

	foundRIB, found := v.RIBByName("inet6.0")
	assert.True(t, found)
	assert.Exactly(t, rib, foundRIB)
}

func TestName(t *testing.T) {
	v := newUntrackedVRF("foo", 0)
	assert.Equal(t, "foo", v.Name())
}

func TestUnregister(t *testing.T) {
	vrfName := "registeredVRF"
	v, err := New(vrfName, 10)
	assert.Nil(t, err, "error must be nil on first invokation")

	_, err = New(vrfName, 10)
	assert.NotNil(t, err, "error must not be nil on second invokation")

	_, found := globalRegistry.vrfs[10]
	assert.True(t, found, "vrf must be in global registry")

	v.Unregister()

	_, found = globalRegistry.vrfs[10]
	assert.False(t, found, "vrf must not be in global registry")

}
