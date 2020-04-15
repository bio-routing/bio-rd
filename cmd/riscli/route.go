package main

import (
	"fmt"

	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/route/api"
)

func printRoute(ar *api.Route) {
	r := route.RouteFromProtoRoute(ar, false)
	fmt.Println(r.Print())
}
