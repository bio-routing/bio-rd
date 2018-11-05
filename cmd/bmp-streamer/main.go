package main

import (
	"flag"
	"os"

	"github.com/bio-routing/bio-rd/cmd/bmp-streamer/pkg/apiserver"
	"github.com/bio-routing/bio-rd/cmd/bmp-streamer/pkg/config"
	"github.com/bio-routing/bio-rd/lib/grpchelper"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"

	pb "github.com/bio-routing/bio-rd/cmd/bmp-streamer/pkg/bmpstreamer"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	cfgFilePath = flag.String("cfg", "config.yml", "Config file")
)

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	cfg, err := config.LoadConfig(*cfgFilePath)
	if err != nil {
		log.Errorf("Unable to load config: %v", err)
		os.Exit(1)
	}

	bmp := startBMPServer(cfg)
	api := apiserver.New(bmp)

	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	streamInterceptors := []grpc.StreamServerInterceptor{}

	srv, err := grpchelper.New(
		cfg.GRPCPort,
		grpchelper.HTTP(cfg.HTTPPort),
		unaryInterceptors,
		streamInterceptors,
	)
	if err != nil {
		log.Errorf("failed to listen: %v", err)
		os.Exit(1)
	}

	pb.RegisterRIBServiceServer(srv.GRPC(), api)
	grpc_prometheus.Register(srv.GRPC())

	if err := srv.Serve(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func startBMPServer(cfg *config.Config) *server.BMPServer {
	bmp := server.NewServer()

	for _, rtr := range cfg.BMPRouters {
		bmp.AddRouter(rtr.Addr, rtr.Port, nil, nil)
	}

	return bmp
}
