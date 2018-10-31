package config

import (
	"time"

	"github.com/bio-routing/bio-rd/routingtable/filter"
)

// Constants for default routing tables in the Linux Kernel
const (
	RtLocal   uint32 = 255 // according to man ip-route: 255 is reserved for built-in use
	RtMain    uint32 = 254 // This is the default table where routes are inserted
	RtDefault uint32 = 253 // according to man ip-route: 253 is reserved for built-in use
	RtUnspec  uint32 = 0   // according to man ip-route: 0 is reserved for built-in use

)

// Netlink holds the configuration of the Netlink protocol
type Netlink struct {
	HoldTime       time.Duration
	UpdateInterval time.Duration
	RoutingTable   uint32
	ImportFilter   *filter.Filter // Which routes are imported from the Kernel
	ExportFilter   *filter.Filter // Which routes are exported to the Kernel
}
