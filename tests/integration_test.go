package tests

import (
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/net"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/net/tcp"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/stretchr/testify/assert"
)

type peer struct {
	config server.PeerConfig
	con    *tcp.MockConn
}

func TestBGP(t *testing.T) {
	b := server.NewBGPServer(server.BGPServerConfig{
		RouterID: net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32(),
	})

	lm := tcp.NewListenerManager(map[string][]string{
		"main": {
			"0.0.0.0:179",
		},
	})

	lm.SetListenerFactory(tcp.NewMockListenerFactory())
	b.SetListenerManager(lm)
	b.Start()

	vrfReg := vrf.NewVRFRegistry()
	mainVRF := vrfReg.CreateVRFIfNotExists(vrf.DefaultVRFName, 0)

	peerA := peer{
		config: server.PeerConfig{
			AdminEnabled: true,
			LocalAS:      100,
			LocalAddress: bnet.IPv4FromOctets(192, 0, 2, 0).Ptr(),
			PeerAS:       200,
			PeerAddress:  bnet.IPv4FromOctets(192, 0, 2, 1).Ptr(),
			Passive:      true,
			VRF:          mainVRF,
			RouterID:     100,
			IPv4: &server.AddressFamilyConfig{
				ImportFilterChain: filter.Chain{filter.NewAcceptAllFilter()},
				ExportFilterChain: filter.Chain{filter.NewAcceptAllFilter()},
			},
		},
	}

	peerB := peer{
		config: server.PeerConfig{
			AdminEnabled: true,
			LocalAS:      100,
			LocalAddress: bnet.IPv4FromOctets(192, 0, 2, 2).Ptr(),
			PeerAS:       255,
			PeerAddress:  bnet.IPv4FromOctets(192, 0, 2, 3).Ptr(),
			Passive:      true,
			VRF:          mainVRF,
			RouterID:     100,
			IPv4: &server.AddressFamilyConfig{
				ImportFilterChain: filter.Chain{filter.NewAcceptAllFilter()},
				ExportFilterChain: filter.Chain{filter.NewAcceptAllFilter()},
			},
		},
	}

	err := b.AddPeer(peerA.config)
	if err != nil {
		t.Errorf("unexpected error adding BGP peer: %v", err)
		return
	}

	l := lm.GetListeners(mainVRF)[0]
	mc := l.(*tcp.MockListener).Connect(bnet.IPv4FromOctets(192, 0, 2, 1).Ptr().ToNetIP(), 31337)
	peerA.con = mc

	openMsg := &packet.BGPOpen{
		Version:       4,
		ASN:           200,
		HoldTime:      90,
		BGPIdentifier: bnet.IPv4FromOctets(192, 0, 2, 1).Ptr().ToUint32(),
	}

	peerA.con.WriteFromOtherEnd(packet.SerializeOpenMsg(openMsg))

	pkt := peerA.con.ReadFromOtherEnd()
	if !assert.Equal(t, []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x2b, // Length
		0x01,       // Type
		0x04,       // Version,
		0x00, 0x64, // ASN
		0x00, 0x00, // Hold time
		0x00, 0x00, 0x00, 0x64, // BGP Identifier
		0x0e, // Opt param length
		0x02, 0xc, 0x45, 0x4, 0x0, 0x1, 0x1, 0x2, 0x41, 0x4, 0x0, 0x0, 0x0, 0x64,
	}, pkt) {
		return
	}

	keepaliveMsg := []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		00, 19, // Length
		0x04, // Type Keepalive
	}
	peerA.con.WriteFromOtherEnd(keepaliveMsg)

	peerA.con.WriteFromOtherEnd([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		00, 27, // Length
		0x02, // Update
		0x00, 0x04,
		24, 10, 0, 0, // withdraw 10.0.0.0/24
		0x00, 0x00,
	})

	peerA.con.WriteFromOtherEnd([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		00, 45, // Length
		0x02,       // Update
		0x00, 0x00, // Withdraw length
		0x00, 20, // attributes length
		0x40, 0x01, 0x01, 0x00, // Origin
		0x40, 0x02, 0x06, 0x02, 0x02, 0x00, 123, 0x00, 250, // AS Path
		0x00, 0x03, 0x04, 0x10, 0x11, 0x12, 0x13, // Next Hop
		0x08, 10, // 10.0.0.0/8
	})

	_ = peerB

	time.Sleep(time.Second)

	r := mainVRF.IPv4UnicastRIB().Get(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8).Ptr())
	if r == nil {
		t.Errorf("route 10.0.0.0/8 not found")
		return
	}

	b.AddPeer(peerB.config)
	if err != nil {
		t.Errorf("unexpected error adding BGP peer: %v", err)
		return
	}

	peerB.con = l.(*tcp.MockListener).Connect(bnet.IPv4FromOctets(192, 0, 2, 3).Ptr().ToNetIP(), 31337)
	openMsg = &packet.BGPOpen{
		Version:       4,
		ASN:           255,
		HoldTime:      90,
		BGPIdentifier: bnet.IPv4FromOctets(192, 0, 2, 3).Ptr().ToUint32(),
	}
	peerB.con.WriteFromOtherEnd(packet.SerializeOpenMsg(openMsg))

	pkt = peerB.con.ReadFromOtherEnd()
	if !assert.Equal(t, []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x2b, // Length
		0x01,       // Type
		0x04,       // Version,
		0x00, 0x64, // ASN
		0x00, 0x00, // Hold time
		0x00, 0x00, 0x00, 0x64, // BGP Identifier
		0x0e, // Opt param length
		0x02, 0xc, 0x45, 0x4, 0x0, 0x1, 0x1, 0x2, 0x41, 0x4, 0x0, 0x0, 0x0, 0x64,
	}, pkt) {
		return
	}

	peerB.con.WriteFromOtherEnd(keepaliveMsg)

	pkt = peerB.con.ReadFromOtherEnd()
	if !assert.Equal(t, []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 19, 4}, pkt) {
		return
	}

	pkt = peerB.con.ReadFromOtherEnd()
	if !assert.Equal(t, []byte{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 47,
		2,
		0, 0,
		0, 22,
		64, 2, 8, 2, 3, 0, 100, 0, 123, 0, 250,
		64, 1, 1, 0,
		64, 3, 4, 192, 0, 2, 2,
		8, 10, // 10.0.0.0/8
	}, pkt) {
		return
	}
}
