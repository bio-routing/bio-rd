package server_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp/fasthttputil"

	bnet "github.com/bio-routing/bio-rd/net"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

func TestServerInMemory(t *testing.T) {
	ip1 := bnet.IPv4FromOctets(172, 17, 0, 3)
	ip2 := bnet.IPv4FromOctets(169, 254, 200, 0)

	s1, ml1, vrfs1, err := startServer(ip1)
	if err != nil {
		t.Fatalf("Unable to start server 1: %v", err)
	}
	s2, ml2, vrfs2, err := startServer(ip2)
	if err != nil {
		t.Fatalf("Unable to start server 2: %v", err)
	}

	as1 := uint32(65100)
	as2 := uint32(65101)

	if err := s1.AddPeer(
		bgpPeerConfig(s1, ip1.Ptr(), ip2.Ptr(), as1, as2,
			vrfs1.CreateVRFIfNotExists("master", 0),
			ml2,
		),
	); err != nil {
		logrus.Fatalf("Unable to add peer: %v", err)
	}

	pc := bgpPeerConfig(s2, ip2.Ptr(), ip1.Ptr(), as2, as1,
		vrfs2.CreateVRFIfNotExists("master", 0),
		ml1,
	)
	pc.Passive = true

	if err := s2.AddPeer(pc); err != nil {
		t.Fatalf("Unable to add peer: %v", err)
	}

	master := vrfs1.GetVRFByName("master")
	locRIB := master.IPv4UnicastRIB()

	pfx := bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8)

	if err := locRIB.AddPath(pfx.Ptr(), generatePath(ip1.Ptr(), ip1.Ptr())); err != nil {
		t.Fatalf("Unable to add path: %v", err)
	}

	assert.Eventually(t, func() bool {
		routes := vrfs2.GetVRFByName("master").IPv4UnicastRIB().LPM(pfx.Ptr())

		if len(routes) != 1 {
			return false
		}

		if len(routes[0].Paths()) != 1 {
			return false
		}

		path := routes[0].Paths()[0]

		return path.NextHop().Equal(ip1) &&
			path.BGPPath.BGPPathA.LocalPref == 0 &&
			path.BGPPath.BGPPathA.MED == 200
	}, time.Second*15, time.Millisecond*100, "Route not received")
}

func startServer(routerID bnet.IP) (bgpserver.BGPServer, *fasthttputil.InmemoryListener, *vrf.VRFRegistry, error) {
	ml := fasthttputil.NewInmemoryListener()
	ml.SetLocalAddr(&net.TCPAddr{IP: routerID.ToNetIP()})

	s := bgpserver.NewBGPServer(uint32(routerID.Lower()))

	s.AddListener(ml)

	if err := s.Start(); err != nil {
		return nil, nil, nil, fmt.Errorf("Unable to start BGP server: %v", err)
	}

	vrfs := vrf.NewVRFRegistry()

	return s, ml, vrfs, nil
}

func bgpPeerConfig(s bgpserver.BGPServer, localAddr, peerAddr *bnet.IP, localAS, peerAS uint32, vrf *vrf.VRF, ml *fasthttputil.InmemoryListener) bgpserver.PeerConfig {
	return bgpserver.PeerConfig{
		LocalAddress:      localAddr,
		LocalAS:           localAS,
		PeerAddress:       peerAddr,
		PeerAS:            peerAS,
		ReconnectInterval: time.Second * 1,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 90 / 3,
		RouterID:          s.RouterID(),
		IPv4: &bgpserver.AddressFamilyConfig{
			ImportFilterChain: filter.NewAcceptAllFilterChain(),
			ExportFilterChain: filter.NewAcceptAllFilterChain(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
		},
		VRF:               vrf,
		TCPDialer:         ml,
		AuthenticationKey: "test",
	}
}

func generatePath(source, nh *bnet.IP) *route.Path {
	p := &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				Source:    source,
				NextHop:   nh,
				LocalPref: 100,
				MED:       200,
				EBGP:      true,
			},
			ASPath: &types.ASPath{
				types.ASPathSegment{
					Type: types.ASSequence,
					ASNs: []uint32{},
				},
			},
		},
	}

	return p
}
