package routingtable

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type RIB interface {
	AddPath(*net.Prefix, *route.Path)
	RemovePath(*net.Prefix, *route.Path)
}
