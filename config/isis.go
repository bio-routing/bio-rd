package config

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type ISISConfig struct {
	NETs                       []NET
	Interfaces                 []ISISInterfaceConfig
	TrafficEngineeringRouterID [4]byte
}

type ISISInterfaceConfig struct {
	Name             string
	Passive          bool
	P2P              bool
	ISISLevel1Config *ISISLevelConfig
	ISISLevel2Config *ISISLevelConfig
}

type ISISLevelConfig struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	Priority      uint8
}

// NET represents an ISO network entity title
type NET struct {
	AFI      byte
	AreaID   types.AreaID
	SystemID types.SystemID
	SEL      byte
}

func parseNET(addr []byte) (*NET, error) {
	l := len(addr)

	if l < 8 {
		return nil, fmt.Errorf("NET too short")
	}

	if l > 20 {
		return nil, fmt.Errorf("NET too long")
	}

	areaID := []byte{}

	for i := 0; i < l-8; i++ {
		areaID = append(areaID, addr[i+1])
	}

	systemID := types.SystemID{
		addr[l-7],
		addr[l-6],
		addr[l-5],
		addr[l-4],
		addr[l-3],
		addr[l-2],
	}

	return &NET{
		AFI:      addr[0],
		AreaID:   areaID,
		SystemID: systemID,
		SEL:      addr[l-1],
	}, nil
}
