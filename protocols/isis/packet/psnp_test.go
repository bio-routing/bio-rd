package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewPSNPs(t *testing.T) {
	tests := []struct {
		name         string
		sourceID     types.SystemID
		lspEntries   []LSPEntry
		maxPDULength int
		expected     []PSNP
	}{
		{
			name:     "All in one packet",
			sourceID: types.SystemID{10, 20, 30, 40, 50, 60},
			lspEntries: []LSPEntry{
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 123,
					},
				},
			},
			maxPDULength: 1492,
			expected: []PSNP{
				{
					PDULength: 24,
					SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
					LSPEntries: []LSPEntry{
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 123,
							},
						},
					},
				},
			},
		},
		{
			name:     "2 packets",
			sourceID: types.SystemID{10, 20, 30, 40, 50, 60},
			lspEntries: []LSPEntry{
				{
					SequenceNumber:    1001,
					RemainingLifetime: 2001,
					LSPChecksum:       112,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
				},
			},
			maxPDULength: 24,
			expected: []PSNP{
				{
					PDULength: 24,
					SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
					LSPEntries: []LSPEntry{
						{
							SequenceNumber:    1001,
							RemainingLifetime: 2001,
							LSPChecksum:       112,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 200,
							},
						},
					},
				},
				{
					PDULength: 24,
					SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
					LSPEntries: []LSPEntry{
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 100,
							},
						},
					},
				},
			},
		},
		{
			name:     "2 packets with odd length",
			sourceID: types.SystemID{10, 20, 30, 40, 50, 60},
			lspEntries: []LSPEntry{
				{
					SequenceNumber:    1001,
					RemainingLifetime: 2001,
					LSPChecksum:       112,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
				},
			},
			maxPDULength: 28,
			expected: []PSNP{
				{
					PDULength: 24,
					SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
					LSPEntries: []LSPEntry{
						{
							SequenceNumber:    1001,
							RemainingLifetime: 2001,
							LSPChecksum:       112,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 200,
							},
						},
					},
				},
				{
					PDULength: 24,
					SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
					LSPEntries: []LSPEntry{
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 100,
							},
						},
					},
				},
			},
		},
		{
			name:     "2 LSPEntries, 1 packet",
			sourceID: types.SystemID{10, 20, 30, 40, 50, 60},
			lspEntries: []LSPEntry{
				{
					SequenceNumber:    1001,
					RemainingLifetime: 2001,
					LSPChecksum:       112,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
				},
			},
			maxPDULength: 40,
			expected: []PSNP{
				{
					PDULength: 40,
					SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
					LSPEntries: []LSPEntry{
						{
							SequenceNumber:    1001,
							RemainingLifetime: 2001,
							LSPChecksum:       112,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 200,
							},
						},
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 100,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		psnps := NewPSNPs(test.sourceID, test.lspEntries, test.maxPDULength)
		assert.Equalf(t, test.expected, psnps, "Test %q", test.name)
	}
}

func TestPSNPSerialize(t *testing.T) {
	tests := []struct {
		name     string
		psnp     PSNP
		expected []byte
	}{
		{
			name: "Test #1",
			psnp: PSNP{
				PDULength: 100,
				SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
				LSPEntries: []LSPEntry{
					{
						SequenceNumber:    123,
						RemainingLifetime: 255,
						LSPChecksum:       111,
						LSPID: LSPID{
							SystemID:     types.SystemID{10, 20, 30, 40, 50, 61},
							PseudonodeID: 11,
						},
					},
				},
			},
			expected: []byte{
				0, 100,
				10, 20, 30, 40, 50, 60,
				0, 0, 0, 123,
				0, 255,
				0, 111,
				10, 20, 30, 40, 50, 61, 0, 11,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.psnp.Serialize(buf)
		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestDecodePSNP(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *PSNP
	}{
		{
			name: "Incomplete PSNP",
			input: []byte{
				0, 24, // Length
				10, 20, 30, 40, 50, 60, // Source ID
			},
			wantFail: true,
		},
		{
			name: "Incomplete PSNP LSPEntry",
			input: []byte{
				0, 24, // Length
				10, 20, 30, 40, 50, 60, // Source ID
				0, 0, 0, 20, // Sequence Number
			},
			wantFail: true,
		},
		{
			name: "PSNP with one LSPEntry",
			input: []byte{
				0, 24, // Length
				10, 20, 30, 40, 50, 60, // Source ID
				0, 0, 0, 20, // Sequence Number
				1, 0, // Remaining Lifetime
				2, 0, // Checksum
				11, 22, 33, 44, 55, 66, // SystemID
				0, 20, // Pseudonode ID
			},
			wantFail: false,
			expected: &PSNP{
				PDULength: 24,
				SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
				LSPEntries: []LSPEntry{
					{
						SequenceNumber:    20,
						RemainingLifetime: 256,
						LSPChecksum:       512,
						LSPID: LSPID{
							SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
							PseudonodeID: 20,
						},
					},
				},
			},
		},
		{
			name: "PSNP with two LSPEntries",
			input: []byte{
				0, 40, // Length
				10, 20, 30, 40, 50, 60, // Source ID
				0, 0, 0, 20, // Sequence Number
				1, 0, // Remaining Lifetime
				2, 0, // Checksum
				11, 22, 33, 44, 55, 66, // SystemID
				0, 20, // Pseudonode ID
				0, 0, 0, 21, // Sequence Number
				2, 0, // Remaining Lifetime
				2, 0, // Checksum
				11, 22, 33, 44, 55, 67, // SystemID
				0, 21, // Pseudonode ID
			},
			wantFail: false,
			expected: &PSNP{
				PDULength: 40,
				SourceID:  types.SystemID{10, 20, 30, 40, 50, 60},
				LSPEntries: []LSPEntry{
					{
						SequenceNumber:    20,
						RemainingLifetime: 256,
						LSPChecksum:       512,
						LSPID: LSPID{
							SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
							PseudonodeID: 20,
						},
					},
					{
						SequenceNumber:    21,
						RemainingLifetime: 512,
						LSPChecksum:       512,
						LSPID: LSPID{
							SystemID:     types.SystemID{11, 22, 33, 44, 55, 67},
							PseudonodeID: 21,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		psnp, err := DecodePSNP(buf)
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

		assert.Equalf(t, test.expected, psnp, "Test %q", test.name)
	}
}
