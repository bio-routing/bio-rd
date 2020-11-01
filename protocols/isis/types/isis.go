package types

import "fmt"

// SystemID is an ISIS System ID
type SystemID [6]byte

// SourceID is a source ID
type SourceID struct {
	SystemID  SystemID
	CircuitID uint8
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
