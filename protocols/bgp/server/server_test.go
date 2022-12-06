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
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	bgpserver "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

func TestInMemoryServerMinimalConnection(t *testing.T) {
	ip1 := bnet.IPv4FromOctets(172, 16, 0, 1)
	as1 := uint32(65101)
	ip2 := bnet.IPv4FromOctets(172, 16, 0, 2)
	as2 := uint32(65102)

	s1, vrfs1, err := startServer(ip1)
	if err != nil {
		t.Fatalf("Unable to start server: %v", err)
	}

	s2, vrfs2, err := startServer(ip2)
	if err != nil {
		t.Fatalf("Unable to start server: %v", err)
	}

	ml1 := addListener(s1, ip1)
	ml2 := addListener(s2, ip2)

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

	// Wait for connection to be established
	assert.Eventually(t, func() bool {
		return s1.GetPeerStatus(ip2.Ptr()) == "established" && s2.GetPeerStatus(ip1.Ptr()) == "established"
	}, time.Second*10, time.Millisecond*100, "Peer not established")

	assert.Zero(t, s1.GetRIBIn(ip2.Ptr(), packet.AFIIPv4, packet.SAFIUnicast).RouteCount())
	assert.Zero(t, s2.GetRIBIn(ip1.Ptr(), packet.AFIIPv4, packet.SAFIUnicast).RouteCount())
}

func TestInMemoryServerMinimalPaths(t *testing.T) {
	ip1 := bnet.IPv4FromOctets(172, 16, 0, 1)
	as1 := uint32(65101)
	ip2 := bnet.IPv4FromOctets(172, 16, 0, 2)
	as2 := uint32(65102)
	ip3 := bnet.IPv4FromOctets(172, 16, 0, 3)
	as3 := uint32(65103)

	s1, vrfs1, err := startServer(ip1)
	if err != nil {
		t.Fatalf("Unable to start server: %v", err)
	}

	s2, vrfs2, err := startServer(ip2)
	if err != nil {
		t.Fatalf("Unable to start server: %v", err)
	}

	s3, vrfs3, err := startServer(ip3)
	if err != nil {
		t.Fatalf("Unable to start server: %v", err)
	}

	ml1 := addListener(s1, ip1)
	ml2 := addListener(s2, ip2)
	ml3 := addListener(s3, ip3)

	if err := connect(s1, s2, ip1, ip2, vrfs1, vrfs2, as1, as2, ml1, ml2); err != nil {
		t.Fatalf("Unable to connect s1 and s2: %v", err)
	}

	if err := connect(s2, s3, ip2, ip3, vrfs2, vrfs3, as2, as3, ml2, ml3); err != nil {
		t.Fatalf("Unable to connect s2 and s3: %v", err)
	}

	pfx := bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8)

	if err := vrfs1.GetVRFByName("master").IPv4UnicastRIB().AddPath(pfx.Ptr(), generatePath(ip1.Ptr(), ip1.Ptr())); err != nil {
		t.Fatalf("Unable to add path: %v", err)
	}

	// Wait for connection to be established
	assert.Eventually(t, func() bool {
		return s1.GetPeerStatus(ip2.Ptr()) == "established" &&
			s2.GetPeerStatus(ip1.Ptr()) == "established" &&
			s2.GetPeerStatus(ip3.Ptr()) == "established" &&
			s3.GetPeerStatus(ip2.Ptr()) == "established"
	}, time.Second*10, time.Millisecond*100, "Peer not established")

	assert.EventuallyWithT(t, func(collect *assert.CollectT) {
		routes := vrfs2.GetVRFByName("master").IPv4UnicastRIB().LPM(pfx.Ptr())

		assert.Equalf(t, 1, len(routes), "Expected 1 route, got %d", len(routes))
		assert.Equalf(t, 1, len(routes[0].Paths()), "Expected 1 path, got %d", len(routes[0].Paths()))
		t.Log(routes[0].Paths())

		path := routes[0].Paths()[0]
		assert.True(t, path.NextHop().Equal(ip1), "Expected next hop to be %s, got %s", ip1, path.NextHop())
	}, time.Second*3, time.Millisecond*500, "Route not received")

	assert.EventuallyWithT(t, func(collect *assert.CollectT) {
		routes := vrfs3.GetVRFByName("master").IPv4UnicastRIB().LPM(pfx.Ptr())

		assert.Equalf(t, 1, len(routes), "Expected 1 route, got %d", len(routes))
		assert.Equalf(t, 1, len(routes[0].Paths()), "Expected 1 path, got %d", len(routes[0].Paths()))
		t.Log(routes[0].Paths())

		path := routes[0].Paths()[0]
		assert.True(t, path.NextHop().Equal(ip1), "Expected next hop to be %s, got %s", ip1, path.NextHop())
	}, time.Second*3, time.Millisecond*500, "Route not received")
}

func startServer(routerID bnet.IP) (bgpserver.BGPServer, *vrf.VRFRegistry, error) {
	s := bgpserver.NewBGPServer(uint32(routerID.Lower()))

	if err := s.Start(); err != nil {
		return nil, nil, fmt.Errorf("Unable to start BGP server: %v", err)
	}

	vrfs := vrf.NewVRFRegistry()

	return s, vrfs, nil
}

func addListener(s bgpserver.BGPServer, routerID bnet.IP) *fasthttputil.InmemoryListener {
	ml := fasthttputil.NewInmemoryListener()
	ml.SetLocalAddr(&net.TCPAddr{IP: routerID.ToNetIP()})

	s.AddListener(ml)

	return ml
}

func connect(r1, r2 bgpserver.BGPServer, ip1, ip2 bnet.IP, vrfs1, vrfs2 *vrf.VRFRegistry, as1, as2 uint32, ml1, ml2 *fasthttputil.InmemoryListener) error {
	if err := r1.AddPeer(
		bgpPeerConfig(r1,
			bnet.IPv4FromBytes(ml1.Addr().(*net.TCPAddr).IP.To4()).Ptr(),
			bnet.IPv4FromBytes(ml2.Addr().(*net.TCPAddr).IP.To4()).Ptr(),
			as1,
			as2,
			vrfs1.CreateVRFIfNotExists("master", 0),
			ml2,
		),
	); err != nil {
		return fmt.Errorf("Unable to add peer: %v", err)
	}

	pc := bgpPeerConfig(r2,
		bnet.IPv4FromBytes(ml2.Addr().(*net.TCPAddr).IP.To4()).Ptr(),
		bnet.IPv4FromBytes(ml1.Addr().(*net.TCPAddr).IP.To4()).Ptr(),
		as2,
		as1,
		vrfs2.CreateVRFIfNotExists("master", 0),
		ml1,
	)
	// pc.Passive = true

	if err := r2.AddPeer(pc); err != nil {
		return fmt.Errorf("Unable to add peer: %v", err)
	}

	return nil
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
		},
		VRF:       vrf,
		TCPDialer: ml,
	}
}

func generatePath(source, nh *bnet.IP) *route.Path {
	p := &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				Source:  source,
				NextHop: nh,
				EBGP:    true,
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
