package rtmirror

import (
	"context"
	"crypto/sha1"
	"io"
	"sync"

	risapi "github.com/bio-routing/bio-rd/cmd/ris/api"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

// RTMirror provides an deduplicated mirror of a router/vrf/afi routing table from a multiple RIS instances
type RTMirror struct {
	cfg         Config
	vrf         *vrf.VRF
	routes      map[[20]byte]*routeContainer
	routesMu    sync.Mutex
	grpcClients []*grpc.ClientConn
	stop        chan struct{}
	wg          sync.WaitGroup
}

// Config is a route mirror config
type Config struct {
	Router string
	VRF    *vrf.VRF
}

// New creates a new RTMirror and starts it
func New(clientConns []*grpc.ClientConn, cfg Config) *RTMirror {
	rtm := &RTMirror{
		cfg:         cfg,
		routes:      make(map[[20]byte]*routeContainer),
		vrf:         cfg.VRF,
		grpcClients: clientConns,
		stop:        make(chan struct{}),
	}

	afis := []risapi.ObserveRIBRequest_AFISAFI{
		risapi.ObserveRIBRequest_IPv4Unicast,
		risapi.ObserveRIBRequest_IPv6Unicast,
	}

	for _, afi := range afis {
		for _, ris := range rtm.grpcClients {
			rtm.wg.Add(1)
			go rtm.client(ris, afi)
		}
	}

	return rtm
}

func (rtm *RTMirror) addRIS(addr string) error {
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "grpc dial failed")
	}

	rtm.grpcClients = append(rtm.grpcClients, cc)

	return nil
}

// Dispose stops the RTMirror
func (rtm *RTMirror) Dispose() {
	close(rtm.stop)

	for _, cc := range rtm.grpcClients {
		cc.Close()
	}

	rtm.wg.Wait()
}

func (rtm *RTMirror) client(cc *grpc.ClientConn, afi risapi.ObserveRIBRequest_AFISAFI) {
	defer rtm.wg.Done()

	risc := risapi.NewRoutingInformationServiceClient(cc)

	for {
		if rtm.stopped() {
			return
		}

		orc, err := risc.ObserveRIB(context.Background(), &risapi.ObserveRIBRequest{
			Router:  rtm.cfg.Router,
			VrfId:   rtm.cfg.VRF.RD(),
			Afisafi: afi,
		}, grpc.WaitForReady(true))
		if err != nil {
			log.WithError(err).Error("ObserveRIB call failed")
			continue
		}

		err = rtm.clientServiceLoop(cc, orc)
		if err != nil {
			log.WithError(err).Error("client service loop failed")
		}

		rtm.dropRoutesFromRIS(cc)
	}
}

func (rtm *RTMirror) dropRoutesFromRIS(cc *grpc.ClientConn) {
	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	for h, rc := range rtm.routes {
		rtm._delRoute(h, cc, rc.route)
	}
}

func (rtm *RTMirror) stopped() bool {
	select {
	case <-rtm.stop:
		return true
	default:
		return false
	}
}

func (rtm *RTMirror) clientServiceLoop(cc *grpc.ClientConn, orc risapi.RoutingInformationService_ObserveRIBClient) error {
	for {
		if rtm.stopped() {
			return nil
		}

		u, err := orc.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return errors.Wrap(err, "recv failed")
		}

		if u.Advertisement {
			rtm.addRoute(cc, u.Route)
			continue
		}

		rtm.delRoute(cc, u.Route)
	}
}

func (rtm *RTMirror) addRoute(cc *grpc.ClientConn, r *routeapi.Route) {
	h, err := hashRoute(r)
	if err != nil {
		log.WithError(err).Error("Hashing failed")
		return
	}

	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	if _, exists := rtm.routes[h]; !exists {
		s := route.RouteFromProtoRoute(r, true)
		rib := rtm.getRIB(s.Prefix().Addr())

		rtm.routes[h] = newRouteContainer(r, cc)
		rib.AddPath(s.Prefix(), s.Paths()[0])
		return
	}

	rtm.routes[h].addSource(cc)
}

func (rtm *RTMirror) getRIB(addr *bnet.IP) *locRIB.LocRIB {
	if addr.IsIPv4() {
		return rtm.vrf.IPv4UnicastRIB()
	}

	return rtm.vrf.IPv6UnicastRIB()
}

func (rtm *RTMirror) delRoute(cc *grpc.ClientConn, r *routeapi.Route) {
	h, err := hashRoute(r)
	if err != nil {
		log.WithError(err).Error("Hashing failed")
		return
	}

	rtm.routesMu.Lock()
	defer rtm.routesMu.Unlock()

	if _, exists := rtm.routes[h]; !exists {
		return
	}

	rtm._delRoute(h, cc, r)
}

func (rtm *RTMirror) _delRoute(h [20]byte, cc *grpc.ClientConn, r *routeapi.Route) {
	rtm.routes[h].removeSource(cc)

	if rtm.routes[h].srcCount() > 0 {
		return
	}

	s := route.RouteFromProtoRoute(r, true)
	rib := rtm.getRIB(s.Prefix().Addr())
	rib.RemovePath(s.Prefix(), s.Paths()[0])
	delete(rtm.routes, h)
}

// GetVRF exposes the mirrors VRF
func (rtm *RTMirror) GetVRF() *vrf.VRF {
	return rtm.vrf
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
