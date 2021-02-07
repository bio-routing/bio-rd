package types

import (
	"testing"

	"github.com/bio-routing/bio-rd/route/api"
	"github.com/stretchr/testify/assert"
)

func TestASPathCompare(t *testing.T) {
	var a *ASPath
	var b *ASPath
	assert.True(t, a.Compare(b))

	a = &ASPath{}
	assert.False(t, a.Compare(b))

	a = nil
	b = &ASPath{}
	assert.False(t, a.Compare(b))

	a = &ASPath{
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 4},
		},
	}

	bEqual := &ASPath{
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 4},
		},
	}

	assert.True(t, a.Compare(bEqual))

	bDifferentLength := &ASPath{
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 5},
		},
	}

	assert.False(t, a.Compare(bDifferentLength))

	bDifferent := &ASPath{
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{1, 2},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 5},
		},
	}

	assert.False(t, a.Compare(bDifferent))
}

func TestASPathGetFirstSequenceSegment(t *testing.T) {
	a := &ASPath{
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 5},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{5, 6},
		},
	}

	actual := a.GetFirstSequenceSegment()
	expected := &ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 5},
	}
	assert.Equal(t, expected, actual)
}

func TestASPathGetFirstSequenceSegmentNoASSequence(t *testing.T) {
	a := &ASPath{
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{2, 2},
		},
	}

	actual := a.GetFirstSequenceSegment()
	assert.Nil(t, actual)
}

func TestASPathGetFirstSequenceSegmentEmpty(t *testing.T) {
	a := &ASPath{}

	actual := a.GetFirstSequenceSegment()
	assert.Nil(t, actual)
}

func TestASPathGetLastSequenceSegment(t *testing.T) {
	a := &ASPath{
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 5},
		},
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{5, 6},
		},
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{3, 5},
		},
	}

	actual := a.GetLastSequenceSegment()
	expected := &ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{5, 6},
	}
	assert.Equal(t, expected, actual)
}

func TestASPathGetLastSequenceSegmentNone(t *testing.T) {
	a := &ASPath{
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{3, 5},
		},
	}

	actual := a.GetLastSequenceSegment()
	assert.Nil(t, actual)
}

func TestASPathSegmentGetFirstASN(t *testing.T) {
	s := ASPathSegment{
		Type: ASSet,
		ASNs: []uint32{3, 5},
	}

	assert.Equal(t, uint32(3), *s.GetFirstASN())
}

func TestASPathSegmentGetFirstASNEmpty(t *testing.T) {
	s := ASPathSegment{
		Type: ASSet,
		ASNs: []uint32{},
	}

	assert.Nil(t, s.GetFirstASN())
}

func TestASPathSegmentGetLastASN(t *testing.T) {
	s := ASPathSegment{
		Type: ASSet,
		ASNs: []uint32{3, 5},
	}

	assert.Equal(t, uint32(5), *s.GetLastASN())
}

func TestASPathSegmentGetLastASNEmpty(t *testing.T) {
	s := ASPathSegment{
		Type: ASSet,
		ASNs: []uint32{},
	}

	assert.Nil(t, s.GetLastASN())
}

func TestASPathSegmentCompareDifferentType(t *testing.T) {
	a := ASPathSegment{
		Type: ASSet,
		ASNs: []uint32{3, 5},
	}

	b := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 5},
	}

	assert.False(t, a.Compare(b))
}

func TestASPathSegmentCompareDifferentLengthASNs(t *testing.T) {
	a := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{1, 3, 5},
	}

	b := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 5},
	}

	assert.False(t, a.Compare(b))
}

func TestASPathSegmentCompareDifferentASNs(t *testing.T) {
	a := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 4},
	}

	b := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 5},
	}

	assert.False(t, a.Compare(b))
}

func TestASPathSegmentCompareSame(t *testing.T) {
	a := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 5},
	}

	b := ASPathSegment{
		Type: ASSequence,
		ASNs: []uint32{3, 5},
	}

	assert.True(t, a.Compare(b))
}

func TestASPathToProto(t *testing.T) {
	a := ASPath{
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 4},
		},
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
	}
	actual := a.ToProto()

	expected := make([]*api.ASPathSegment, 2, 2)
	expected[0] = &api.ASPathSegment{
		AsSequence: true,
		Asns:       []uint32{3, 4},
	}
	expected[1] = &api.ASPathSegment{
		AsSequence: false,
		Asns:       []uint32{1, 2},
	}

	assert.Equal(t, expected, actual)
}

func TestASPathFromProtoASPath(t *testing.T) {
	p := make([]*api.ASPathSegment, 2, 2)
	p[0] = &api.ASPathSegment{
		AsSequence: true,
		Asns:       []uint32{3, 4},
	}
	p[1] = &api.ASPathSegment{
		AsSequence: false,
		Asns:       []uint32{1, 2},
	}

	actual := ASPathFromProtoASPath(p)

	expected := &ASPath{
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 4},
		},
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestASPathString(t *testing.T) {
	tests := []struct {
		name     string
		asPath   *ASPath
		expected string
	}{
		{
			name: "test two Sequences + one Set",
			asPath: &ASPath{
				ASPathSegment{
					Type: ASSequence,
					ASNs: []uint32{3, 4},
				},

				ASPathSegment{
					Type: ASSequence,
					ASNs: []uint32{5, 62},
				},
				ASPathSegment{
					Type: ASSet,
					ASNs: []uint32{100, 2},
				},
			},
			expected: "3 4 5 62 (100 2)",
		}, {
			name: "test one Set",
			asPath: &ASPath{
				ASPathSegment{
					Type: ASSet,
					ASNs: []uint32{1, 2},
				},
			},
			expected: "(1 2)",
		}, {
			name:     "test empty",
			asPath:   &ASPath{},
			expected: "",
		},
	}

	for _, test := range tests {

		actual := test.asPath.String()
		assert.Equal(t, test.expected, actual, test.name)
	}
}

func TestASPathStringNil(t *testing.T) {
	var a *ASPath
	actual := a.String()
	assert.Empty(t, actual)
}

func TestASPathLength(t *testing.T) {
	a := &ASPath{
		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{3, 4},
		},

		ASPathSegment{
			Type: ASSequence,
			ASNs: []uint32{5, 6},
		},
		ASPathSegment{
			Type: ASSet,
			ASNs: []uint32{1, 2},
		},
	}

	actual := a.Length()
	assert.Equal(t, uint16(5), actual)
}
