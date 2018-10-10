package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestScanSRMSSN(t *testing.T) {
	ifa := &netIf{}
	ifb := &netIf{}
	db := &lsdb{
		lsps: map[packet.LSPID]*lsdbEntry{
			packet.LSPID{SystemID: types.SystemID{1, 2, 3, 4, 5, 6}, PseudonodeID: 0}: {
				lspdu: &packet.LSPDU{
					SequenceNumber: 100,
				},
				srmFlags: map[*netIf]struct{}{
					ifa: struct{}{},
				},
				ssnFlags: map[*netIf]struct{}{
					ifb: struct{}{},
				},
			},
			packet.LSPID{SystemID: types.SystemID{1, 2, 3, 4, 5, 7}, PseudonodeID: 0}: {
				lspdu: &packet.LSPDU{
					SequenceNumber: 200,
				},
				srmFlags: map[*netIf]struct{}{
					ifb: struct{}{},
				},
				ssnFlags: map[*netIf]struct{}{
					ifa: struct{}{},
				},
			},
		},
	}

	lspdus, psnpEntries := db.scanSRMSSN(ifa)
	assert.Equal(t, []*packet.LSPDU{
		{
			SequenceNumber: 100,
		},
	}, lspdus)

	assert.Equal(t, []*packet.LSPEntry{
		{
			SequenceNumber: 200,
		},
	}, psnpEntries)

	assert.Equal(t, &lsdb{
		lsps: map[packet.LSPID]*lsdbEntry{
			packet.LSPID{SystemID: types.SystemID{1, 2, 3, 4, 5, 6}, PseudonodeID: 0}: {
				lspdu: &packet.LSPDU{
					SequenceNumber: 100,
				},
				srmFlags: map[*netIf]struct{}{},
				ssnFlags: map[*netIf]struct{}{
					ifb: struct{}{},
				},
			},
			packet.LSPID{SystemID: types.SystemID{1, 2, 3, 4, 5, 7}, PseudonodeID: 0}: {
				lspdu: &packet.LSPDU{
					SequenceNumber: 200,
				},
				srmFlags: map[*netIf]struct{}{
					ifb: struct{}{},
				},
				ssnFlags: map[*netIf]struct{}{},
			},
		},
	}, db)

	/*for _, test := range tests {
		lsps, psnpEntries :=

		assert.Equal(t, test.expectedLSPs, lsps, "Test %q (LSPs)", test.name)
		assert.Equal(t, test.expectedPSNPEntries, psnpEntries, "Test %q (PSNP Entries)", test.name)
	}*/
}
