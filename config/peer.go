package config

import (
	"net"
	"time"

	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

type Peer struct {
	AdminEnabled      bool
	ReconnectInterval time.Duration
	KeepAlive         time.Duration
	HoldTime          time.Duration
	LocalAddress      net.IP
	PeerAddress       net.IP
	LocalAS           uint32
	PeerAS            uint32
	Passive           bool
	RouterID          uint32
	AddPathSend       routingtable.ClientOptions
	AddPathRecv       bool
	ImportFilter      *filter.Filter
	ExportFilter      *filter.Filter
	RouteServerClient bool
}
