package config

import (
	"net"
)

type Peer struct {
	AdminEnabled bool
	KeepAlive    uint16
	HoldTimer    uint16
	LocalAddress net.IP
	PeerAddress  net.IP
	LocalAS      uint32
	PeerAS       uint32
	Passive      bool
	RouterID     uint32
}
