package mergedlocrib

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

func (rc *routeContainer) addSource(src interface{}) {
	rc.sources = append(rc.sources, src)
}

func (rc *routeContainer) removeSource(src interface{}) {
	i := rc.getSourceIndex(src)
	if i < 0 {
		return
	}

	rc.sources[i] = rc.sources[len(rc.sources)-1]
	rc.sources = rc.sources[:len(rc.sources)-1]
}

func (rc *routeContainer) getSourceIndex(src interface{}) int {
	for i := range rc.sources {
		if rc.sources[i] == src {
			return i
		}
	}

	return -1
}

func (rc *routeContainer) srcCount() int {
	return len(rc.sources)
}
