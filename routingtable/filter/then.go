package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type Then interface {
	Do(p net.Prefix, pa *route.Path) (bool, *route.Path)
}
