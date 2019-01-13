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
					packet.LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 0x00,
						LSPNumber:    1,
					}: &lsdbEntry{
						lspdu: &packet.LSPDU{
							RemainingLifetime: 5,
						},
					},
					packet.LSPID{
						SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
						PseudonodeID: 0x00,
						LSPNumber:    1,
					}: &lsdbEntry{
						lspdu: &packet.LSPDU{
							RemainingLifetime: 1,
						},
					},
				},
			},
			expected: &lsdb{
				lsps: map[packet.LSPID]*lsdbEntry{
					packet.LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 0x00,
						LSPNumber:    1,
					}: &lsdbEntry{
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
			packet.LSPID{
				SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: &lsdbEntry{
				lspdu: &packet.LSPDU{
					RemainingLifetime: 5,
				},
			},
			packet.LSPID{
				SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: &lsdbEntry{
				lspdu: &packet.LSPDU{
					RemainingLifetime: 1,
				},
			},
		},
	}
	expected := &lsdb{
		done: make(chan struct{}),
		lsps: map[packet.LSPID]*lsdbEntry{
			packet.LSPID{
				SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: &lsdbEntry{
				lspdu: &packet.LSPDU{
					RemainingLifetime: 5,
				},
			},
			packet.LSPID{
				SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
				PseudonodeID: 0x00,
				LSPNumber:    1,
			}: &lsdbEntry{
				lspdu: &packet.LSPDU{
					RemainingLifetime: 1,
				},
			},
		},
	}
	ticker := btime.NewMockTicker()
	db.start(ticker)

	expected.decrementRemainingLifetimes()
	expected.decrementRemainingLifetimes()
	expected.decrementRemainingLifetimes()
	expected.decrementRemainingLifetimes()

	ticker.Tick()
	ticker.Tick()
	ticker.Tick()
	ticker.Tick()

	db.stop()
	assert.Equal(t, db.lsps, expected.lsps)
}
