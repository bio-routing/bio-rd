package types

import "fmt"

// SystemID is an ISIS System ID
type SystemID [6]byte

// MACAddress is an Ethernet MAC address
type MACAddress [6]byte

// AreaID is an ISIS Area ID
type AreaID []byte

func (sysID *SystemID) String() string {
	return fmt.Sprintf("%d%d.%d%d.%d%d", sysID[0], sysID[1], sysID[2], sysID[3], sysID[4], sysID[5])
}
