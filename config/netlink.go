package config

import (
	"time"

	"github.com/bio-routing/bio-rd/routingtable/filter"
)

// Constants for default routing tables in the Linux Kernel
const (
	RtLocal   int = 255 // according to man ip-route: 255 is reserved fro built-in use
	RtMain    int = 254 // This is the default table where routes are inserted
	RtDefault int = 253 // according to man ip-route: 253 is reserved fro built-in use
	RtUnspec  int = 0   // according to man ip-route: 0 is reserved fro built-in use

)

// Netlink holds the configuration of the Netlink protocol
type Netlink struct {
	HoldTime       time.Duration
	UpdateInterval time.Duration
	RoutingTable   int
	ImportFilter   *filter.Filter // Which routes are imported from the Kernel
	ExportFilter   *filter.Filter // Which routes are exported to the Kernel
}
