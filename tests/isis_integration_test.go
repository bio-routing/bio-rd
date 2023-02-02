package tests

import (
	"sort"
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"

	bbclock "github.com/benbjohnson/clock"
	bnet "github.com/bio-routing/bio-rd/net"
)

const testTimeLayout = "Jan 2, 2006 at 15:04:05.000"

type neighbor struct {
	mac         ethernet.MACAddr
	systemID    types.SystemID
	interfaceIP bnet.IP
}

var (
	allIntermediateSystems = ethernet.MACAddr{
		0x09, 0x00, 0x2b, 0x00, 0x00, 0x05,
	}
	adjacencyNeighborADown = &server.Adjacency{
		Name:          "",
		SystemID:      types.SystemID{0xde, 0xad, 0xbe, 0xef, 0xff, 0x01},
		Address:       ethernet.MACAddr{0xde, 0xad, 0xbe, 0xef, 0x12, 0x34},
		InterfaceName: "eth0",
		Level:         2,
		Priority:      0,
		IPAddresses: []bnet.IP{
			bnet.IPv4FromOctets(169, 254, 100, 1),
		},
		Status: packet.P2PAdjStateDown,
	}
	adjacencyNeighborAUp = &server.Adjacency{
		Name:          "",
		SystemID:      types.SystemID{0xde, 0xad, 0xbe, 0xef, 0xff, 0x01},
		Address:       ethernet.MACAddr{0xde, 0xad, 0xbe, 0xef, 0x12, 0x34},
		InterfaceName: "eth0",
		Level:         2,
		Priority:      0,
		IPAddresses: []bnet.IP{
			bnet.IPv4FromOctets(169, 254, 100, 1),
		},
		Status: packet.P2PAdjStateUp,
	}
	adjacencyNeighborAInit = &server.Adjacency{
		Name:          "",
		SystemID:      types.SystemID{0xde, 0xad, 0xbe, 0xef, 0xff, 0x01},
		Address:       ethernet.MACAddr{0xde, 0xad, 0xbe, 0xef, 0x12, 0x34},
		InterfaceName: "eth0",
		Level:         2,
		Priority:      0,
		IPAddresses: []bnet.IP{
			bnet.IPv4FromOctets(169, 254, 100, 1),
		},
		Status: packet.P2PAdjStateInit,
	}
	helloFromNeighborADown = []byte{
		0x00, // DSAP
		0x00, // CSAP
		0x00, // CF
		// ISO 10589 header
		131, // Intradomain Routing Protocol Discriminator: ISIS
		20,  // Length indicator
		1,   // Version / Protocol ID Extension
		0,   // ID Length
		17,  // Type
		1,   // Version
		0,   // Reserved
		0,   // Maximum Area Addresses
		// ISIS hello
		2,                          // Level 2 only
		222, 173, 190, 239, 255, 1, // System ID
		0, 16, // Holding timer
		0, 42, // PDU length
		2, // Local Circuit ID
		// P2P Adj. State TLV
		240, // Type
		5,   // Length
		2,   // Adj State down
		0, 0, 0, 100,
		// Protocols supported TLV
		129, // Type
		2,   // Length
		204, // IPv4
		142, // IPv6
		// IP Interface addresses TLV
		132,              // Type
		4,                // Length
		169, 254, 100, 1, // IP Address
		// Area Addresses TLV
		1,       // Type
		3,       // Length
		2,       // Area length
		0x49, 0, // Area
	}
	helloFromNeighborAInit = []byte{
		// ISO 10589 header
		131, // Intradomain Routing Protocol Discriminator: ISIS
		20,  // Length indicator
		1,   // Version / Protocol ID Extension
		0,   // ID Length
		17,  // Type
		1,   // Version
		0,   // Reserved
		0,   // Maximum Area Addresses
		// ISIS hello
		2,                      // Level 2 only
		12, 12, 12, 13, 13, 13, // System ID
		0, 16, // Holding timer
		0, 52, // PDU length
		1, // Local Circuit ID
		// P2P Adj. State TLV <--- Important part
		240,        // Type
		15,         // Length
		1,          // Adj State Init
		0, 0, 0, 0, // extended local circuit id
		222, 173, 190, 239, 255, 1, // Neighbor system ID
		0, 0, 0, 100, // Neighbor extended local circuit id
		// Protocols supported TLV
		129, // Type
		2,   // Length
		204, // IPv4
		142, // IPv6
		// IP Interface addresses TLV
		132,              // Type
		4,                // Length
		169, 254, 100, 0, // IP Address
		// Area Addresses TLV
		1,       // Type
		3,       // Length
		2,       // Area length
		0x49, 0, // Area
	}
	helloFromNeighborAUp = []byte{
		0x00, // DSAP
		0x00, // CSAP
		0x00, // CF
		// ISO 10589 header
		131, // Intradomain Routing Protocol Discriminator: ISIS
		20,  // Length indicator
		1,   // Version / Protocol ID Extension
		0,   // ID Length
		17,  // Type
		1,   // Version
		0,   // Reserved
		0,   // Maximum Area Addresses
		// ISIS hello
		2,                          // Level 2 only
		222, 173, 190, 239, 255, 1, // System ID
		0, 16, // Holding timer
		0, 52, // PDU length
		1, // Local Circuit ID
		// P2P Adj. State TLV <--- Important part
		240,          // Type
		15,           // Length
		2,            // Adj State up
		0, 0, 0, 100, // extended local circuit id
		12, 12, 12, 13, 13, 13, // Neighbor system ID
		0, 0, 0, 0, // Neighbor extended local circuit id
		// Protocols supported TLV
		129, // Type
		2,   // Length
		204, // IPv4
		142, // IPv6
		// IP Interface addresses TLV
		132,              // Type
		4,                // Length
		169, 254, 100, 1, // IP Address
		// Area Addresses TLV
		1,       // Type
		3,       // Length
		2,       // Area length
		0x49, 0, // Area
	}
	helloToNeighborADown = []byte{
		// ISO 10589 header
		131, // Intradomain Routing Protocol Discriminator: ISIS
		20,  // Length indicator
		1,   // Version / Protocol ID Extension
		0,   // ID Length
		17,  // Type
		1,   // Version
		0,   // Reserved
		0,   // Maximum Area Addresses
		// ISIS hello
		2,                      // Level 2 only
		12, 12, 12, 13, 13, 13, // System ID
		0, 16, // Holding timer
		0, 42, // PDU length
		1, // Local Circuit ID
		// P2P Adj. State TLV
		240, // Type
		5,   // Length
		2,   // Adj State down
		0, 0, 0, 0,
		// Protocols supported TLV
		129, // Type
		2,   // Length
		204, // IPv4
		142, // IPv6
		// IP Interface addresses TLV
		132,              // Type
		4,                // Length
		169, 254, 100, 0, // IP Address
		// Area Addresses TLV
		1,       // Type
		3,       // Length
		2,       // Area length
		0x49, 0, // Area
	}
	lspLocalInitial = []byte{
		// Header
		131,  // Intradomain Routing Protocol Discriminator: ISIS
		0x1b, // Length indicator
		1,    // Version
		0,    // ID Length
		20,   // Type = LSP
		1,    // Version
		0,    // Reserved
		0,    // Max. Area addresses
		// LSP
		0, 0x5f, // Length
		7, 6, // Remaining Lifetime
		12, 12, 12, 13, 13, 13, 0, 0, // LSP ID
		0, 0, 0, 2, // Sequence number
		0x54, 0xc9, // Checksum
		0, // Type block
		// TLVs
		1, // Area
		2, // Length
		1, 0,
		129,      // Protocols Supported
		2,        // Length
		204, 142, // IPv4 + IPv6
		132,              // IP interface addresses
		4,                // Length
		169, 254, 100, 0, // IP
		0x87,               // Extended IP reachability TLV
		0x9,                // Length
		0x0, 0x0, 0x0, 0xa, // Metric
		0x1f,                  // Prefix length
		0xa9, 0xfe, 0x64, 0x0, // IP address

		0x16,                              // Extended IS Reachability TLV
		0x21,                              // Length
		0xde, 0xad, 0xbe, 0xef, 0xff, 0x1, // IS Neighbor ID
		0x0, 0x0, 0x0, 0xa, // Metric
		0x16, // Sub TLV length
		0x6,  // IPv4 ainterface address sub TLV
		0x4,  // length
		0xa9, 0xfe, 0x64, 0x0,

		0x8, // IPv4 neighbor address sub TLV
		0x4, // length
		0xa9, 0xfe, 0x64, 0x1,

		0x4,                // Link local/remote identifiers sub TLV
		0x8,                // length
		0x0, 0x0, 0x0, 0x0, // Local id
		0x0, 0x0, 0x0, 100, // Remote id

		0x89, // Hostname
		6,    // Length
		0x66, 0x75, 0x63, 0x6b, 0x75, 0x70,
	}
)

func TestISISServer(t *testing.T) {
	hostnameFunc := func() (string, error) {
		return "fuckup", nil
	}

	clock := bbclock.NewMock()
	start, _ := time.Parse(testTimeLayout, "January 23, 2023 at 00:00:00.000")
	clock.Set(start)
	server.SetClock(clock)

	neighborA := neighbor{
		mac:         ethernet.MACAddr{0xde, 0xad, 0xbe, 0xef, 0x12, 0x34},
		systemID:    types.SystemID{0xde, 0xad, 0xbe, 0xef, 0xff, 0x01},
		interfaceIP: bnet.IPv4FromOctets(169, 254, 100, 1),
	}

	du := &device.MockServer{}
	s, err := server.New([]*types.NET{
		{
			AFI:      0x49,
			AreaID:   types.AreaID{0x00},
			SystemID: types.SystemID{12, 12, 12, 13, 13, 13},
			SEL:      0x00,
		},
	}, du, 3600)

	if err != nil {
		t.Errorf("unexpected failure creating IS-IS server: %v", err)
		return
	}

	s.Start()
	s.SetEthernetInterfaceFactory(ethernet.NewMockEthernetInterfaceFactory())
	s.SetHostnameFunc(hostnameFunc)

	err = s.AddInterface(&server.InterfaceConfig{
		Name:         "eth0",
		Passive:      false,
		PointToPoint: true,
		Level2: &server.InterfaceLevelConfig{
			HelloInterval: 4,
			HoldingTimer:  16,
			Metric:        10,
			Passive:       false,
		},
	})

	if err != nil {
		t.Errorf("unexpected failure while adding interface: %v", err)
		return
	}

	du.DeviceUpEvent("eth0", []*bnet.Prefix{
		bnet.NewPfx(bnet.IPv4FromOctets(169, 254, 100, 0), 31).Ptr(),
	})

	eth0 := s.GetEthernetInterface("eth0").(*ethernet.MockEthernetInterface)
	clock.Add(time.Second * 4)
	dstMacAddr, pkt := eth0.ReceiveAtRemote()
	if !assert.Equal(t, helloToNeighborADown, pkt) {
		return
	}

	if !assert.Equal(t, allIntermediateSystems, dstMacAddr) {
		return
	}

	// send a hello from neighborA
	eth0.SendFromRemote(neighborA.mac, helloFromNeighborADown)
	time.Sleep(time.Millisecond * 10)
	// checking if the adjancency exists
	if !assertAdjacencies([]*server.Adjacency{
		adjacencyNeighborAInit,
	}, s.GetAdjacencies()) {
		t.Errorf("Adjacency table mismatch")
		return
	}

	// let's see if the neighborA is not listed in the hello packet
	clock.Add(time.Second * 4)
	pkt = readNext(packet.P2P_HELLO, eth0)
	if !assert.Equal(t, helloFromNeighborAInit, pkt) {
		return
	}

	// lets send a hello from neighborA that contains the other router
	eth0.SendFromRemote(neighborA.mac, helloFromNeighborAUp)
	time.Sleep(time.Millisecond * 10)

	clock.Add(time.Second)
	// checking if the adjancency is up
	if !assertAdjacencies([]*server.Adjacency{
		adjacencyNeighborAUp,
	}, s.GetAdjacencies()) {
		t.Errorf("Adjacency table mismatch")
		return
	}

	clock.Add(time.Second * 3)
	pkt = readNext(packet.L2_LS_PDU_TYPE, eth0)
	if !assert.Equal(t, lspLocalInitial, pkt) {
		return
	}

	// let's check if the LSP gets regenerated when it's life time goes down to 5 minutes:
	// We'll need to move the clock a few times and send a few hellos and then check for a fresh LSP packet with increased sequence number
	remainingLifetime := s.GetLSDB()[0].GetLSPDU().RemainingLifetime
	sequenceNumber := s.GetLSDB()[0].GetLSPDU().SequenceNumber
	for {
		clock.Add(time.Second * 4)
		eth0.SendFromRemote(neighborA.mac, helloFromNeighborAUp)

		eth0.DrainBuffer()
		lspdu := s.GetLSDB()[0].GetLSPDU()
		if lspdu.RemainingLifetime > remainingLifetime && lspdu.SequenceNumber > sequenceNumber {
			break
		}

		remainingLifetime = lspdu.RemainingLifetime
	}

	time.Sleep(time.Millisecond * 10)

	// lets provoke a timeout of the adjacency
	clock.Add(time.Second * 17)

	// checking if the adjancency is down
	if !assertAdjacencies([]*server.Adjacency{
		adjacencyNeighborADown,
	}, s.GetAdjacencies()) {
		t.Errorf("Adjacency table mismatch")
		return
	}

	eth0.DrainBuffer() // discards all packets that have been sent meanwhile that we didn't receive yet
	clock.Add(time.Second * 4)

	// check if hello does not contain a neighbor anymore
	pkt = readNext(packet.P2P_HELLO, eth0)
	if !assert.Equal(t, helloToNeighborADown, pkt) {
		return
	}

	// let's bring up the adjacency again and verfiy the adjacency goes down when the interface goes down
	eth0.SendFromRemote(neighborA.mac, helloFromNeighborAUp)

	clock.Add(time.Second * 4)
	time.Sleep(time.Millisecond * 10)
	if !assertAdjacencies([]*server.Adjacency{
		adjacencyNeighborAUp,
	}, s.GetAdjacencies()) {
		t.Errorf("Adjacency table mismatch")
		return
	}

	eth0.DrainBuffer()
	du.DeviceDownEvent("eth0", []*bnet.Prefix{
		bnet.NewPfx(bnet.IPv4FromOctets(169, 254, 100, 0), 31).Ptr(),
	})

	time.Sleep(time.Millisecond * 10)
	if !assertAdjacencies([]*server.Adjacency{
		adjacencyNeighborADown,
	}, s.GetAdjacencies()) {
		t.Errorf("Adjacency table mismatch")
		return
	}
}

func readNext(typ uint8, mi *ethernet.MockEthernetInterface) []byte {
	for {
		_, pkt := mi.ReceiveAtRemote()
		if pkt[4] == typ {
			return pkt
		}
	}
}

func assertAdjacencies(listA, listB []*server.Adjacency) bool {
	if len(listA) != len(listB) {
		return false
	}

	sort.Slice(listA, func(i, j int) bool {
		return listA[i].InterfaceName < listA[i].InterfaceName
	})
	sort.Slice(listB, func(i, j int) bool {
		return listB[i].InterfaceName < listB[i].InterfaceName
	})

	for i := range listA {
		if !adjacencyEqual(listA[i], listB[i]) {
			return false
		}
	}

	return true
}

func adjacencyEqual(a, b *server.Adjacency) bool {
	return a.Name == b.Name && a.SystemID == b.SystemID && a.Address == b.Address &&
		a.InterfaceName == b.InterfaceName && a.Level == b.Level && a.Priority == b.Priority &&
		a.Status == b.Status
}
