package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodePathAttrs(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *PathAttribute
	}{
		{
			name: "Valid attribute set",
			input: []byte{
				0,              // Attr. Flags
				1,              // Attr. Type Code
				1,              // Attr. Length
				1,              // EGP
				0,              // Attr. Flags
				3,              // Next Hop
				4,              // Attr. Length
				10, 20, 30, 40, // IP-Address
			},
			wantFail: false,
			expected: &PathAttribute{
				TypeCode: 1,
				Length:   1,
				Value:    uint8(1),
				Next: &PathAttribute{
					TypeCode: 3,
					Length:   4,
					Value:    [4]byte{10, 20, 30, 40},
				},
			},
		},
		{
			name: "Incomplete data",
			input: []byte{
				0, // Attr. Flags
				1, // Attr. Type Code
				1, // Attr. Length
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		res, err := decodePathAttrs(bytes.NewBuffer(test.input), uint16(len(test.input)))

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodePathAttr(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *PathAttribute
	}{
		{
			name: "Valid origin",
			input: []byte{
				0, // Attr. Flags
				1, // Attr. Type Code
				1, // Attr. Length
				1, // EGP
			},
			wantFail: false,
			expected: &PathAttribute{
				Length:         1,
				Optional:       false,
				Transitive:     false,
				Partial:        false,
				ExtendedLength: false,
				TypeCode:       OriginAttr,
				Value:          uint8(1),
			},
		},
		{
			name: "Missing TypeCode",
			input: []byte{
				0, // Attr. Flags
			},
			wantFail: true,
		},
		{
			name: "Missing Length",
			input: []byte{
				0, // Attr. Flags
				1, // Attr. Type Code
			},
			wantFail: true,
		},
		{
			name: "Missing Value ORIGIN",
			input: []byte{
				0, // Attr. Flags
				1, // Attr. Type Code
				1, // Attr. Length
			},
			wantFail: true,
		},
		{
			name: "Missing value AS_PATH",
			input: []byte{
				0, // Attr. Flags
				2, // Attr. Type Code
				8, // Attr. Length
			},
			wantFail: true,
		},
		{
			name: "Missing value NextHop",
			input: []byte{
				0, // Attr. Flags
				3, // Attr. Type Code
				4, // Attr. Length
			},
			wantFail: true,
		},
		{
			name: "Missing value MED",
			input: []byte{
				0, // Attr. Flags
				4, // Attr. Type Code
				4, // Attr. Length
			},
			wantFail: true,
		},
		{
			name: "Missing value LocalPref",
			input: []byte{
				0, // Attr. Flags
				5, // Attr. Type Code
				4, // Attr. Length
			},
			wantFail: true,
		},
		{
			name: "Missing value AGGREGATOR",
			input: []byte{
				0, // Attr. Flags
				7, // Attr. Type Code
				4, // Attr. Length
			},
			wantFail: true,
		},
		{
			name: "Not supported attribute",
			input: []byte{
				0,   // Attr. Flags
				111, // Attr. Type Code
				4,   // Attr. Length
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		res, _, err := decodePathAttr(bytes.NewBuffer(test.input))

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodeOrigin(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *PathAttribute
	}{
		{
			name: "Test #1",
			input: []byte{
				0, // Origin: IGP
			},
			wantFail: false,
			expected: &PathAttribute{
				Value:  uint8(IGP),
				Length: 1,
			},
		},
		{
			name: "Test #2",
			input: []byte{
				1, // Origin: EGP
			},
			wantFail: false,
			expected: &PathAttribute{
				Value:  uint8(EGP),
				Length: 1,
			},
		},
		{
			name: "Test #3",
			input: []byte{
				2, // Origin: INCOMPLETE
			},
			wantFail: false,
			expected: &PathAttribute{
				Value:  uint8(INCOMPLETE),
				Length: 1,
			},
		},
		{
			name:     "Test #4",
			input:    []byte{},
			wantFail: true,
		},
	}

	for _, test := range tests {
		pa := &PathAttribute{
			Length: uint16(len(test.input)),
		}
		err := pa.decodeOrigin(bytes.NewBuffer(test.input))

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		if err != nil {
			continue
		}

		assert.Equal(t, test.expected, pa)
	}
}

func TestDecodeASPath(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "Test #1",
			input: []byte{
				2, // AS_SEQUENCE
				4, // Path Length
				0, 100, 0, 200, 0, 222, 0, 240,
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 10,
				Value: ASPath{
					ASPathSegment{
						Type:  2,
						Count: 4,
						ASNs: []uint32{
							100, 200, 222, 240,
						},
					},
				},
			},
		},
		{
			name: "Test #2",
			input: []byte{
				1, // AS_SEQUENCE
				3, // Path Length
				0, 100, 0, 222, 0, 240,
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 8,
				Value: ASPath{
					ASPathSegment{
						Type:  1,
						Count: 3,
						ASNs: []uint32{
							100, 222, 240,
						},
					},
				},
			},
		},
		{
			name:           "Empty input",
			input:          []byte{},
			explicitLength: 5,
			wantFail:       true,
		},
		{
			name: "Incomplete AS_PATH",
			input: []byte{
				1, // AS_SEQUENCE
				3, // Path Length
				0, 100, 0, 222,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		l := uint16(len(test.input))
		if test.explicitLength != 0 {
			l = test.explicitLength
		}
		pa := &PathAttribute{
			Length: l,
		}
		err := pa.decodeASPath(bytes.NewBuffer(test.input))

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		if err != nil {
			continue
		}

		assert.Equal(t, test.expected, pa)
	}
}

func TestDecodeNextHop(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "Test #1",
			input: []byte{
				10, 20, 30, 40,
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 4,
				Value: [4]byte{
					10, 20, 30, 40,
				},
			},
		},
		{
			name:           "Test #2",
			input:          []byte{},
			explicitLength: 5,
			wantFail:       true,
		},
		{
			name:           "Incomplete IP-Address",
			input:          []byte{10, 20, 30},
			explicitLength: 5,
			wantFail:       true,
		},
	}

	for _, test := range tests {
		l := uint16(len(test.input))
		if test.explicitLength != 0 {
			l = test.explicitLength
		}
		pa := &PathAttribute{
			Length: l,
		}
		err := pa.decodeNextHop(bytes.NewBuffer(test.input))

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		if err != nil {
			continue
		}

		assert.Equal(t, test.expected, pa)
	}
}

func TestDecodeMED(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "Test #1",
			input: []byte{
				0, 0, 3, 232,
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 4,
				Value:  uint32(1000),
			},
		},
		{
			name:           "Test #2",
			input:          []byte{},
			explicitLength: 5,
			wantFail:       true,
		},
	}

	for _, test := range tests {
		l := uint16(len(test.input))
		if test.explicitLength != 0 {
			l = test.explicitLength
		}
		pa := &PathAttribute{
			Length: l,
		}
		err := pa.decodeMED(bytes.NewBuffer(test.input))

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		if err != nil {
			continue
		}

		assert.Equal(t, test.expected, pa)
	}
}

func TestDecodeLocalPref(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "Test #1",
			input: []byte{
				0, 0, 3, 232,
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 4,
				Value:  uint32(1000),
			},
		},
		{
			name:           "Test #2",
			input:          []byte{},
			explicitLength: 5,
			wantFail:       true,
		},
	}

	for _, test := range tests {
		l := uint16(len(test.input))
		if test.explicitLength != 0 {
			l = test.explicitLength
		}
		pa := &PathAttribute{
			Length: l,
		}
		err := pa.decodeLocalPref(bytes.NewBuffer(test.input))

		if test.wantFail {
			if err != nil {
				continue
			}
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, pa)
	}
}

func TestDecodeAggregator(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "Valid aggregator",
			input: []byte{
				0, 222, // ASN
				10, 20, 30, 40, // Aggregator IP
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 6,
				Value: Aggretator{
					ASN:  222,
					Addr: [4]byte{10, 20, 30, 40},
				},
			},
		},
		{
			name: "Incomplete Address",
			input: []byte{
				0, 222, // ASN
				10, 20, // Aggregator IP
			},
			wantFail: true,
		},
		{
			name: "Missing Address",
			input: []byte{
				0, 222, // ASN
			},
			wantFail: true,
		},
		{
			name:     "Empty input",
			input:    []byte{},
			wantFail: true,
		},
	}

	for _, test := range tests {
		l := uint16(len(test.input))
		if test.explicitLength != 0 {
			l = test.explicitLength
		}
		pa := &PathAttribute{
			Length: l,
		}
		err := pa.decodeAggregator(bytes.NewBuffer(test.input))

		if test.wantFail {
			if err != nil {
				continue
			}
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, pa)
	}
}

func TestSetLength(t *testing.T) {
	tests := []struct {
		name             string
		input            []byte
		ExtendedLength   bool
		wantFail         bool
		expected         *PathAttribute
		expectedConsumed int
	}{
		{
			name:           "Valid input",
			ExtendedLength: false,
			input:          []byte{22},
			expected: &PathAttribute{
				ExtendedLength: false,
				Length:         22,
			},
			expectedConsumed: 1,
		},
		{
			name:           "Valid input (extended)",
			ExtendedLength: true,
			input:          []byte{1, 1},
			expected: &PathAttribute{
				ExtendedLength: true,
				Length:         257,
			},
			expectedConsumed: 2,
		},
		{
			name:           "Invalid input",
			ExtendedLength: true,
			input:          []byte{},
			wantFail:       true,
		},
		{
			name:           "Invalid input (extended)",
			ExtendedLength: true,
			input:          []byte{1},
			wantFail:       true,
		},
	}

	for _, test := range tests {
		pa := &PathAttribute{
			ExtendedLength: test.ExtendedLength,
		}
		consumed, err := pa.setLength(bytes.NewBuffer(test.input))

		if test.wantFail {
			if err != nil {
				continue
			}
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, pa)
		assert.Equal(t, test.expectedConsumed, consumed)
	}
}

func TestDecodeUint32(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       uint32
	}{
		{
			name:     "Valid input",
			input:    []byte{0, 0, 1, 1},
			expected: 257,
		},
		{
			name:     "Valid input with additional crap",
			input:    []byte{0, 0, 1, 1, 200},
			expected: 257,
		},
		{
			name:           "Valid input with additional crap and invalid length",
			input:          []byte{0, 0, 1, 1, 200},
			explicitLength: 8,
			wantFail:       true,
		},
		{
			name:     "Invalid input",
			input:    []byte{0, 0, 1},
			wantFail: true,
		},
	}

	for _, test := range tests {
		l := uint16(len(test.input))
		if test.explicitLength > 0 {
			l = test.explicitLength
		}
		pa := &PathAttribute{
			Length: l,
		}
		res, err := pa.decodeUint32(bytes.NewBuffer(test.input))

		if test.wantFail {
			if err != nil {
				continue
			}
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestASPathString(t *testing.T) {
	tests := []struct {
		name     string
		pa       *PathAttribute
		expected string
	}{
		{
			name: "Test #1",
			pa: &PathAttribute{
				Value: &ASPath{
					{
						Type: ASSequence,
						ASNs: []uint32{10, 20, 30},
					},
				},
			},
			expected: "10 20 30",
		},
		{
			name: "Test #2",
			pa: &PathAttribute{
				Value: &ASPath{
					{
						Type: ASSequence,
						ASNs: []uint32{10, 20, 30},
					},
					{
						Type: ASSet,
						ASNs: []uint32{200, 300},
					},
				},
			},
			expected: "10 20 30 (200 300)",
		},
	}

	for _, test := range tests {
		res := test.pa.ASPathString()
		assert.Equal(t, test.expected, res)
	}
}
