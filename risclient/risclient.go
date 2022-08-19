package risclient

import (
	"context"
	"io"
	"sync"

	risapi "github.com/bio-routing/bio-rd/cmd/ris/api"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/util/log"
	"google.golang.org/grpc"
)

// Client is a client interface
type Client interface {
	AddRoute(src interface{}, r *routeapi.Route) error
	RemoveRoute(src interface{}, r *routeapi.Route) error
	DropAllBySrc(src interface{})
}

// RISClient represents a RIS client
type RISClient struct {
	req    *Request
	cc     *grpc.ClientConn
	c      Client
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// Request is a RISClient config
type Request struct {
	Router string
	VRFRD  uint64
	AFI    risapi.ObserveRIBRequest_AFISAFI
}

func (r *Request) toProtoRequest() *risapi.ObserveRIBRequest {
	return &risapi.ObserveRIBRequest{
		Router:  r.Router,
		VrfId:   r.VRFRD,
		Afisafi: r.AFI,
	}
}

// New creates a new RISClient
func New(req *Request, cc *grpc.ClientConn, c Client) *RISClient {
	return &RISClient{
		req:    req,
		cc:     cc,
		c:      c,
		stopCh: make(chan struct{}),
	}
}

// Stop stops the client
func (r *RISClient) Stop() {
	close(r.stopCh)
}

// Start starts the client
func (r *RISClient) Start() {
	r.wg.Add(1)

	go r.run()
}

// Wait blocks until the client is fully stopped
func (r *RISClient) Wait() {
	r.wg.Wait()
}

func (r *RISClient) stopped() bool {
	select {
	case <-r.stopCh:
		return true
	default:
		return false
	}
}

func (r *RISClient) run() {
	for {
		if r.stopped() {
			return
		}

		risc := risapi.NewRoutingInformationServiceClient(r.cc)

		orc, err := risc.ObserveRIB(context.Background(), r.req.toProtoRequest(), grpc.WaitForReady(true))
		if err != nil {
			log.WithError(err).Error("ObserveRIB call failed")
			continue
		}

		err = r.serviceLoop(orc)
		if err == nil {
			return
		}

		r.serviceLoopLogging(err)
	}
}

func (r *RISClient) serviceLoopLogging(err error) {
	if err == io.EOF {
		log.WithError(err).WithFields(log.Fields{
			"component": "RISClient",
			"function":  "run",
		}).Info("ObserveRIB ended")
		return
	}

	log.WithError(err).WithFields(log.Fields{
		"component": "RISClient",
		"function":  "run",
	}).Error("ObserveRIB ended")
}

func (r *RISClient) serviceLoop(orc risapi.RoutingInformationService_ObserveRIBClient) error {
	defer r.processDownEvent()

	for {
		if r.stopped() {
			return nil
		}

		u, err := orc.Recv()
		if err != nil {
			return err
		}

		r.processUpdate(u)
	}
}

func (r *RISClient) processUpdate(u *risapi.RIBUpdate) {
	if u.Advertisement {
		r.processAdvertisement(u)
		return
	}

	r.processWithdraw(u)
}

func (r *RISClient) processAdvertisement(u *risapi.RIBUpdate) {
	r.c.AddRoute(r.cc, u.Route)
}

func (r *RISClient) processWithdraw(u *risapi.RIBUpdate) {
	r.c.RemoveRoute(r.cc, u.Route)
}

func (r *RISClient) processDownEvent() {
	r.c.DropAllBySrc(r.cc)
}
