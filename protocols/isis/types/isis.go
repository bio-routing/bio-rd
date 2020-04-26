package types

import (
	"fmt"
)

// SystemID is an ISIS System ID
type SystemID [6]byte

// SourceID is a source ID
type SourceID struct {
	SystemID  SystemID
	CircuitID uint8
}

// MACAddress is an Ethernet MAC address
type MACAddress [6]byte

// AreaID is an ISIS Area ID
type AreaID []byte

// Equal checks if area IDs are equal
func (a AreaID) Equal(b AreaID) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (sysID *SystemID) String() string {
	return fmt.Sprintf("%d%d.%d%d.%d%d", sysID[0], sysID[1], sysID[2], sysID[3], sysID[4], sysID[5])
}

// NewSourceID creates a new SourceID
func NewSourceID(sysID SystemID, circuitID uint8) SourceID {
	return SourceID{
		SystemID:  sysID,
		CircuitID: circuitID,
	}
}

// Serialize serializes a source ID
func (srcID *SourceID) Serialize() []byte {
	return []byte{
		srcID.SystemID[0], srcID.SystemID[1], srcID.SystemID[2],
		srcID.SystemID[3], srcID.SystemID[4], srcID.SystemID[5],
		srcID.CircuitID,
	}
}

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
