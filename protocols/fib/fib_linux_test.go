package fib

import (
	"net"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestNetlinkRouteDiff(t *testing.T) {
	tests := []struct {
		name     string
		left     []netlink.Route
		right    []netlink.Route
		expected []netlink.Route
	}{
		{
			name: "Equal",
			left: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
			right: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
			expected: []netlink.Route{},
		}, {
			name: "Left empty",
			left: []netlink.Route{},
			right: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
			expected: []netlink.Route{},
		}, {
			name: "Right empty",
			left: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
			right: []netlink.Route{},
			expected: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
		}, {
			name: "Diff",
			left: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
			right: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 1,
				},
			},
			expected: []netlink.Route{
				{
					Dst: &net.IPNet{
						IP:   net.IPv4(20, 0, 0, 1),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					Table: 2,
				},
			},
		},
	}

	for _, test := range tests {
		res := NetlinkRouteDiff(test.left, test.right)
		assert.Equal(t, test.expected, res)
	}

}

func TestNewPathsFromNetlinkRoute(t *testing.T) {
	tests := []struct {
		name          string
		source        netlink.Route
		expectedPfx   bnet.Prefix
		expectedPaths []*Path
		expectError   bool
	}{
		{
			name: "Simple",
			source: netlink.Route{
				Dst:      bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).GetIPNet(),
				Src:      bnet.IPv4(456).Bytes(),
				Gw:       bnet.IPv4(789).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			expectedPaths: []*Path{
				{
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv4(456),
						NextHop:  bnet.IPv4(789),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Multiple nexthop",
			source: netlink.Route{
				Dst: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).GetIPNet(),
				Src: bnet.IPv4(456).Bytes(),
				MultiPath: []*netlink.NexthopInfo{
					{
						LinkIndex: 1,
						Hops:      1,
						Gw:        bnet.IPv4(123).Bytes(),
						Flags:     0,
						NewDst:    nil,
						Encap:     nil,
					}, {
						LinkIndex: 2,
						Hops:      1,
						Gw:        bnet.IPv4(345).Bytes(),
						Flags:     0,
						NewDst:    nil,
						Encap:     nil,
					},
				},
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			expectedPaths: []*Path{
				{
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv4(456),
						NextHop:  bnet.IPv4(123),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				}, {
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv4(456),
						NextHop:  bnet.IPv4(345),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "No source but destination",
			source: netlink.Route{
				Dst:      bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).GetIPNet(),
				Gw:       bnet.IPv4(789).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
			expectedPaths: []*Path{
				{
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv4(0),
						NextHop:  bnet.IPv4(789),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Source but no destination",
			source: netlink.Route{
				Src:      bnet.IPv4(456).Bytes(),
				Gw:       bnet.IPv4(789).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx: bnet.NewPfx(bnet.IPv4FromOctets(0, 0, 0, 0), 0),
			expectedPaths: []*Path{
				{
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv4(456),
						NextHop:  bnet.IPv4(789),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "No source but no destination",
			source: netlink.Route{
				Gw:       bnet.IPv4(789).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx:   bnet.Prefix{},
			expectedPaths: []*Path{},
			expectError:   true,
		},
		{
			name: "No source but destination IPv6",
			source: netlink.Route{
				Dst:      bnet.NewPfx(bnet.IPv6(2001, 0), 48).GetIPNet(),
				Gw:       bnet.IPv6(2001, 123).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx: bnet.NewPfx(bnet.IPv6(2001, 0), 48),
			expectedPaths: []*Path{
				{
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv6(0, 0),
						NextHop:  bnet.IPv6(2001, 123),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Source but no destination IPv6",
			source: netlink.Route{
				Src:      bnet.IPv6(2001, 456).Bytes(),
				Gw:       bnet.IPv6(2001, 789).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx: bnet.NewPfx(bnet.IPv6(0, 0), 0),
			expectedPaths: []*Path{
				{
					Type: NetlinkPathType,
					NetlinkPath: &NetlinkPath{
						Src:      bnet.IPv6(2001, 456),
						NextHop:  bnet.IPv6(2001, 789),
						Protocol: ProtoKernel,
						Priority: 1,
						Table:    254,
						Type:     1,
						Kernel:   true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "no source no destination",
			source: netlink.Route{
				Gw:       bnet.IPv4(123).Bytes(),
				Protocol: ProtoKernel,
				Priority: 1,
				Table:    254,
				Type:     1,
			},
			expectedPfx:   bnet.NewPfx(bnet.IPv4(0), 0),
			expectedPaths: []*Path{{}},
			expectError:   true,
		},
	}

	for _, test := range tests {
		pfx, paths, err := NewPathsFromNlRoute(test.source, true)
		if test.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equalf(t, test.expectedPaths, paths, test.name)
			assert.Equalf(t, test.expectedPfx, pfx, test.name)
		}
	}
}
