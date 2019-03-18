package main

import (
	"github.com/bio-routing/bio-rd/route"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/fib"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/sirupsen/logrus"
)

func addPath(v *vrf.VRF) {
	pfx := bnet.NewPfx(bnet.IPv4FromOctets(169, 254, 0, 0), uint8(24))
	fibPath := &route.Path{
		Type: route.FIBPathType,
		FIBPath: &route.FIBPath{
			NextHop: bnet.IPv4FromOctets(169, 254, 1, 1),
		},
	}

	rib, found := v.RIBByName("inet.254")
	if !found {
		logrus.Fatal("Unable to find RIB inet.254")
	}

	err := rib.AddPath(pfx, fibPath)
	if err != nil {
		logrus.Errorf("Unable to add Path: Pfx: %s Path: %s", pfx.String(), fibPath.String())
	}
}

func main() {
	v, err := vrf.NewDefaultVRF()
	if err != nil {
		logrus.Fatal(err)
	}

	f, err := fib.New(v)
	if err != nil {
		logrus.Fatal(err)
	}

	err = f.Start()
	if err != nil {
		logrus.Fatalf("Unable to start FIB: %v", err)
	}

	time.Sleep(5 * time.Second)

	addPath(v)
}
