package main

import (
	"os"
	"time"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/fib"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	log "github.com/sirupsen/logrus"
)

func strAddr(s string) uint32 {
	ret, _ := bnet.StrToAddr(s)
	return ret
}

func main() {
	log.SetLevel(log.DebugLevel)

	f, err := os.OpenFile("/var/log/bio-rd.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Info("bio-routing started...\n")

	v, err := vrf.New("master")
	if err != nil {
		log.Fatal(err)
	}

	b := server.NewBgpServer()
	rib := startBGPServer(b, v)

	// FIB communication
	n := fib.NewFIB(&config.Netlink{
		HoldTime:       time.Second * 15,
		UpdateInterval: time.Second * 15,
		RoutingTable:   config.RtMain,
	}, rib)
	n.Start()

	go func() {
		for {
			log.Debugf("LocRIB count: %d", rib.Count())
			log.Debugf(rib.String())
			time.Sleep(time.Second * 10)
		}
	}()

	select {}
}
