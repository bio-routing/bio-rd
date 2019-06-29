package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	prom_bgp "github.com/bio-routing/bio-rd/metrics/bgp/adapter/prom"
	prom_vrf "github.com/bio-routing/bio-rd/metrics/vrf/adapter/prom"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
)

func strAddr(s string) uint32 {
	ret, _ := bnet.StrToAddr(s)
	return ret
}

func main() {
	logrus.Printf("This is a BGP speaker\n")

	b := server.NewBgpServer()
	v, err := vrf.New("master", 0)
	if err != nil {
		log.Fatal(err)
	}

	go startMetricsEndpoint(b)

	startServer(b, v)

	select {}
}

func startMetricsEndpoint(server server.BGPServer) {
	prometheus.MustRegister(prom_bgp.NewCollector(server))
	prometheus.MustRegister(prom_vrf.NewCollector(vrf.GetGlobalRegistry()))

	http.Handle("/metrics", promhttp.Handler())

	logrus.Info("Metrics are available :8080/metrics")
	logrus.Error(http.ListenAndServe(":8080", nil))
}
