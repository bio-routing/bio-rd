package rtmirror

import (
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"google.golang.org/grpc"
)

// routeContainer groups a route with one ore multiple source the route was received from
type routeContainer struct {
	route   *routeapi.Route
	sources []*grpc.ClientConn
}

func newRouteContainer(route *routeapi.Route, source *grpc.ClientConn) *routeContainer {
	return &routeContainer{
		route:   route,
		sources: []*grpc.ClientConn{source},
	}
}

func (rc *routeContainer) addSource(cc *grpc.ClientConn) {
	rc.sources = append(rc.sources, cc)
}

func (rc *routeContainer) removeSource(cc *grpc.ClientConn) {
	i := rc.getSourceIndex(cc)
	if i < 0 {
		return
	}

	rc.sources[i] = rc.sources[len(rc.sources)-1]
	rc.sources = rc.sources[:len(rc.sources)-1]
}

func (rc *routeContainer) getSourceIndex(cc *grpc.ClientConn) int {
	for i := range rc.sources {
		if rc.sources[i] == cc {
			return i
		}
	}

	return -1
}

func (rc *routeContainer) srcCount() int {
	return len(rc.sources)
}
