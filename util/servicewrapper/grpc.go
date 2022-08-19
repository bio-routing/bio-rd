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
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// CodeToLogrusLevel is the default 0-stor implementation
// of gRPC return codes to log(rus) levels for server side.
func codeToLogrusLevel(code codes.Code) logrus.Level {
	if level, ok := _GRPCCodeToLogrusLevelMapping[code]; ok {
		return level
	}
	return logrus.ErrorLevel
}

var _GRPCCodeToLogrusLevelMapping = map[codes.Code]logrus.Level{
	codes.OK:                 logrus.DebugLevel,
	codes.Canceled:           logrus.DebugLevel,
	codes.InvalidArgument:    logrus.DebugLevel,
	codes.NotFound:           logrus.DebugLevel,
	codes.AlreadyExists:      logrus.DebugLevel,
	codes.Unauthenticated:    logrus.InfoLevel,
	codes.PermissionDenied:   logrus.InfoLevel,
	codes.DeadlineExceeded:   logrus.WarnLevel,
	codes.ResourceExhausted:  logrus.WarnLevel,
	codes.FailedPrecondition: logrus.WarnLevel,
	codes.Aborted:            logrus.WarnLevel,
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
func New(grpcPort uint16, h *http.Server, unaryInterceptors []grpc.UnaryServerInterceptor, streamInterceptors []grpc.StreamServerInterceptor, keepalivePol keepalive.EnforcementPolicy) (*Server, error) {
	s := &Server{
		grpcSrv: &grpcSrv{port: grpcPort},
		httpSrv: h,
	}

	logrusEntry := logrus.NewEntry(logrus.StandardLogger())
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

	opts := make([]grpc.ServerOption, 0)
	opts = append(opts, grpc_middleware.WithUnaryServerChain(unaryInterceptors...))
	opts = append(opts, grpc_middleware.WithStreamServerChain(streamInterceptors...))
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalivePol))

	s.grpcSrv.srv = grpc.NewServer(opts...)
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
		return fmt.Errorf("unable to listen: %v", err)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := s.grpcSrv.srv.Serve(grpcLis)
		logrus.Fatalf("GRPC serving failed: %v", err)
		os.Exit(1)
		wg.Done()
	}(&wg)

	// HTTP
	http.Handle("/metrics", promhttp.Handler())
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := s.httpSrv.ListenAndServe()
		logrus.Fatalf("HTTP serving failed: %v", err)
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
