package rtmirror

import (
	routeapi "github.com/bio-routing/bio-rd/route/api"
)

// routeContainer groups a route with one ore multiple source the route was received from
type routeContainer struct {
	route   *routeapi.Route
	sources []interface{}
}

func newRouteContainer(route *routeapi.Route, source interface{}) *routeContainer {
	return &routeContainer{
		route:   route,
		sources: []interface{}{source},
	}
}

func (rc *routeContainer) addSource(cc interface{}) {
	rc.sources = append(rc.sources, cc)
}

func (rc *routeContainer) removeSource(cc interface{}) {
	i := rc.getSourceIndex(cc)
	if i < 0 {
		return
	}

	rc.sources[i] = rc.sources[len(rc.sources)-1]
	rc.sources = rc.sources[:len(rc.sources)-1]
}

func (rc *routeContainer) getSourceIndex(cc interface{}) int {
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
