package adjRIBOutAddPath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNewID(t *testing.T) {
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
			_, err := m.getNewID()
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

func TestReleaseID(t *testing.T) {
	tests := []struct {
		name     string
		pm       *pathIDManager
		release  uint32
		expected *pathIDManager
	}{
		{
			name: "Release existent",
			pm: &pathIDManager{
				ids: map[uint32]struct{}{
					0: struct{}{},
					1: struct{}{},
					2: struct{}{},
				},
				last: 2,
				used: 3,
			},
			release: 1,
			expected: &pathIDManager{
				ids: map[uint32]struct{}{
					0: struct{}{},
					2: struct{}{},
				},
				last: 2,
				used: 2,
			},
		},
		{
			name: "Release non-existent",
			pm: &pathIDManager{
				ids: map[uint32]struct{}{
					0: struct{}{},
					1: struct{}{},
					2: struct{}{},
				},
				last: 2,
				used: 3,
			},
			release: 3,
			expected: &pathIDManager{
				ids: map[uint32]struct{}{
					0: struct{}{},
					1: struct{}{},
					2: struct{}{},
				},
				last: 2,
				used: 3,
			},
		},
	}

	for _, test := range tests {
		test.pm.releaseID(test.release)
		assert.Equalf(t, test.expected, test.pm, "%s", test.name)
	}
}
