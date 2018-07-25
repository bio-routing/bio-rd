package config

import (
	"time"
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
}
