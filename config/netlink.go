package config

import (
	"time"

	"github.com/bio-routing/bio-rd/routingtable/filter"
)

const (
	RtLocal   int = 255
	RtMain    int = 254
	RtDefault int = 253
	RtUnspec  int = 0
)

type Netlink struct {
	HoldTime       time.Duration
	UpdateInterval time.Duration
	RoutingTable   int
	ImportFilter   *filter.Filter // Which routes are imported from the Kernel
	ExportFilter   *filter.Filter // Which routes are exportet to the Kernel
}
