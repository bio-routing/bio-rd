package types

import "fmt"

const (
	minNETLen = 8
	maxNETLen = 20
)

// NET represents an ISO network entity title
type NET struct {
	AreaID   AreaID
	SystemID SystemID
	SEL      byte
}

// ParseNET parses an network entity title
func ParseNET(addr []byte) (*NET, error) {
	addrLen := len(addr)

	if addrLen < minNETLen {
		return nil, fmt.Errorf("NET too short")
	}

	if addrLen > maxNETLen {
		return nil, fmt.Errorf("NET too long")
	}

	areaID := []byte{}

	for i := 0; i < addrLen-systemIDLen-1; i++ {
		areaID = append(areaID, addr[i])
	}

	systemID := SystemID{
		addr[addrLen-7],
		addr[addrLen-6],
		addr[addrLen-5],
		addr[addrLen-4],
		addr[addrLen-3],
		addr[addrLen-2],
	}

	return &NET{
		AreaID:   areaID,
		SystemID: systemID,
		SEL:      addr[addrLen-1],
	}, nil
}
