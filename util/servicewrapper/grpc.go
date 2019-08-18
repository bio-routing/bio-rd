package servicewrapper

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"

	log "github.com/sirupsen/logrus"
)

// CodeToLogrusLevel is the default 0-stor implementation
// of gRPC return codes to log(rus) levels for server side.
func codeToLogrusLevel(code codes.Code) log.Level {
	if level, ok := _GRPCCodeToLogrusLevelMapping[code]; ok {
		return level
	}
	return log.ErrorLevel
}

var _GRPCCodeToLogrusLevelMapping = map[codes.Code]log.Level{
	codes.OK:                 log.DebugLevel,
	codes.Canceled:           log.DebugLevel,
	codes.InvalidArgument:    log.DebugLevel,
	codes.NotFound:           log.DebugLevel,
	codes.AlreadyExists:      log.DebugLevel,
	codes.Unauthenticated:    log.InfoLevel,
	codes.PermissionDenied:   log.InfoLevel,
	codes.DeadlineExceeded:   log.WarnLevel,
	codes.ResourceExhausted:  log.WarnLevel,
	codes.FailedPrecondition: log.WarnLevel,
	codes.Aborted:            log.WarnLevel,
}

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

	unaryInterceptors = append(unaryInterceptors,
		grpc_prometheus.UnaryServerInterceptor,
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(),
		grpc_logrus.UnaryServerInterceptor(logrusEntry, levelOpt),
	)

	streamInterceptors = append(streamInterceptors,
		grpc_prometheus.StreamServerInterceptor,
		grpc_ctxtags.StreamServerInterceptor(),
		grpc_recovery.StreamServerInterceptor(),
		grpc_logrus.StreamServerInterceptor(logrusEntry, levelOpt),
	)
	unaryOpts := grpc_middleware.WithUnaryServerChain(unaryInterceptors...)
	streamOpts := grpc_middleware.WithStreamServerChain(streamInterceptors...)

	s.grpcSrv.srv = grpc.NewServer(unaryOpts, streamOpts)
	reflection.Register(s.grpcSrv.srv)
	grpc_prometheus.Register(s.grpcSrv.srv)
	grpc_prometheus.EnableClientHandlingTimeHistogram()
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s.GRPC())

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
