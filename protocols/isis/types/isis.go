package types

import "fmt"

// SystemID is an ISIS System ID
type SystemID [6]byte

// SourceID is a source ID
type SourceID struct {
	SystemID  SystemID
	CircuitID uint8
}

// MACAddress is an Ethernet MAC address
type MACAddress [6]byte

func (m MACAddress) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
}

// AreaID is an ISIS Area ID
type AreaID []byte

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
