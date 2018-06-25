package adjRIBOut

import (
	"testing"

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
			_, err := m.addPath(&route.Path{BGPPath: &route.BGPPath{LocalPref: uint32(i)}})
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
		pm       *pathIDManager
		release  *route.Path
		expected *pathIDManager
		wantFail bool
	}{
		{
			name: "Release existent",
			pm: &pathIDManager{
				ids: map[uint32]uint64{
					0: 1,
					1: 1,
					2: 1,
				},
				idByPath: map[route.BGPPath]uint32{
					route.BGPPath{
						LocalPref: 0,
					}: 0,
					route.BGPPath{
						LocalPref: 1,
					}: 1,
					route.BGPPath{
						LocalPref: 2,
					}: 2,
				},
				last: 2,
				used: 3,
			},
			release: &route.Path{BGPPath: &route.BGPPath{
				LocalPref: 2,
			}},
			expected: &pathIDManager{
				ids: map[uint32]uint64{
					0: 1,
					1: 1,
				},
				idByPath: map[route.BGPPath]uint32{
					route.BGPPath{
						LocalPref: 0,
					}: 0,
					route.BGPPath{
						LocalPref: 1,
					}: 1,
				},
				last: 2,
				used: 2,
			},
		},
		{
			name: "Release non-existent",
			pm: &pathIDManager{
				ids: map[uint32]uint64{
					0: 1,
					1: 1,
					2: 1,
				},
				idByPath: map[route.BGPPath]uint32{
					route.BGPPath{
						LocalPref: 0,
					}: 0,
					route.BGPPath{
						LocalPref: 1,
					}: 1,
					route.BGPPath{
						LocalPref: 2,
					}: 2,
				},
				last: 2,
				used: 3,
			},
			release: &route.Path{BGPPath: &route.BGPPath{
				LocalPref: 4,
			}},
			expected: &pathIDManager{
				ids: map[uint32]uint64{
					0: 1,
					1: 1,
					2: 1,
				},
				idByPath: map[route.BGPPath]uint32{
					route.BGPPath{
						LocalPref: 0,
					}: 0,
					route.BGPPath{
						LocalPref: 1,
					}: 1,
					route.BGPPath{
						LocalPref: 2,
					}: 2,
				},
				last: 2,
				used: 3,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		_, err := test.pm.releasePath(test.release)
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

		assert.Equalf(t, test.expected, test.pm, "%s", test.name)
	}
}

