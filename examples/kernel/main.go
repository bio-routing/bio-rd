package main

import (
	"os"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/kernel"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	log "github.com/sirupsen/logrus"
)

func main() {
	vrf, err := vrf.New("inet.0", 0)
	if err != nil {
		log.Errorf("Unable to create VRF: %v", err)
		os.Exit(1)
	}

	rib4 := vrf.IPv4UnicastRIB()
	rib4.AddPath(bnet.NewPfx(bnet.IPv4FromOctets(8, 8, 8, 0), 24), &route.Path{
		Type: route.StaticPathType,
		StaticPath: &route.StaticPath{
			NextHop: bnet.IPv4FromOctets(127, 0, 0, 1),
		},
	})

	k, err := kernel.New()
	if err != nil {
		log.Errorf("Unable to create protocol kernel: %v", err)
		os.Exit(1)
	}
	defer k.Dispose()

	rib4.Register(k)

	time.Sleep(time.Second * 10)
}
