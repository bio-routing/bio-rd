package adjRIBOut

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestAddPath(t *testing.T) {
	tests := []struct {
		name     string
		maxIDs   uint32
		count    int
		wantFail bool
	}{
		{
			name:     "Out of path IDs",
			maxIDs:   10,
			count:    11,
			wantFail: true,
		},
		{
			name:     "Success",
			maxIDs:   10,
			count:    10,
			wantFail: false,
		},
	}

X:
	for _, test := range tests {
		maxUint32 = test.maxIDs
		m := newPathIDManager()
		for i := 0; i < test.count; i++ {
			_, err := m.addPath(&route.Path{BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					NextHop:   net.IPv4(0).Ptr(),
					Source:    net.IPv4(0).Ptr(),
					LocalPref: uint32(i),
				},
			}})
			if err != nil {
				if test.wantFail {
					continue X
				}

				t.Errorf("Unexpected failure for test %q: %v", test.name, err)
				continue X
			}
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}
	}

}

func TestReleasePath(t *testing.T) {
	tests := []struct {
		name     string
		adds     []*route.Path
		release  *route.Path
		expected []*route.Path
		wantFail bool
	}{
		{
			name: "Release existent",
			adds: []*route.Path{
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 0,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 1,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 2,
						},
					},
				},
			},
			release: &route.Path{BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					Source:    net.IPv4(0).Ptr(),
					NextHop:   net.IPv4(0).Ptr(),
					LocalPref: 2,
				},
			}},
			expected: []*route.Path{
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 0,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 1,
						},
					},
				},
			},
		},
		{
			name: "Release non-existent",
			adds: []*route.Path{
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 0,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 1,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 2,
						},
					},
				},
			},
			release: &route.Path{BGPPath: &route.BGPPath{
				BGPPathA: &route.BGPPathA{
					Source:    net.IPv4(0).Ptr(),
					NextHop:   net.IPv4(0).Ptr(),
					LocalPref: 5,
				},
			}},
			expected: []*route.Path{
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 0,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 1,
						},
					},
				},
				{
					BGPPath: &route.BGPPath{
						BGPPathA: &route.BGPPathA{
							Source:    net.IPv4(0).Ptr(),
							NextHop:   net.IPv4(0).Ptr(),
							LocalPref: 2,
						},
					},
				},
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		pm := newPathIDManager()
		for _, add := range test.adds {
			pm.addPath(add)
		}

		_, err := pm.releasePath(test.release)
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		expectedPM := newPathIDManager()
		for _, x := range test.expected {
			expectedPM.addPath(x)
		}
		expectedPM.last++

		assert.Equalf(t, expectedPM, pm, "%s", test.name)
	}
}
