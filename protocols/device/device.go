package device

import (
	"net"
)

const (
	IfAdminDown = 0
	IfLinkDown  = 1
	IfUp        = 2
	IfUnknown   = 255
)

type LinkUpdate struct {
	Index        uint64
	MTU          uint16
	Name         string
	HardwareAddr net.HardwareAddr
	Flags        net.Flags
	OperState    uint8
}
