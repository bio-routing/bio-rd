package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	prom_bgp "github.com/bio-routing/bio-rd/metrics/bgp/adapter/prom"
	prom_vrf "github.com/bio-routing/bio-rd/metrics/vrf/adapter/prom"
	api "github.com/bio-routing/bio-rd/protocols/bgp/api"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
)

func main() {
	ipv4 := flag.Bool("ipv4", true, "Enable IPv4 Listener and AFI")
	ipv6 := flag.Bool("ipv6", true, "Enable IPv6 Listener and AFI")
	flag.Parse()

	logrus.Printf("This is a BGP speaker\n")

	listen := []string{}

	if *ipv4 {
		listen = append(listen, "0.0.0.0:179")
	}

	if *ipv6 {
		listen = append(listen, "[::]:179")
	}

	b := server.NewBGPServer(0, listen)
	v, err := vrf.New("master", 0)
	if err != nil {
		logrus.Fatal(err)
	}

	go startMetricsEndpoint(b)
	go startAPIEndpoint(b)

	if err := b.Start(); err != nil {
		log.Fatalf("Unable to start BGP server: %v", err)
	}

	if *ipv4 {
		addPeersIPv4(b, v)
	}

	if *ipv6 {
		addPeersIPv6(b, v)
	}

	select {}
}

func startMetricsEndpoint(server server.BGPServer) {
	prometheus.MustRegister(prom_bgp.NewCollector(server))
	prometheus.MustRegister(prom_vrf.NewCollector(vrf.GetGlobalRegistry()))

	http.Handle("/metrics", promhttp.Handler())

	logrus.Info("Metrics are available :8080/metrics")
	logrus.Error(http.ListenAndServe(":8080", nil))
}

func startAPIEndpoint(b server.BGPServer) {
	apiSrv := server.NewBGPAPIServer(b)

	lis, err := net.Listen("tcp", ":1337")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	api.RegisterBgpServiceServer(grpcServer, apiSrv)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
