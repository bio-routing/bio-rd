package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"net/http"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/metrics"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logrus.Printf("This is a BGP speaker\n")

	rib := locRIB.New()
	b := server.NewBgpServer()
	metrics.RegisterMetrics(prometheus.DefaultRegisterer)
	err := b.Start(&config.Global{
		Listen: true,
	})
	if err != nil {
		logrus.Fatalf("Unable to start BGP server: %v", err)
	}

	b.AddPeer(config.Peer{
		AdminEnabled: true,
		LocalAS:      65200,
		PeerAS:       65300,
		PeerAddress:  net.IP([]byte{169, 254, 200, 1}),
		LocalAddress: net.IP([]byte{169, 254, 200, 0}),
		HoldTimer:    90,
		KeepAlive:    30,
		Passive:      true,
		RouterID:     b.RouterID(),
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
	}, rib)

	time.Sleep(time.Second * 15)

	b.AddPeer(config.Peer{
		AdminEnabled: true,
		LocalAS:      65200,
		PeerAS:       65100,
		PeerAddress:  net.IP([]byte{169, 254, 100, 0}),
		LocalAddress: net.IP([]byte{169, 254, 100, 1}),
		HoldTimer:    90,
		KeepAlive:    30,
		Passive:      true,
		RouterID:     b.RouterID(),
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
		AddPathRecv: true,
	}, rib)

	go func() {
		for {
			fmt.Print(rib.Print())
			time.Sleep(time.Second * 10)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		httpErr := http.ListenAndServe("0.0.0.0:8080", http.DefaultServeMux)
		logrus.WithError(httpErr).Error("stopped http metrics endpoint.")
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
