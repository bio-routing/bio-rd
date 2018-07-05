package config

import (
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
)

// Peer defines the configuration for a BGP session
type Peer struct {
	AdminEnabled            bool
	ReconnectInterval       time.Duration
	KeepAlive               time.Duration
	HoldTime                time.Duration
	LocalAddress            bnet.IP
	PeerAddress             bnet.IP
	LocalAS                 uint32
	PeerAS                  uint32
	Passive                 bool
	RouterID                uint32
	AddPathSend             routingtable.ClientOptions
	AddPathRecv             bool
	ImportFilter            *filter.Filter
	ExportFilter            *filter.Filter
	RouteServerClient       bool
	RouteReflectorClient    bool
	RouteReflectorClusterID uint32
	IPv6                    bool
}
