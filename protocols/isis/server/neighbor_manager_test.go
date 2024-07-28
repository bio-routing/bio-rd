package server

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateAreasL1(t *testing.T) {
	tests := []struct {
		name          string
		nifa          *netIfa
		receivedAreas []types.AreaID
		expected      bool
	}{
		{
			name: "test #1",
			nifa: &netIfa{
				srv: &Server{
					nets: []*types.NET{
						{
							AreaID: []byte{0x49, 1},
						},
					},
				},
			},
			receivedAreas: []types.AreaID{
				{
					0x49, 0x01,
				},
			},
			expected: true,
		},
		{
			name: "test #2",
			nifa: &netIfa{
				srv: &Server{
					nets: []*types.NET{
						{
							AreaID: []byte{0x49},
						},
						{
							AreaID: []byte{0x49, 0xff},
						},
					},
				},
			},
			receivedAreas: []types.AreaID{
				{
					0x49, 0xff,
				},
			},
			expected: true,
		},
		{
			name: "test #3",
			nifa: &netIfa{
				srv: &Server{
					nets: []*types.NET{
						{
							AreaID: []byte{0x49, 1},
						},
						{
							AreaID: []byte{0x49, 0xff},
						},
					},
				},
			},
			receivedAreas: []types.AreaID{
				{
					0x49, 0xfe,
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		res := test.nifa.validateAreasL1(test.receivedAreas)
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestValidateProtocolsSupported(t *testing.T) {
	tests := []struct {
		name      string
		protocols []uint8
		expected  bool
	}{
		{
			name: "Test #1",
			protocols: []uint8{
				packet.NLPIDIPv4,
				packet.NLPIDIPv6,
			},
			expected: true,
		},
		{
			name: "Test #2",
			protocols: []uint8{
				packet.NLPIDIPv4,
			},
			expected: false,
		},
		{
			name: "Test #3",
			protocols: []uint8{
				packet.NLPIDIPv6,
			},
			expected: false,
		},
		{
			name:      "Test #4",
			protocols: []uint8{},
			expected:  false,
		},
	}

	for _, test := range tests {
		res := validateProtocolsSupported(test.protocols)
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestValidateIPv4Addresses(t *testing.T) {
	tests := []struct {
		name     string
		nifa     *netIfa
		addrs    []uint32
		expected bool
	}{
		{
			name: "Test #1",
			nifa: &netIfa{
				devStatus: &mockDevice{
					addrs: []*bnet.Prefix{
						bnet.NewPfx(bnet.IPv4(110), 31).Ptr(),
					},
				},
			},
			addrs: []uint32{
				111,
			},
			expected: true,
		},
		{
			name: "Test #2",
			nifa: &netIfa{
				devStatus: &mockDevice{
					addrs: []*bnet.Prefix{
						bnet.NewPfx(bnet.IPv4(110), 32).Ptr(),
					},
				},
			},
			addrs: []uint32{
				111,
			},
			expected: false,
		},
		{
			name: "Test #3",
			nifa: &netIfa{
				devStatus: &mockDevice{
					addrs: []*bnet.Prefix{
						bnet.NewPfx(bnet.IPv4(220), 31).Ptr(),
						bnet.NewPfx(bnet.IPv4(110), 31).Ptr(),
					},
				},
			},
			addrs: []uint32{
				111,
			},
			expected: true,
		},
	}

	for _, test := range tests {
		res := test.nifa.validateIPv4Addresses(test.addrs)
		assert.Equal(t, test.expected, res, test.name)
	}
}
