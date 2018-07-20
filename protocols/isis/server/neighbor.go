package server

import "github.com/bio-routing/bio-rd/protocols/isis/types"

type neighbor struct {
	systemID types.SystemID
	ifa      *netIf
}

func newNeighbor(sysID types.SystemID, ifa *netIf) *neighbor {
	return &neighbor{
		systemID: sysID,
		ifa:      ifa,
	}
}
