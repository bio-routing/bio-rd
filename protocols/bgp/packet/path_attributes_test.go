package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
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
					Value:    strAddr("10.20.30.40"),
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
		res, err := decodePathAttrs(bytes.NewBuffer(test.input), uint16(len(test.input)), &types.Options{})

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
			continue
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		assert.Equalf(t, test.expected, res, "%s", test.name)
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
		res, _, err := decodePathAttr(bytes.NewBuffer(test.input), &types.Options{})

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
		use4OctetASNs  bool
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
				Value: types.ASPath{
					types.ASPathSegment{
						Type: 2,
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
				Value: types.ASPath{
					types.ASPathSegment{
						Type: 1,
						ASNs: []uint32{
							100, 222, 240,
						},
					},
				},
			},
		},
		{
			name: "32 bit ASNs in AS_PATH",
			input: []byte{
				1, // AS_SEQUENCE
				3, // Path Length
				0, 0, 0, 100, 0, 0, 0, 222, 0, 0, 0, 240,
			},
			wantFail:      false,
			use4OctetASNs: true,
			expected: &PathAttribute{
				Length: 14,
				Value: types.ASPath{
					types.ASPathSegment{
						Type: 1,
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
		t.Run(test.name, func(t *testing.T) {
			l := uint16(len(test.input))
			if test.explicitLength != 0 {
				l = test.explicitLength
			}
			pa := &PathAttribute{
				Length: l,
			}

			asnLength := uint8(2)
			if test.use4OctetASNs {
				asnLength = 4
			}

			err := pa.decodeASPath(bytes.NewBuffer(test.input), asnLength)

			if test.wantFail && err == nil {
				t.Errorf("Expected error did not happen for test %q", test.name)
			}

			if !test.wantFail && err != nil {
				t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			}

			if err != nil {
				return
			}

			assert.Equal(t, test.expected, pa)
		})
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
				Value:  strAddr("10.20.30.40"),
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
					Addr: strAddr("10.20.30.40"),
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

func TestDecodeLargeCommunity(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "two valid large communities",
			input: []byte{
				0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 0, 0, 0, 5, 0, 0, 0, 6, // (1, 2, 3), (4, 5, 6)
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 24,
				Value: []types.LargeCommunity{
					{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           3,
					},
					{
						GlobalAdministrator: 4,
						DataPart1:           5,
						DataPart2:           6,
					},
				},
			},
		},
		{
			name:     "Empty input",
			input:    []byte{},
			wantFail: false,
			expected: &PathAttribute{
				Length: 0,
				Value:  []types.LargeCommunity{},
			},
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
		err := pa.decodeLargeCommunities(bytes.NewBuffer(test.input))

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

func TestDecodeCommunity(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		wantFail       bool
		explicitLength uint16
		expected       *PathAttribute
	}{
		{
			name: "two valid communities",
			input: []byte{
				0, 2, 0, 8, 1, 0, 4, 1, // (2,8), (256,1025)
			},
			wantFail: false,
			expected: &PathAttribute{
				Length: 8,
				Value: []uint32{
					131080, 16778241,
				},
			},
		},
		{
			name:     "Empty input",
			input:    []byte{},
			wantFail: false,
			expected: &PathAttribute{
				Length: 0,
				Value:  []uint32{},
			},
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
		err := pa.decodeCommunities(bytes.NewBuffer(test.input))

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
		err := pa.decodeUint32(bytes.NewBuffer(test.input), "test")

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

		assert.Equal(t, test.expected, pa.Value)
	}
}

func TestSetOptional(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected uint8
	}{
		{
			name:     "Test #1",
			input:    0,
			expected: 128,
		},
	}

	for _, test := range tests {
		res := setOptional(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}

func TestSetTransitive(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected uint8
	}{
		{
			name:     "Test #1",
			input:    0,
			expected: 64,
		},
	}

	for _, test := range tests {
		res := setTransitive(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}

func TestSetPartial(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected uint8
	}{
		{
			name:     "Test #1",
			input:    0,
			expected: 32,
		},
	}

	for _, test := range tests {
		res := setPartial(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}

func TestSetExtendedLength(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected uint8
	}{
		{
			name:     "Test #1",
			input:    0,
			expected: 16,
		},
	}

	for _, test := range tests {
		res := setExtendedLength(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}

func TestSerializeOrigin(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: OriginAttr,
				Value:    uint8(0), // IGP
			},
			expectedLen: 4,
			expected:    []byte{64, 1, 1, 0},
		},
		{
			name: "Test #2",
			input: &PathAttribute{
				TypeCode: OriginAttr,
				Value:    uint8(2), // INCOMPLETE
			},
			expectedLen: 4,
			expected:    []byte{64, 1, 1, 2},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		n := test.input.serializeOrigin(buf)
		if test.expectedLen != n {
			t.Errorf("Unexpected length for test %q: %d", test.name, n)
			continue
		}

		assert.Equal(t, test.expected, buf.Bytes())
	}
}

func TestSerializeNextHop(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: NextHopAttr,
				Value:    strAddr("100.110.120.130"),
			},
			expected:    []byte{64, 3, 4, 100, 110, 120, 130},
			expectedLen: 7,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		n := test.input.serializeNextHop(buf)
		if n != test.expectedLen {
			t.Errorf("Unexpected length for test %q: %d", test.name, n)
			continue
		}

		assert.Equal(t, test.expected, buf.Bytes())
	}
}

func TestSerializeMED(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: MEDAttr,
				Value:    uint32(1000),
			},
			expected: []byte{
				128,          // Attribute flags
				4,            // Type
				4,            // Length
				0, 0, 3, 232, // Value = 1000
			},
			expectedLen: 7,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		n := test.input.serializeMED(buf)
		if n != test.expectedLen {
			t.Errorf("Unexpected length for test %q: %d", test.name, n)
			continue
		}

		assert.Equal(t, test.expected, buf.Bytes())
	}
}

func TestSerializeLocalPref(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: LocalPrefAttr,
				Value:    uint32(1000),
			},
			expected: []byte{
				64,           // Attribute flags
				5,            // Type
				4,            // Length
				0, 0, 3, 232, // Value = 1000
			},
			expectedLen: 7,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		n := test.input.serializeLocalpref(buf)
		if n != test.expectedLen {
			t.Errorf("Unexpected length for test %q: %d", test.name, n)
			continue
		}

		assert.Equal(t, test.expected, buf.Bytes())
	}
}

func TestSerializeAtomicAggregate(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: AtomicAggrAttr,
			},
			expected: []byte{
				64, // Attribute flags
				6,  // Type
				0,  // Length
			},
			expectedLen: 3,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		n := test.input.serializeAtomicAggregate(buf)
		if n != test.expectedLen {
			t.Errorf("Unexpected length for test %q: %d", test.name, n)
			continue
		}

		assert.Equal(t, test.expected, buf.Bytes())
	}
}

func TestSerializeAggregator(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: AggregatorAttr,
				Value:    uint16(174),
			},
			expected: []byte{
				192,    // Attribute flags
				7,      // Type
				2,      // Length
				0, 174, // Value = 174
			},
			expectedLen: 5,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		n := test.input.serializeAggregator(buf)
		if n != test.expectedLen {
			t.Errorf("Unexpected length for test %q: %d", test.name, n)
			continue
		}

		assert.Equal(t, test.expected, buf.Bytes())
	}
}

func TestSerializeASPath(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
		use32BitASN bool
	}{
		{
			name: "Test #1",
			input: &PathAttribute{
				TypeCode: ASPathAttr,
				Value: types.ASPath{
					{
						Type: 2, // Sequence
						ASNs: []uint32{
							100, 200, 210,
						},
					},
				},
			},
			expected: []byte{
				64,     // Attribute flags
				2,      // Type
				8,      // Length
				2,      // AS_SEQUENCE
				3,      // ASN count
				0, 100, // ASN 100
				0, 200, // ASN 200
				0, 210, // ASN 210
			},
			expectedLen: 10,
		},
		{
			name: "32bit ASN",
			input: &PathAttribute{
				TypeCode: ASPathAttr,
				Value: types.ASPath{
					{
						Type: 2, // Sequence
						ASNs: []uint32{
							100, 200, 210,
						},
					},
				},
			},
			expected: []byte{
				64,           // Attribute flags
				2,            // Type
				14,           // Length
				2,            // AS_SEQUENCE
				3,            // ASN count
				0, 0, 0, 100, // ASN 100
				0, 0, 0, 200, // ASN 200
				0, 0, 0, 210, // ASN 210
			},
			expectedLen: 16,
			use32BitASN: true,
		},
	}

	t.Parallel()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			opt := &types.Options{
				Supports4OctetASN: test.use32BitASN,
			}
			n := test.input.serializeASPath(buf, opt)
			if n != test.expectedLen {
				t.Fatalf("Unexpected length for test %q: %d", test.name, n)
			}

			assert.Equal(t, test.expected, buf.Bytes())
		})
	}
}

func TestSerializeLargeCommunities(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "2 large communities",
			input: &PathAttribute{
				TypeCode: LargeCommunitiesAttr,
				Value: []types.LargeCommunity{
					{
						GlobalAdministrator: 1,
						DataPart1:           2,
						DataPart2:           3,
					},
					{
						GlobalAdministrator: 4,
						DataPart1:           5,
						DataPart2:           6,
					},
				},
			},
			expected: []byte{
				0xe0,                                                                   // Attribute flags
				32,                                                                     // Type
				24,                                                                     // Length
				0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 0, 0, 0, 5, 0, 0, 0, 6, // Communities (1, 2, 3), (4, 5, 6)
			},
			expectedLen: 24,
		},
		{
			name: "empty list of communities",
			input: &PathAttribute{
				TypeCode: LargeCommunitiesAttr,
				Value:    []types.LargeCommunity{},
			},
			expected:    []byte{},
			expectedLen: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			n := test.input.serializeLargeCommunities(buf)
			if n != test.expectedLen {
				t.Fatalf("Unexpected length for test %q: %d", test.name, n)
			}

			assert.Equal(t, test.expected, buf.Bytes())
		})
	}
}

func TestSerializeCommunities(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint8
	}{
		{
			name: "2 communities",
			input: &PathAttribute{
				TypeCode: LargeCommunitiesAttr,
				Value: []uint32{
					131080, 16778241,
				},
			},
			expected: []byte{
				0xe0,                   // Attribute flags
				8,                      // Type
				8,                      // Length
				0, 2, 0, 8, 1, 0, 4, 1, // Communities (2,8), (256,1025)
			},
			expectedLen: 8,
		},
		{
			name: "empty list of communities",
			input: &PathAttribute{
				TypeCode: CommunitiesAttr,
				Value:    []uint32{},
			},
			expected:    []byte{},
			expectedLen: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(te *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			n := test.input.serializeCommunities(buf)
			if n != test.expectedLen {
				t.Fatalf("Unexpected length for test %q: %d", test.name, n)
			}

			assert.Equal(t, test.expected, buf.Bytes())
		})
	}
}

func TestSerializeUnknownAttribute(t *testing.T) {
	tests := []struct {
		name        string
		input       *PathAttribute
		expected    []byte
		expectedLen uint16
	}{
		{
			name: "Arbritary attribute",
			input: &PathAttribute{
				TypeCode:   200,
				Value:      []byte{1, 2, 3, 4},
				Transitive: true,
			},
			expected: []byte{
				64,         // Attribute flags
				200,        // Type
				4,          // Length
				1, 2, 3, 4, // Payload
			},
			expectedLen: 6,
		},
		{
			name: "Extended length",
			input: &PathAttribute{
				TypeCode:       200,
				Value:          make([]byte, 256),
				Transitive:     true,
				ExtendedLength: true,
			},
			expected: []byte{
				80,   // Attribute flags
				200,  // Type
				1, 0, // Length
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // Payload
			},
			expectedLen: 258,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			n := test.input.serializeUnknownAttribute(buf)

			assert.Equal(t, test.expectedLen, n)
			assert.Equal(t, test.expected, buf.Bytes())
		})
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name     string
		msg      *BGPUpdate
		expected []byte
		wantFail bool
	}{
		{
			name: "Withdraw only",
			msg: &BGPUpdate{
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("100.110.120.0"),
					Pfxlen: 24,
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 27, // Length
				2,    // Msg Type
				0, 4, // Withdrawn Routes Length
				24, 100, 110, 120, // NLRI
				0, 0, // Total Path Attribute Length
			},
		},
		{
			name: "NLRI only",
			msg: &BGPUpdate{
				NLRI: &NLRI{
					IP:     strAddr("100.110.128.0"),
					Pfxlen: 17,
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 27, // Length
				2,    // Msg Type
				0, 0, // Withdrawn Routes Length
				0, 0, // Total Path Attribute Length
				17, 100, 110, 128, // NLRI
			},
		},
		{
			name: "Path Attributes only",
			msg: &BGPUpdate{
				PathAttributes: &PathAttribute{
					Optional:   true,
					Transitive: true,
					TypeCode:   OriginAttr,
					Value:      uint8(0), // IGP
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 27, // Length
				2,    // Msg Type
				0, 0, // Withdrawn Routes Length
				0, 4, // Total Path Attribute Length
				64, // Attr. Flags
				1,  // Attr. Type Code
				1,  // Length
				0,  // Value
			},
		},
		{
			name: "Full test",
			msg: &BGPUpdate{
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				PathAttributes: &PathAttribute{
					TypeCode: OriginAttr,
					Value:    uint8(0),
					Next: &PathAttribute{
						TypeCode: ASPathAttr,
						Value: types.ASPath{
							{
								Type: 2,
								ASNs: []uint32{100, 155, 200},
							},
							{
								Type: 1,
								ASNs: []uint32{10, 20},
							},
						},
						Next: &PathAttribute{
							TypeCode: NextHopAttr,
							Value:    strAddr("10.20.30.40"),
							Next: &PathAttribute{
								TypeCode: MEDAttr,
								Value:    uint32(100),
								Next: &PathAttribute{
									TypeCode: LocalPrefAttr,
									Value:    uint32(500),
									Next: &PathAttribute{
										TypeCode: AtomicAggrAttr,
										Next: &PathAttribute{
											TypeCode: AggregatorAttr,
											Value:    uint16(200),
										},
									},
								},
							},
						},
					},
				},
				NLRI: &NLRI{
					IP:     strAddr("8.8.8.0"),
					Pfxlen: 24,
					Next: &NLRI{
						IP:     strAddr("185.65.240.0"),
						Pfxlen: 22,
					},
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 86, // Length
				2, // Msg Type

				// Withdraws
				0, 5, // Withdrawn Routes Length
				8, 10, // Withdraw 10/8
				16, 192, 168, // Withdraw 192.168/16

				0, 50, // Total Path Attribute Length

				// ORIGIN
				64, // Attr. Flags
				1,  // Attr. Type Code
				1,  // Length
				0,  // Value
				// ASPath
				64,                     // Attr. Flags
				2,                      // Attr. Type Code
				14,                     // Attr. Length
				2,                      // Path Segment Type = AS_SEQUENCE
				3,                      // Path Segment Length
				0, 100, 0, 155, 0, 200, // ASNs
				1,            // Path Segment Type = AS_SET
				2,            // Path Segment Type = AS_SET
				0, 10, 0, 20, // ASNs
				// Next Hop
				64,             // Attr. Flags
				3,              // Attr. Type Code
				4,              // Length
				10, 20, 30, 40, // Next Hop Address
				// MED
				128,          // Attr. Flags
				4,            // Attr Type Code
				4,            // Length
				0, 0, 0, 100, // MED = 100
				// LocalPref
				64,           // Attr. Flags
				5,            // Attr. Type Code
				4,            // Length
				0, 0, 1, 244, // Localpref
				// Atomic Aggregate
				64, // Attr. Flags
				6,  // Attr. Type Code
				0,  // Length
				// Aggregator
				192,    // Attr. Flags
				7,      // Attr. Type Code
				2,      // Length
				0, 200, // Aggregator ASN = 200

				// NLRI
				24, 8, 8, 8, // 8.8.8.0/24
				22, 185, 65, 240, // 185.65.240.0/22
			},
		},
	}

	for _, test := range tests {
		opt := &types.Options{
			AddPathRX: false,
		}
		res, err := test.msg.SerializeUpdate(opt)
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

		assert.Equalf(t, test.expected, res, "%s", test.name)
	}
}

func TestSerializeAddPath(t *testing.T) {
	tests := []struct {
		name     string
		msg      *BGPUpdate
		expected []byte
		wantFail bool
	}{
		{
			name: "Withdraw only",
			msg: &BGPUpdate{
				WithdrawnRoutes: &NLRI{
					PathIdentifier: 257,
					IP:             strAddr("100.110.120.0"),
					Pfxlen:         24,
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 31, // Length
				2,    // Msg Type
				0, 8, // Withdrawn Routes Length
				0, 0, 1, 1, // Path Identifier
				24, 100, 110, 120, // NLRI
				0, 0, // Total Path Attribute Length
			},
		},
		{
			name: "NLRI only",
			msg: &BGPUpdate{
				NLRI: &NLRI{
					PathIdentifier: 257,
					IP:             strAddr("100.110.128.0"),
					Pfxlen:         17,
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 31, // Length
				2,    // Msg Type
				0, 0, // Withdrawn Routes Length
				0, 0, // Total Path Attribute Length
				0, 0, 1, 1, // Path Identifier
				17, 100, 110, 128, // NLRI
			},
		},
		{
			name: "Path Attributes only",
			msg: &BGPUpdate{
				PathAttributes: &PathAttribute{
					Optional:   true,
					Transitive: true,
					TypeCode:   OriginAttr,
					Value:      uint8(0), // IGP
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 27, // Length
				2,    // Msg Type
				0, 0, // Withdrawn Routes Length
				0, 4, // Total Path Attribute Length
				64, // Attr. Flags
				1,  // Attr. Type Code
				1,  // Length
				0,  // Value
			},
		},
		{
			name: "Full test",
			msg: &BGPUpdate{
				WithdrawnRoutes: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
				PathAttributes: &PathAttribute{
					TypeCode: OriginAttr,
					Value:    uint8(0),
					Next: &PathAttribute{
						TypeCode: ASPathAttr,
						Value: types.ASPath{
							{
								Type: 2,
								ASNs: []uint32{100, 155, 200},
							},
							{
								Type: 1,
								ASNs: []uint32{10, 20},
							},
						},
						Next: &PathAttribute{
							TypeCode: NextHopAttr,
							Value:    strAddr("10.20.30.40"),
							Next: &PathAttribute{
								TypeCode: MEDAttr,
								Value:    uint32(100),
								Next: &PathAttribute{
									TypeCode: LocalPrefAttr,
									Value:    uint32(500),
									Next: &PathAttribute{
										TypeCode: AtomicAggrAttr,
										Next: &PathAttribute{
											TypeCode: AggregatorAttr,
											Value:    uint16(200),
										},
									},
								},
							},
						},
					},
				},
				NLRI: &NLRI{
					IP:     strAddr("8.8.8.0"),
					Pfxlen: 24,
					Next: &NLRI{
						IP:     strAddr("185.65.240.0"),
						Pfxlen: 22,
					},
				},
			},
			expected: []byte{
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0, 102, // Length
				2, // Msg Type

				// Withdraws
				0, 13, // Withdrawn Routes Length
				0, 0, 0, 0, // Path Identifier
				8, 10, // Withdraw 10/8
				0, 0, 0, 0, // Path Identifier
				16, 192, 168, // Withdraw 192.168/16

				0, 50, // Total Path Attribute Length

				// ORIGIN
				64, // Attr. Flags
				1,  // Attr. Type Code
				1,  // Length
				0,  // Value
				// ASPath
				64,                     // Attr. Flags
				2,                      // Attr. Type Code
				14,                     // Attr. Length
				2,                      // Path Segment Type = AS_SEQUENCE
				3,                      // Path Segment Length
				0, 100, 0, 155, 0, 200, // ASNs
				1,            // Path Segment Type = AS_SET
				2,            // Path Segment Type = AS_SET
				0, 10, 0, 20, // ASNs
				// Next Hop
				64,             // Attr. Flags
				3,              // Attr. Type Code
				4,              // Length
				10, 20, 30, 40, // Next Hop Address
				// MED
				128,          // Attr. Flags
				4,            // Attr Type Code
				4,            // Length
				0, 0, 0, 100, // MED = 100
				// LocalPref
				64,           // Attr. Flags
				5,            // Attr. Type Code
				4,            // Length
				0, 0, 1, 244, // Localpref
				// Atomic Aggregate
				64, // Attr. Flags
				6,  // Attr. Type Code
				0,  // Length
				// Aggregator
				192,    // Attr. Flags
				7,      // Attr. Type Code
				2,      // Length
				0, 200, // Aggregator ASN = 200

				// NLRI
				0, 0, 0, 0, // Path Identifier
				24, 8, 8, 8, // 8.8.8.0/24
				0, 0, 0, 0, // Path Identifier
				22, 185, 65, 240, // 185.65.240.0/22
			},
		},
	}

	for _, test := range tests {
		opt := &types.Options{
			AddPathRX: true,
		}
		res, err := test.msg.SerializeUpdate(opt)
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

		assert.Equalf(t, test.expected, res, "%s", test.name)
	}
}

func TestFourBytesToUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    [4]byte
		expected uint32
	}{
		{
			name:     "Test #1",
			input:    [4]byte{0, 0, 0, 200},
			expected: 200,
		},
		{
			name:     "Test #2",
			input:    [4]byte{1, 0, 0, 200},
			expected: 16777416,
		},
	}

	for _, test := range tests {
		res := fourBytesToUint32(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: Got: %d Want: %d", test.name, res, test.expected)
		}
	}
}
