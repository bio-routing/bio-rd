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

	_, found := globalRegistry.vrfs[vrfName]
	assert.True(t, found, "vrf must be in global registry")

	v.Unregister()

	_, found = globalRegistry.vrfs[vrfName]
	assert.False(t, found, "vrf must not be in global registry")

}

func TestRouteDistinguisherHumanReadable(t *testing.T) {
	tests := []struct {
		name     string
		rdi      uint64
		expected string
	}{
		{
			name:     "Test #1",
			rdi:      0,
			expected: "0:0",
		},
		{
			name:     "Test #2",
			rdi:      123,
			expected: "0:123",
		},
		{
			name:     "Test #3",
			rdi:      220434901565105,
			expected: "51324:65201",
		},
	}

	for _, test := range tests {
		res := RouteDistinguisherHumanReadable(test.rdi)
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestParseHumanReadableRouteDistinguisher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
		wantFail bool
	}{
		{
			name:     "Test #1",
			input:    "51324:65201",
			expected: 0x0000C87C0000FEB1,
			wantFail: false,
		},
		{
			name:     "Test #2",
			input:    "51324",
			wantFail: true,
		},
		{
			name:     "First part invalid",
			input:    "foo:2342",
			wantFail: true,
		},
		{
			name:     "First part invalid",
			input:    "foo:2342",
			wantFail: true,
		},
		{
			name:     "First part too large",
			input:    "4294967297:1",
			wantFail: true,
		},
		{
			name:     "Second part invalid",
			input:    "42:foo",
			wantFail: true,
		},
		{
			name:     "Second part too large",
			input:    "42:4294967297",
			wantFail: true,
		},
	}

	for _, test := range tests {
		res, err := ParseHumanReadableRouteDistinguisher(test.input)
		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q", test.name)
			continue
		}

		if test.wantFail && err == nil {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equal(t, test.expected, res, test.name)
	}
}
