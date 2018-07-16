package filter

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
)

// Action performs actions on a `route.Path`
type Action interface {
	Do(p net.Prefix, pa *route.Path) actions.Result
}
