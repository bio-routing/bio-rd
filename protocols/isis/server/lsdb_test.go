package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"
	"github.com/stretchr/testify/assert"
)

func TestLSDPDispose(t *testing.T) {
	l := newLSDB(&Server{})
	l.dispose()

	if l.srv != nil {
		t.Errorf("srv reference not cleared")
	}
}

func TestDecrementRemainingLifetimes(t *testing.T) {
	tests := []struct {
		name     string
		lsdb     *lsdb
		expected *lsdb
	}{
		{
			name: "Test #1",
			lsdb: &lsdb{
				lsps: map[packet.LSPID]*lsdbEntry{
					{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 0x00,
						LSPNumber:    1,
					}: {
						lspdu: &packet.LSPDU{
							RemainingLifetime: 5,
						},
					},
					{
						SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
						PseudonodeID: 0x00,
						LSPNumber:    1,
					}: {
						lspdu: &packet.LSPDU{
							RemainingLifetime: 1,
						},
					},
				},
			},
			expected: &lsdb{
				lsps: map[packet.LSPID]*lsdbEntry{
					{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 0x00,
						LSPNumber:    1,
					}: {
						lspdu: &packet.LSPDU{
							RemainingLifetime: 4,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.lsdb.decrementRemainingLifetimes()
		assert.Equal(t, test.expected, test.lsdb)
	}
}

func TestStartStop(t *testing.T) {
	db := &lsdb{
		done: make(chan struct{}),
		lsps: map[packet.LSPID]*lsdbEntry{
			{
				SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: {
				lspdu: &packet.LSPDU{
					RemainingLifetime: 5,
				},
			},
			{
				SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: {
				lspdu: &packet.LSPDU{
					RemainingLifetime: 1,
				},
			},
		},
	}
	expected := &lsdb{
		done: make(chan struct{}),
		lsps: map[packet.LSPID]*lsdbEntry{
			{
				SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: {
				lspdu: &packet.LSPDU{
					RemainingLifetime: 5,
				},
			},
			{
				SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: {
				lspdu: &packet.LSPDU{
					RemainingLifetime: 1,
				},
			},
		},
	}
	lifetimeTicker := btime.NewMockTicker()
	sendTicker := btime.NewMockTicker()
	psnpTicker := btime.NewMockTicker()
	csnpTicker := btime.NewMockTicker()
	db.start(lifetimeTicker, sendTicker, psnpTicker, csnpTicker)

	expected.decrementRemainingLifetimes()
	expected.decrementRemainingLifetimes()
	expected.decrementRemainingLifetimes()
	expected.decrementRemainingLifetimes()

	lifetimeTicker.Tick()
	lifetimeTicker.Tick()
	lifetimeTicker.Tick()
	lifetimeTicker.Tick()

	db.stop()
	assert.Equal(t, db.lsps, expected.lsps)
}
