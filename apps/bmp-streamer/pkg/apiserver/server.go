package apiserver

import (
	"fmt"

	pb "github.com/bio-routing/bio-rd/apps/bmp-streamer/pkg/bmpstreamer"
	net "github.com/bio-routing/bio-rd/net"
	netapi "github.com/bio-routing/bio-rd/net/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
)

// APIServer implements the BMP server API
type APIServer struct {
	bmpServer *server.BMPServer
}

// New creates an new API server
func New(bmpServer *server.BMPServer) *APIServer {
	return &APIServer{
		bmpServer: bmpServer,
	}
}

func (u update) toRIBUpdate() *pb.RIBUpdate {
	toSend := &pb.RIBUpdate{
		Advertisement: u.advertisement,
		Peer:          u.route.Paths()[0].BGPPath.Source.ToProto(),
		Route:         u.route.ToProto(),
	}

	return toSend
}

// AdjRIBInStream offers RIBs as a stream
func (a *APIServer) AdjRIBInStream(req *pb.AdjRIBInStreamRequest, stream pb.RIBService_AdjRIBInStreamServer) error {
	r4 := newRIBClient()
	r6 := newRIBClient()

	addr := net.IP{}
	if req.Router.Version == netapi.IP_IPv4 {
		addr = net.IPv4(uint32(req.Router.Lower))
	} else if req.Router.Version == netapi.IP_IPv6 {
		addr = net.IPv6(req.Router.Higher, req.Router.Lower)
	} else {
		return fmt.Errorf("Unknown protocol")
	}

	ret := make(chan error)
	go func() {
		var err error
		u := update{}
		for {
			select {
			case u = <-r4.ch:
			case u = <-r6.ch:
			}

			toSend := u.toRIBUpdate()

			err = stream.Send(toSend)
			if err != nil {
				ret <- err
				return
			}
		}
	}()

	a.bmpServer.SubscribeRIBs(r4, addr.ToNetIP(), packet.IPv4AFI)
	defer a.bmpServer.UnsubscribeRIBs(r4, addr.ToNetIP(), packet.IPv4AFI)

	a.bmpServer.SubscribeRIBs(r6, addr.ToNetIP(), packet.IPv6AFI)
	defer a.bmpServer.UnsubscribeRIBs(r6, addr.ToNetIP(), packet.IPv6AFI)

	err := <-ret
	if err != nil {
		return fmt.Errorf("Stream ended: %v", err)
	}

	return nil
}

type update struct {
	advertisement bool
	route         *route.Route
}

type ribClient struct {
	ch chan update
}

func newRIBClient() *ribClient {
	return &ribClient{
		ch: make(chan update),
	}
}

func (r *ribClient) AddPath(pfx net.Prefix, path *route.Path) error {
	r.ch <- update{
		advertisement: true,
		route:         route.NewRoute(pfx, path),
	}

	return nil
}

func (r *ribClient) RemovePath(pfx net.Prefix, path *route.Path) bool {
	r.ch <- update{
		advertisement: false,
		route:         route.NewRoute(pfx, path),
	}

	return false
}

func (r *ribClient) UpdateNewClient(routingtable.RouteTableClient) error {
	return nil
}

func (r *ribClient) Register(routingtable.RouteTableClient) {
}

func (r *ribClient) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
}

func (r *ribClient) Unregister(routingtable.RouteTableClient) {
}

func (r *ribClient) RouteCount() int64 {
	return -1
}
