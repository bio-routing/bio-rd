package packet

import (
	"bytes"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestDecodeNLRIs(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *NLRI
	}{
		{
			name: "Valid NRLI #1",
			input: []byte{
				24, 192, 168, 0,
				8, 10,
				17, 172, 16, 0,
			},
			wantFail: false,
			expected: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 0).Dedup(), 24).Dedup(),
				Next: &NLRI{
					Prefix: bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0).Dedup(), 8).Dedup(),
					Next: &NLRI{
						Prefix: bnet.NewPfx(bnet.IPv4FromOctets(172, 16, 0, 0).Dedup(), 17).Dedup(),
					},
				},
			},
		},
		{
			name: "Invalid NRLI #1",
			input: []byte{
				24, 192, 168, 0,
				8, 10,
				17, 172, 16,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, err := decodeNLRIs(buf, uint16(len(test.input)), IPv4AFI, false)

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodeNLRIv6(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		addPath  bool
		wantFail bool
		expected *NLRI
	}{
		{
			name: "IPv6 default",
			input: []byte{
				0,
			},
			wantFail: false,
			expected: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv6FromBlocks(0, 0, 0, 0, 0, 0, 0, 0).Dedup(), 0).Dedup(),
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, _, err := decodeNLRI(buf, IPv6AFI, test.addPath)

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestDecodeNLRI(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		addPath  bool
		wantFail bool
		expected *NLRI
	}{
		{
			name: "Valid NRLI #1",
			input: []byte{
				24, 192, 168, 0,
			},
			wantFail: false,
			expected: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 0).Dedup(), 24).Dedup(),
			},
		},
		{
			name: "Valid NRLI #2",
			input: []byte{
				25, 192, 168, 0, 128,
			},
			wantFail: false,
			expected: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 128).Dedup(), 25).Dedup(),
			},
		},
		{
			name: "Incomplete NLRI #1",
			input: []byte{
				25, 192, 168, 0,
			},
			wantFail: true,
		},
		{
			name: "Incomplete NLRI #2",
			input: []byte{
				25,
			},
			wantFail: true,
		},

		{
			name: "Valid NRLI #1 add path",
			input: []byte{
				0, 0, 0, 10, 24, 192, 168, 0,
			},
			addPath:  true,
			wantFail: false,
			expected: &NLRI{
				PathIdentifier: 10,
				Prefix:         bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 0).Dedup(), 24).Dedup(),
			},
		},
		{
			name: "Valid NRLI #2 add path",
			input: []byte{
				0, 0, 1, 0, 25, 192, 168, 0, 128,
			},
			addPath:  true,
			wantFail: false,
			expected: &NLRI{
				PathIdentifier: 256,
				Prefix:         bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 128).Dedup(), 25).Dedup(),
			},
		},
		{
			name: "Incomplete path Identifier",
			input: []byte{
				0, 0, 0,
			},
			addPath:  true,
			wantFail: true,
		},
		{
			name: "Incomplete NLRI #1  add path",
			input: []byte{
				0, 0, 1, 0, 25, 192, 168, 0,
			},
			addPath:  true,
			wantFail: true,
		},
		{
			name: "Incomplete NLRI #2  add path",
			input: []byte{
				0, 0, 1, 0, 25,
			},
			addPath:  true,
			wantFail: true,
		},

		{
			name:     "Empty input",
			input:    []byte{},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, _, err := decodeNLRI(buf, IPv4AFI, test.addPath)

		if test.wantFail && err == nil {
			t.Errorf("Expected error did not happen for test %q", test.name)
		}

		if !test.wantFail && err != nil {
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
		}

		assert.Equal(t, test.expected, res)
	}
}

func TestBytesInAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		expected uint8
	}{
		{
			name:     "Test #1",
			input:    24,
			expected: 3,
		},
		{
			name:     "Test #2",
			input:    25,
			expected: 4,
		},
		{
			name:     "Test #3",
			input:    32,
			expected: 4,
		},
		{
			name:     "Test #4",
			input:    0,
			expected: 0,
		},
		{
			name:     "Test #5",
			input:    9,
			expected: 2,
		},
	}

	for _, test := range tests {
		res := BytesInAddr(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}

func TestNLRISerialize(t *testing.T) {
	tests := []struct {
		name     string
		nlri     *NLRI
		addPath  bool
		expected []byte
	}{
		{
			name: "Test #1",
			nlri: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv4FromOctets(1, 2, 3, 0).Dedup(), 25).Dedup(),
			},
			expected: []byte{25, 1, 2, 3, 0},
		},
		{
			name: "Test #2",
			nlri: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv4FromOctets(1, 2, 3, 0).Dedup(), 24).Dedup(),
			},
			expected: []byte{24, 1, 2, 3},
		},
		{
			name: "Test #3",
			nlri: &NLRI{
				Prefix: bnet.NewPfx(bnet.IPv4FromOctets(100, 200, 128, 0).Dedup(), 17).Dedup(),
			},
			expected: []byte{17, 100, 200, 128},
		},
		{
			name: "with add-path #1",
			nlri: &NLRI{
				PathIdentifier: 100,
				Prefix:         bnet.NewPfx(bnet.IPv4FromOctets(1, 2, 3, 0).Dedup(), 25).Dedup(),
			},
			addPath:  true,
			expected: []byte{0, 0, 0, 100, 25, 1, 2, 3, 0},
		},
		{
			name: "with add-path #2",
			nlri: &NLRI{
				PathIdentifier: 100,
				Prefix:         bnet.NewPfx(bnet.IPv4FromOctets(1, 2, 3, 0).Dedup(), 24).Dedup(),
			},
			addPath:  true,
			expected: []byte{0, 0, 0, 100, 24, 1, 2, 3},
		},
		{
			name: "with add-path #3",
			nlri: &NLRI{
				PathIdentifier: 100,
				Prefix:         bnet.NewPfx(bnet.IPv4FromOctets(100, 200, 128, 0).Dedup(), 17).Dedup(),
			},
			addPath:  true,
			expected: []byte{0, 0, 0, 100, 17, 100, 200, 128},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.nlri.serialize(buf, test.addPath)
		res := buf.Bytes()
		assert.Equal(t, test.expected, res)
	}
}
