package risclient

import (
	"context"
	"io"
	"sync"
	"time"

	risapi "github.com/bio-routing/bio-rd/cmd/ris/api"
	routeapi "github.com/bio-routing/bio-rd/route/api"
	"github.com/bio-routing/bio-rd/util/log"
	"github.com/cenkalti/backoff/v4"
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
	req     *Request
	cc      *grpc.ClientConn
	c       Client
	backoff *backoff.ExponentialBackOff
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// Request is a RISClient config
type Request struct {
	Router          string
	VRFRD           uint64
	AFI             risapi.ObserveRIBRequest_AFISAFI
	AllowUnreadyRib bool
}

func (r *Request) toProtoRequest() *risapi.ObserveRIBRequest {
	return &risapi.ObserveRIBRequest{
		Router:          r.Router,
		VrfId:           r.VRFRD,
		Afisafi:         r.AFI,
		AllowUnreadyRib: r.AllowUnreadyRib,
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

// this method is responible for actually calling the ORC. If it returns nil, the ris client shall be gracefully stopped.
// if an error is returned, retry with backoff
func (r *RISClient) runORC() error {
	if r.stopped() {
		return nil
	}

	risc := risapi.NewRoutingInformationServiceClient(r.cc)

	orc, err := risc.ObserveRIB(context.Background(), r.req.toProtoRequest(), grpc.WaitForReady(true))
	if err != nil {
		log.WithError(err).Error("ObserveRIB call failed")
		return err
	}

	err = r.serviceLoop(orc)
	if err != nil {
		r.serviceLoopLogging(err)
		return err
	}
	// apparently, all went well. reset backoff.
	// If we need to backoff again after here, it is a new error.
	r.backoff.Reset()

	return nil
}

func (r *RISClient) run() {
	// initialize backoff
	r.backoff = backoff.NewExponentialBackOff()
	r.backoff.MaxElapsedTime = 0
	r.backoff.MaxInterval = 5 * time.Minute
	r.backoff.InitialInterval = 5 * time.Second

	// execute runORC until in returns the error nil
	err := backoff.Retry(r.runORC, r.backoff)
	if err != nil {
		log.WithError(err).Error("Running the observeRibClient with backoff failed.")
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
