package rtmirror

import (
	"crypto/sha1"
	"sync"

	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// RTMirror provides an deduplicated routing table
type RTMirror struct {
	routes   map[[20]byte]*routeContainer
	routesMu sync.Mutex
	rt       *routingtable.RoutingTable
}

// New creates a new RTMirror and starts it
func New(rt *routingtable.RoutingTable) *RTMirror {
	return &RTMirror{
		routes: make(map[[20]byte]*routeContainer),
		rt:     rt,
	}
}

// DropRIS drops all routes learned from a RIS
func (rtm *RTMirror) DropRIS(cc interface{}) {
	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	for h, rc := range rtm.routes {
		rtm._delRoute(h, cc, rc.route)
	}
}

// AddRoute adds a route
func (rtm *RTMirror) AddRoute(cc interface{}, r *routeapi.Route) error {
	h, err := hashRoute(r)
	if err != nil {
		return errors.Wrap(err, "Hashing failed")
	}

	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	if _, exists := rtm.routes[h]; !exists {
		s := route.RouteFromProtoRoute(r, true)
		rtm.routes[h] = newRouteContainer(r, cc)
		rtm.rt.AddPath(s.Prefix(), s.Paths()[0])
		return nil
	}

	rtm.routes[h].addSource(cc)
	return nil
}

// RemoveRoute deletes a route
func (rtm *RTMirror) RemoveRoute(cc interface{}, r *routeapi.Route) error {
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

func (rtm *RTMirror) _delRoute(h [20]byte, src interface{}, r *routeapi.Route) {
	rtm.routes[h].removeSource(src)

	if rtm.routes[h].srcCount() > 0 {
		return
	}

	s := route.RouteFromProtoRoute(r, true)
	rtm.rt.RemovePath(s.Prefix(), s.Paths()[0])
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
