package server

import (
	"bytes"
	"errors"
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"

	"github.com/stretchr/testify/assert"
)

func TestWithDrawPrefixes(t *testing.T) {
	testcases := []struct {
		Name          string
		Prefix        []net.Prefix
		Expected      []byte
		ExpectedError error
	}{
		{
			Name:   "One withdraw",
			Prefix: []net.Prefix{net.NewPfx(net.IPv4(1413010532), 24)},
			Expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // BGP Marker
				0x00, 0x1b, // BGP Message Length
				0x02,       // BGP Message Type == Update
				0x00, 0x04, // WithDraw Octet length
				0x18,             // Prefix Length
				0x54, 0x38, 0xd4, // Prefix,
				0x00, 0x00, // Total Path Attribute Length
			},
			ExpectedError: nil,
		},
		{
			Name:   "two withdraws",
			Prefix: []net.Prefix{net.NewPfx(net.IPv4(1413010532), 24), net.NewPfx(net.IPv4(1413010534), 25)},
			Expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // BGP Marker
				0x00, 0x20, // BGP Message Length
				0x02,       // BGP Message Type == Update
				0x00, 0x09, // WithDraw Octet length
				0x18,             // Prefix Length first
				0x54, 0x38, 0xd4, // Prefix,
				0x19,                   // Prefix Length second
				0x54, 0x38, 0xd4, 0x66, // Prefix,
				0x00, 0x00, // Total Path Attribute Length
			},
			ExpectedError: nil,
		},
	}
	for _, tc := range testcases {
		buf := bytes.NewBuffer([]byte{})
		opt := &types.Options{}
		err := withDrawPrefixes(buf, opt, tc.Prefix...)
		assert.Equal(t, tc.ExpectedError, err, "error mismatch in testcase %v", tc.Name)
		assert.Equal(t, tc.Expected, buf.Bytes(), "expected different bytes in testcase %v", tc.Name)
	}
}

func TestWithDrawPrefixesMultiProtocol(t *testing.T) {
	tests := []struct {
		Name     string
		Prefix   net.Prefix
		Expected []byte
	}{
		{
			Name:   "IPv6 MP_UNREACH_NLRI",
			Prefix: net.NewPfx(net.IPv6FromBlocks(0x2804, 0x148c, 0, 0, 0, 0, 0, 0), 32),
			Expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // BGP Marker
				0x00, 0x22, // BGP Message Length
				0x02,       // BGP Message Type == Update
				0x00, 0x00, // WithDraw Octet length
				0x00, 0x0b, // Length
				0x80,       // Flags
				0x0f,       // Attribute Code
				0x08,       // Attribute length
				0x00, 0x02, // AFI
				0x01,                         // SAFI
				0x20, 0x28, 0x04, 0x14, 0x8c, // Prefix
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			opt := &types.Options{
				AddPathRX: false,
			}
			err := withDrawPrefixesMultiProtocol(buf, opt, test.Prefix, packet.IPv6AFI, packet.UnicastSAFI)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, test.Expected, buf.Bytes())
		})
	}
}

func TestWithDrawPrefixesAddPath(t *testing.T) {
	testcases := []struct {
		Name          string
		Prefix        net.Prefix
		Path          *route.Path
		Expected      []byte
		ExpectedError error
	}{
		{
			Name:   "Normal withdraw",
			Prefix: net.NewPfx(net.IPv4(1413010532), 24),
			Path: &route.Path{
				Type: route.BGPPathType,
				BGPPath: &route.BGPPath{
					PathIdentifier: 1,
				},
			},
			Expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // BGP Marker
				0x00, 0x1f, // BGP Message Length
				0x02,       // BGP Message Type == Update
				0x00, 0x08, // WithDraw Octet length
				0x00, 0x00, 0x00, 0x01, // NLRI Path Identifier
				0x18,             // Prefix Length
				0x54, 0x38, 0xd4, // Prefix,
				0x00, 0x00, // Total Path Attribute Length
			},
			ExpectedError: nil,
		},
		{
			Name:   "Non bgp withdraw",
			Prefix: net.NewPfx(net.IPv4(1413010532), 24),
			Path: &route.Path{
				Type: route.StaticPathType,
			},
			Expected:      []byte{},
			ExpectedError: errors.New("wrong path type, expected BGPPathType"),
		},
		{
			Name:   "Nil BGPPathType",
			Prefix: net.NewPfx(net.IPv4(1413010532), 24),
			Path: &route.Path{
				Type: route.BGPPathType,
			},
			Expected:      []byte{},
			ExpectedError: errors.New("got nil BGPPath"),
		},
	}
	for _, tc := range testcases {
		buf := bytes.NewBuffer([]byte{})
		opt := &types.Options{
			AddPathRX: true,
		}
		err := withDrawPrefixesAddPath(buf, opt, tc.Prefix, tc.Path)
		assert.Equal(t, tc.ExpectedError, err, "error mismatch in testcase %v", tc.Name)
		assert.Equal(t, tc.Expected, buf.Bytes(), "expected different bytes in testcase %v", tc.Name)
	}
}
