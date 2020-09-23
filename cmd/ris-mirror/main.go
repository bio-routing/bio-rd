package main

import (
	"flag"
	"os"
	"time"

	"github.com/bio-routing/bio-rd/cmd/ris-mirror/config"
	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror"
	pb "github.com/bio-routing/bio-rd/cmd/ris/api"
	"github.com/bio-routing/bio-rd/cmd/ris/risserver"
	prom_grpc_cm "github.com/bio-routing/bio-rd/metrics/grpc/clientmanager/adapter/prom"
	prom_ris_mirror "github.com/bio-routing/bio-rd/metrics/ris-mirror/adapter/prom"
	"github.com/bio-routing/bio-rd/util/grpc/clientmanager"
	"github.com/bio-routing/bio-rd/util/servicewrapper"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	grpcPort             = flag.Uint("grpc_port", 4321, "gRPC server port")
	httpPort             = flag.Uint("http_port", 4320, "HTTP server port")
	grpcKeepaliveMinTime = flag.Uint("grpc_keepalive_min_time", 1, "Minimum time (seconds) for a client to wait between GRPC keepalive pings")
	risTimeout           = flag.Uint("ris_timeout", 5, "RIS timeout in seconds")
	configFilePath       = flag.String("config.file", "ris_mirror.yml", "Configuration file")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configFilePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to load config")
	}

	grpcClientManager := clientmanager.New()
	for _, instance := range cfg.GetRISInstances() {
		err := grpcClientManager.AddIfNotExists(instance, grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			Timeout:             time.Second * time.Duration(*risTimeout),
			PermitWithoutStream: true,
		}))

		if err != nil {
			log.WithError(err).Fatal("GRPC clientmanager add failed")
		}
	}

	m := rismirror.New()
	prometheus.MustRegister(prom_ris_mirror.NewCollector(m))
	prometheus.MustRegister(prom_grpc_cm.NewCollector(grpcClientManager))

	for _, rcfg := range cfg.RIBConfigs {
		for _, vrdRD := range rcfg.GetVRFs() {
			srcs := make([]*grpc.ClientConn, 0)
			for _, srcInstance := range rcfg.SrcRISInstances {
				srcs = append(srcs, grpcClientManager.Get(srcInstance))
			}

			m.AddTarget(rcfg.Router, rcfg.GetRouter(), vrdRD, srcs)
		}
	}

	s := risserver.NewServer(m)
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
