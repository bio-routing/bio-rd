package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"

	"errors"

	"bytes"

	"github.com/bio-routing/bio-rd/net"
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
