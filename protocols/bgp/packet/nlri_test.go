package packet

import (
	"bytes"
	"testing"

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
				IP:     strAddr("192.168.0.0"),
				Pfxlen: 24,
				Next: &NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &NLRI{
						IP:     strAddr("172.16.0.0"),
						Pfxlen: 17,
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
		res, err := decodeNLRIs(buf, uint16(len(test.input)))

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
				IP:     strAddr("192.168.0.0"),
				Pfxlen: 24,
			},
		},
		{
			name: "Valid NRLI #2",
			input: []byte{
				25, 192, 168, 0, 128,
			},
			wantFail: false,
			expected: &NLRI{
				IP:     strAddr("192.168.0.128"),
				Pfxlen: 25,
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
			name:     "Empty input",
			input:    []byte{},
			wantFail: true,
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		res, _, err := decodeNLRI(buf)

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
		res := bytesInAddr(test.input)
		if res != test.expected {
			t.Errorf("Unexpected result for test %q: %d", test.name, res)
		}
	}
}

func TestNLRISerialize(t *testing.T) {
	tests := []struct {
		name     string
		nlri     *NLRI
		expected []byte
	}{
		{
			name: "Test #1",
			nlri: &NLRI{
				IP:     strAddr("1.2.3.0"),
				Pfxlen: 25,
			},
			expected: []byte{25, 1, 2, 3, 0},
		},
		{
			name: "Test #2",
			nlri: &NLRI{
				IP:     strAddr("1.2.3.0"),
				Pfxlen: 24,
			},
			expected: []byte{24, 1, 2, 3},
		},
		{
			name: "Test #3",
			nlri: &NLRI{
				IP:     strAddr("100.200.128.0"),
				Pfxlen: 17,
			},
			expected: []byte{17, 100, 200, 128},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.nlri.serialize(buf)
		res := buf.Bytes()
		assert.Equal(t, test.expected, res)
	}
}

func TestNLRIAddPathSerialize(t *testing.T) {
	tests := []struct {
		name     string
		nlri     *NLRIAddPath
		expected []byte
	}{
		{
			name: "Test #1",
			nlri: &NLRIAddPath{
				PathIdentifier: 100,
				IP:             strAddr("1.2.3.0"),
				Pfxlen:         25,
			},
			expected: []byte{0, 0, 0, 100, 25, 1, 2, 3, 0},
		},
		{
			name: "Test #2",
			nlri: &NLRIAddPath{
				PathIdentifier: 100,
				IP:             strAddr("1.2.3.0"),
				Pfxlen:         24,
			},
			expected: []byte{0, 0, 0, 100, 24, 1, 2, 3},
		},
		{
			name: "Test #3",
			nlri: &NLRIAddPath{
				PathIdentifier: 100,
				IP:             strAddr("100.200.128.0"),
				Pfxlen:         17,
			},
			expected: []byte{0, 0, 0, 100, 17, 100, 200, 128},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.nlri.serialize(buf)
		res := buf.Bytes()
		assert.Equal(t, test.expected, res)
	}
}
