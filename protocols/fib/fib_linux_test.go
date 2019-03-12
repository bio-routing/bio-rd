package fib

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestConvertNlRouteToFIBPath(t *testing.T) {
	tests := []struct {
		name        string
		source      []netlink.Route
		expected    []route.PrefixPathsPair
		expectError bool
	}{
		{
			name: "Simple",
			source: []netlink.Route{
				netlink.Route{
					Dst:      bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).GetIPNet(),
					Src:      bnet.IPv4(456).Bytes(),
					Gw:       bnet.IPv4(789).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},
			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
					Paths: []*route.FIBPath{
						&route.FIBPath{
							Src:      bnet.IPv4(456),
							NextHop:  bnet.IPv4(789),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
					},
					Err: nil,
				},
			},
			expectError: false,
		},
		{
			name: "Multiple nexthop",
			source: []netlink.Route{
				netlink.Route{
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
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},
			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
					Paths: []*route.FIBPath{
						&route.FIBPath{
							Src:      bnet.IPv4(456),
							NextHop:  bnet.IPv4(123),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
						&route.FIBPath{
							Src:      bnet.IPv4(456),
							NextHop:  bnet.IPv4(345),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "No source but destination",
			source: []netlink.Route{
				netlink.Route{
					Dst:      bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).GetIPNet(),
					Gw:       bnet.IPv4(789).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},

			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8),
					Paths: []*route.FIBPath{
						&route.FIBPath{
							Src:      bnet.IPv4(0),
							NextHop:  bnet.IPv4(789),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Source but no destination",
			source: []netlink.Route{
				netlink.Route{
					Src:      bnet.IPv4(456).Bytes(),
					Gw:       bnet.IPv4(789).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},

			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx: bnet.NewPfx(bnet.IPv4FromOctets(0, 0, 0, 0), 0),
					Paths: []*route.FIBPath{
						&route.FIBPath{
							Src:      bnet.IPv4(456),
							NextHop:  bnet.IPv4(789),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "No source but no destination",
			source: []netlink.Route{
				netlink.Route{
					Gw:       bnet.IPv4(789).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},

			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx:   bnet.Prefix{},
					Paths: make([]*route.FIBPath, 0),
				},
			},
			expectError: true,
		},
		{
			name: "No source but destination IPv6",
			source: []netlink.Route{
				netlink.Route{
					Dst:      bnet.NewPfx(bnet.IPv6(2001, 0), 48).GetIPNet(),
					Gw:       bnet.IPv6(2001, 123).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},
			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx: bnet.NewPfx(bnet.IPv6(2001, 0), 48),
					Paths: []*route.FIBPath{
						&route.FIBPath{
							Src:      bnet.IPv6(0, 0),
							NextHop:  bnet.IPv6(2001, 123),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Source but no destination IPv6",
			source: []netlink.Route{
				netlink.Route{
					Src:      bnet.IPv6(2001, 456).Bytes(),
					Gw:       bnet.IPv6(2001, 789).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},
			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx: bnet.NewPfx(bnet.IPv6(0, 0), 0),
					Paths: []*route.FIBPath{
						&route.FIBPath{
							Src:      bnet.IPv6(2001, 456),
							NextHop:  bnet.IPv6(2001, 789),
							Protocol: route.ProtoKernel,
							Priority: 1,
							Table:    254,
							Type:     1,
							Kernel:   true,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "no source no destination",
			source: []netlink.Route{
				netlink.Route{
					Gw:       bnet.IPv4(123).Bytes(),
					Protocol: route.ProtoKernel,
					Priority: 1,
					Table:    254,
					Type:     1,
				},
			},
			expected: []route.PrefixPathsPair{
				route.PrefixPathsPair{
					Pfx:   bnet.NewPfx(bnet.IPv4(0), 0),
					Paths: make([]*route.FIBPath, 0),
				},
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		expected := convertNlRouteToFIBPath(test.source, true)
		if test.expectError {
			//assert.Error(t, err)
		} else {
			//assert.NoError(t, err)
			assert.Equalf(t, test.expected[0].Paths, expected[0].Paths, test.name)
			assert.Equalf(t, test.expected[0].Pfx, expected[0].Pfx, test.name)
		}
	}
}
