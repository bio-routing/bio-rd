package config

import (
	"net"

	"github.com/bio-routing/bio-rd/routingtable/filter"

	"time"

	"github.com/bio-routing/bio-rd/routingtable"
)

type Peer struct {
	AdminEnabled      bool
	KeepAlive         uint16
	HoldTimer         uint16
	LocalAddress      net.IP
	PeerAddress       net.IP
	LocalAS           uint32
	PeerAS            uint32
	Passive           bool
	RouterID          uint32
	AddPathSend       routingtable.ClientOptions
	AddPathRecv       bool
	ReconnectInterval time.Duration
	ImportFilter      *filter.Filter
	ExportFilter      *filter.Filter
}
