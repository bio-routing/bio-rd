package risserver

import (
	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
)

type update struct {
	advertisement bool
	prefix        net.Prefix
	path          *route.Path
}

type ribClient struct {
	fifo    *updateFIFO
	stopped chan struct{}
}

func newRIBClient(fifo *updateFIFO) *ribClient {
	return &ribClient{
		fifo:    fifo,
		stopped: make(chan struct{}),
	}
}

func (r *ribClient) AddPath(pfx *net.Prefix, path *route.Path) error {
	return r.addPath(pfx, path, false)
}

func (r *ribClient) AddPathInitialDump(pfx *net.Prefix, path *route.Path) error {
	return r.addPath(pfx, path, true)
}

func (r *ribClient) addPath(pfx *net.Prefix, path *route.Path, isInitalDump bool) error {
	r.fifo.queue(&pb.RIBUpdate{
		Advertisement: true,
		IsInitialDump: isInitalDump,
		Route: &routeapi.Route{
			Pfx: pfx.ToProto(),
			Paths: []*routeapi.Path{
				path.ToProto(),
			},
		},
	})

	return nil
}

func (r *ribClient) RemovePath(pfx *net.Prefix, path *route.Path) bool {
	r.fifo.queue(&pb.RIBUpdate{
		Advertisement: false,
		Route: &routeapi.Route{
			Pfx: pfx.ToProto(),
			Paths: []*routeapi.Path{
				path.ToProto(),
			},
		},
	})

	return false
}

func (r *ribClient) RefreshRoute(*net.Prefix, []*route.Path) {}

// ReplacePath is here to fulfill an interface
func (r *ribClient) ReplacePath(*net.Prefix, *route.Path, *route.Path) {}

// Dispose stopps the ribClient. This is triggered when a BMP connection is lost so we can drop subscribed clients.
func (r *ribClient) Dispose() {
	close(r.stopped)
}
