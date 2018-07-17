package server

import (
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestBgpServerPeerSnapshot(t *testing.T) {
	s := NewBgpServer()
	err := s.Start(&config.Global{
		LocalASN: 204880,
		RouterID: 2137,
	})
	if err != nil {
		t.Fatalf("server should have started, got err: %v", err)
	}

	info := s.GetPeerInfoAll()
	if len(info) != 0 {
		t.Fatalf("empty server should have 0 peers, has %d", len(info))
	}

	rib := locRIB.New()
	pc := config.Peer{
		AdminEnabled:      true,
		PeerAS:            65300,
		PeerAddress:       bnet.IPv4FromOctets(169, 254, 200, 1),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 200, 0),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          s.RouterID(),
		AddPathSend: routingtable.ClientOptions{
			MaxPaths: 10,
		},
		IPv4: &config.AddressFamilyConfig{
			RIB:          rib,
			ImportFilter: filter.NewDrainFilter(),
			ExportFilter: filter.NewAcceptAllFilter(),
		},
	}
	s.AddPeer(pc)

	info = s.GetPeerInfoAll()
	if want, got := 1, len(info); want != got {
		t.Fatalf("empty server should have %d peers, has %d", want, got)
	}

	var peer PeerInfo
	for _, v := range info {
		peer = v
		break
	}

	if want, got := bnet.IPv4FromOctets(169, 254, 200, 1), peer.PeerAddr; !want.Equal(got) {
		t.Errorf("PeerAddr: got %v, want %v", got, want)
	}
	if want, got := uint32(65300), peer.PeerASN; want != got {
		t.Errorf("PeerASN: got %v, want %v", got, want)
	}
	if want, got := uint32(204880), peer.LocalASN; want != got {
		t.Errorf("PeerASN: got %v, want %v", got, want)
	}
}
