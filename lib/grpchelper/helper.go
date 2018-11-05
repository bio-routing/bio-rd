package grpchelper

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/q3k/statusz"
	log "github.com/sirupsen/logrus"
)

// Server represents an exarpc server wrapper
type Server struct {
	grpcSrv *grpcSrv
	httpSrv *http.Server
	opt     grpc.ServerOption
}

type grpcSrv struct {
	port uint16
	srv  *grpc.Server
}

// New creates a new exarpc server wrapper
func New(grpcPort uint16, h *http.Server, unaryInterceptors []grpc.UnaryServerInterceptor, streamInterceptors []grpc.StreamServerInterceptor) (*Server, error) {
	s := &Server{
		grpcSrv: &grpcSrv{port: grpcPort},
		httpSrv: h,
	}

	logrusEntry := log.NewEntry(log.StandardLogger())
	levelOpt := grpc_logrus.WithLevels(codeToLogrusLevel)

	streamInterceptors = append(streamInterceptors,
		grpc_prometheus.StreamServerInterceptor,
		grpc_ctxtags.StreamServerInterceptor(),
		grpc_recovery.StreamServerInterceptor(),
		grpc_logrus.StreamServerInterceptor(logrusEntry, levelOpt),
	)

	unaryInterceptors = append(unaryInterceptors,
		grpc_prometheus.UnaryServerInterceptor,
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrusEntry, levelOpt),
	)
	opts := grpc_middleware.WithUnaryServerChain(unaryInterceptors...)

	s.grpcSrv.srv = grpc.NewServer(opts)
	reflection.Register(s.grpcSrv.srv)
	grpc_prometheus.Register(s.grpcSrv.srv)
	grpc_prometheus.EnableClientHandlingTimeHistogram()
	grpc_prometheus.EnableHandlingTimeHistogram()

	return s, nil
}

// HTTP creates an HTTP server instance
func HTTP(port uint16) *http.Server {
	h := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return h
}

// Serve starts GRPC and HTTP serving
func (s *Server) Serve() error {
	var wg sync.WaitGroup

	// GRPC
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcSrv.port))
	if err != nil {
		return fmt.Errorf("Unable to listen: %v", err)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := s.grpcSrv.srv.Serve(grpcLis)
		log.Fatalf("GRPC serving failed: %v", err)
		os.Exit(1)
		wg.Done()
	}(&wg)

	// HTTP
	http.Handle("/metrics", promhttp.Handler())
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := s.httpSrv.ListenAndServe()
		log.Fatalf("HTTP serving failed: %v", err)
		os.Exit(1)
		wg.Done()
	}(&wg)

	wg.Wait()
	return nil
}

// GRPC get's the GRPC server object
func (s *Server) GRPC() *grpc.Server {
	return s.grpcSrv.srv
}
