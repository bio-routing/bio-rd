package server

import (
	"sync"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type neighbors struct {
	db   map[types.MACAddress]*neighbor
	dbMu sync.RWMutex
}

func newNeighbors() *neighbors {
	return nil
}
