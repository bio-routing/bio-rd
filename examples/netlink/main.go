package main

import (
	"net"
	"os"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/netlink"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"

	bnet "github.com/bio-routing/bio-rd/net"
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

	cfg := &config.Global{
		Listen: true,
		LocalAddressList: []net.IP{
			net.IPv4(169, 254, 0, 2),
		},
	}

	rib := locRIB.New()
	b := server.NewBgpServer()
	startBGPServer(b, rib, cfg)

	// Netlink communication
	n := protocolnetlink.NewNetlink(&config.Netlink{
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
