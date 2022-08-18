package main

import (
	"flag"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/bio-routing/bio-rd/cmd/ris/config"
	"github.com/bio-routing/bio-rd/cmd/ris/risserver"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/util/servicewrapper"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/keepalive"

	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	prom_bmp "github.com/bio-routing/bio-rd/metrics/bmp/adapter/prom"
	log "github.com/sirupsen/logrus"
)

var (
	grpcPort             = flag.Uint("grpc_port", 4321, "gRPC server port")
	httpPort             = flag.Uint("http_port", 4320, "HTTP server port")
	bmpListenAddr        = flag.String("bmp_addr", "0.0.0.0:30119", "BMP listen addr (set empty to disable listening)")
	grpcKeepaliveMinTime = flag.Uint("grpc_keepalive_min_time", 1, "Minimum time (seconds) for a client to wait between GRPC keepalive pings")
	configFilePath       = flag.String("config.file", "", "Configuration file")
	tcpKeepaliveInterval = flag.Uint("tcp-keepalive-interval", 1, "TCP keepalive interval (seconds)")
	allowAny             = flag.Bool("allow.any", false, "Allow BMP sessions from anywhere")
)

func main() {
	flag.Parse()

	cfg := &config.RISConfig{}
	if *configFilePath != "" {
		c, err := config.LoadConfig(*configFilePath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		cfg = c
	}

	b := server.NewServer(server.BMPServerConfig{
		KeepalivePeriod: time.Duration(*tcpKeepaliveInterval) * time.Second,
		AcceptAny:       *allowAny,
	})

	if *bmpListenAddr != "" {
		go func() {
			if err := b.Listen(*bmpListenAddr); err != nil {
				log.WithError(err).Error("error while starting listener")
			}
		}()
	}
	defer b.Close()

	prometheus.MustRegister(prom_bmp.NewCollector(b))

	for _, r := range cfg.BMPServers {
		ip := net.ParseIP(r.Address)
		if ip == nil {
			log.Fatalf("unable to convert %q to net.IP", r.Address)
		}
		b.AddRouter(ip, r.Port, r.Passive, false)
	}

	s := risserver.NewServer(b)
	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	streamInterceptors := []grpc.StreamServerInterceptor{}
	srv, err := servicewrapper.New(
		uint16(*grpcPort),
		servicewrapper.HTTP(uint16(*httpPort)),
		unaryInterceptors,
		streamInterceptors,
		keepalive.EnforcementPolicy{
			MinTime:             time.Duration(*grpcKeepaliveMinTime) * time.Second,
			PermitWithoutStream: true,
		},
	)
	if err != nil {
		log.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}

	pb.RegisterRoutingInformationServiceServer(srv.GRPC(), s)
	if err := srv.Serve(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
