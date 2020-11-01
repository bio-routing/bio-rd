package types

import "fmt"

const (
	minNETLen = 8
	maxNETLen = 20
)

// NET represents an ISO network entity title
type NET struct {
	AFI      byte
	AreaID   AreaID
	SystemID SystemID
	SEL      byte
}

// ParseNET parses an network entity title
func ParseNET(addr []byte) (*NET, error) {
	l := len(addr)

	if l < minNETLen {
		return nil, fmt.Errorf("NET too short")
	}

	if l > maxNETLen {
		return nil, fmt.Errorf("NET too long")
	}

	areaID := []byte{}

	for i := 0; i < l-8; i++ {
		areaID = append(areaID, addr[i+1])
	}

	systemID := SystemID{
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
