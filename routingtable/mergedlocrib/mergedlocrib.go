package mergedlocrib

import (
	"crypto/sha1"
	"sync"

	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// MergedLocRIB provides an deduplicated routing table
type MergedLocRIB struct {
	routes   map[[20]byte]*routeContainer
	routesMu sync.Mutex
	locRIB   *locRIB.LocRIB
}

// New creates a new MergedLocRIB and starts it
func New(locRIB *locRIB.LocRIB) *MergedLocRIB {
	return &MergedLocRIB{
		routes: make(map[[20]byte]*routeContainer),
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
		return errors.Wrap(err, "Hashing failed")
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
		return errors.Wrap(err, "Hashing failed")
	}

	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	if _, exists := rtm.routes[h]; !exists {
		return nil
	}

	rtm._delRoute(h, cc, r)
	return nil
}

func (rtm *MergedLocRIB) _delRoute(h [20]byte, src interface{}, r *routeapi.Route) {
	rtm.routes[h].removeSource(src)

	if rtm.routes[h].srcCount() > 0 {
		return
	}

	s := route.RouteFromProtoRoute(r, true)
	rtm.locRIB.RemovePath(s.Prefix(), s.Paths()[0])
	delete(rtm.routes, h)
}

func hashRoute(route *routeapi.Route) ([20]byte, error) {
	m, err := proto.Marshal(route)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "Proto marshal failed")
	}

	h := sha1.New()
	_, err = h.Write(m)
	if err != nil {
		return [20]byte{}, errors.Wrap(err, "Write failed")
	}
	res := [20]byte{}
	x := h.Sum(nil)
	copy(res[:], x)

	return res, nil
}
