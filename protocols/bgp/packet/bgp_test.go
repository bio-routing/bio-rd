package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAFIName(t *testing.T) {
	afiIPv4 := AFIName(1)
	assert.Equal(t, "IPv4", afiIPv4)

	afiIPv6 := AFIName(2)
	assert.Equal(t, "IPv6", afiIPv6)

	afiUnknown := AFIName(0)
	assert.Equal(t, "Unknown AFI", afiUnknown)
}

func TestBGPErrorError(t *testing.T) {
	e := BGPError{
		ErrorCode:    2,
		ErrorSubCode: 3,
		ErrorStr:     "Unknown Error TestBGPErrorError",
	}

	actual := e.Error()
	expected := "Unknown Error TestBGPErrorError"
	assert.Equal(t, expected, actual)
}
