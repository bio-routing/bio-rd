package mergedlocrib

import (
	"crypto/sha1"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/mergedlocrib/metrics"
)

// MergedLocRIB provides an deduplicated routing table
type MergedLocRIB struct {
	routes   map[[sha1.Size]byte]*routeContainer
	routesMu sync.RWMutex
	locRIB   *locRIB.LocRIB
}

// New creates a new MergedLocRIB and starts it
func New(locRIB *locRIB.LocRIB) *MergedLocRIB {
	return &MergedLocRIB{
		routes: make(map[[sha1.Size]byte]*routeContainer),
		locRIB: locRIB,
	}
}

// DropAllBySrc drops all routes learned from a source
func (rtm *MergedLocRIB) DropAllBySrc(src interface{}) {
	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	for h, rc := range rtm.routes {
		rtm._delRoute(h, src, rc.route)
	}
}

// AddRoute adds a route
func (rtm *MergedLocRIB) AddRoute(cc interface{}, r *routeapi.Route) error {
	h, err := hashRoute(r)
	if err != nil {
		return fmt.Errorf("hashing failed: %w", err)
	}

	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	if _, exists := rtm.routes[h]; !exists {
		s := route.RouteFromProtoRoute(r, true)
		rtm.routes[h] = newRouteContainer(r, cc)
		rtm.locRIB.AddPath(s.Prefix(), s.Paths()[0])
		return nil
	}

	rtm.routes[h].addSource(cc)
	return nil
}

// RemoveRoute deletes a route
func (rtm *MergedLocRIB) RemoveRoute(cc interface{}, r *routeapi.Route) error {
	h, err := hashRoute(r)
	if err != nil {
		return fmt.Errorf("hashing failed: %w", err)
	}

	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	if _, exists := rtm.routes[h]; !exists {
		return nil
	}

	rtm._delRoute(h, cc, r)
	return nil
}

func (rtm *MergedLocRIB) _delRoute(h [sha1.Size]byte, src interface{}, r *routeapi.Route) {
	rtm.routes[h].removeSource(src)

	if rtm.routes[h].srcCount() > 0 {
		return
	}

	s := route.RouteFromProtoRoute(r, true)
	rtm.locRIB.RemovePath(s.Prefix(), s.Paths()[0])
	delete(rtm.routes, h)
}

func hashRoute(route *routeapi.Route) ([sha1.Size]byte, error) {
	m, err := proto.Marshal(route)
	if err != nil {
		return [sha1.Size]byte{}, fmt.Errorf("proto marshal failed: %w", err)
	}

	h := sha1.New()
	_, err = h.Write(m)
	if err != nil {
		return [sha1.Size]byte{}, fmt.Errorf("write failed: %w", err)
	}
	res := [sha1.Size]byte{}
	x := h.Sum(nil)
	copy(res[:], x)

	return res, nil
}

// Metrics gets the metrics
func (rtm *MergedLocRIB) Metrics() *metrics.MergedLocRIBMetrics {
	rtm.routesMu.RLock()
	defer rtm.routesMu.RUnlock()

	return &metrics.MergedLocRIBMetrics{
		RIBName:                     rtm.locRIB.Name(),
		UniqueRouteCount:            uint64(len(rtm.routes)),
		RoutesWithSingleSourceCount: rtm._getRoutesWithSingleSourceCount(),
	}
}

func (rtm *MergedLocRIB) _getRoutesWithSingleSourceCount() uint64 {
	n := uint64(0)

	for _, r := range rtm.routes {
		if len(r.sources) == 1 {
			n++
		}
	}

	return n
}
