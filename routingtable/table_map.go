package routingtable

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type RoutingTableMap struct {
	table map[*bnet.Prefix][]route.Route
	mu    sync.RWMutex
}

func NewRoutingTableMap() *RoutingTableMap {
	return &RoutingTableMap{
		table: make(map[*bnet.Prefix][]route.Route),
	}
}

func (rtm *RoutingTableMap) Dump() []*route.Route {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()

	pathCount := 0
	for _, x := range rtm.table {
		pathCount += len(x)
	}

	res := make([]*route.Route, 0, pathCount)
	for _, v := range rtm.table {
		for _, r := range v {
			res = append(res, &r)
		}
	}

	return res
}

func (rtm *RoutingTableMap) AddPath(pfx *net.Prefix, p *route.Path) error {
	return nil
}

func (rtm *RoutingTableMap) ReplacePath(pfx *net.Prefix, p *route.Path) []*route.Path {
	return nil
}

func (rtm *RoutingTableMap) RemovePath(pfx *net.Prefix, p *route.Path) {

}

func (rtm *RoutingTableMap) RemovePfx(pfx *net.Prefix) []*route.Path {
	return nil
}
