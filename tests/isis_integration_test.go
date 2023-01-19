package tests

import (
	"os"
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

func TestServer(t *testing.T) {
	clock := bbclock.NewMock()
	now, _ := time.Parse(testTimeLayout, "January 23, 2023 at 00:00:00.000")
	clock.Set(now)
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
	dst, pkt := eth0.ReceiveAtRemote()
	if !assert.Equal(t, []byte{
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
	}, pkt) {
		return
	}

	if !assert.Equal(t, ethernet.MACAddr{
		0x09, 0x00, 0x2b, 0x00, 0x00, 0x05,
	}, dst) {
		return
	}

	// send a hello from neighborA
	eth0.SendFromRemote(neighborA.mac, []byte{
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
		0, 0, 0, 0,
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
	})

	// checking if the adjancency exists
	for _, a := range s.GetAdjacencies() {
		assert.Equal(t, neighborA.mac.String(), a.Address.String())
		assert.Equal(t, "eth0", a.InterfaceName)
		assert.Equal(t, neighborA.systemID.String(), a.SystemID.String())
		assert.Equal(t, packet.P2PAdjStateInit, int(a.Status))
	}

	// let's see if the neighborA is not listed in the hello packet
	clock.Add(time.Second * 4)
	dst, pkt = eth0.ReceiveAtRemote()
	if !assert.Equal(t, []byte{
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
		1,          // Adj State down
		0, 0, 0, 0, // extended local circuit id
		222, 173, 190, 239, 255, 1, // Neighbor system ID
		0, 0, 0, 0, // Neighbor extended local circuit id
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
	}, pkt) {
		return
	}

	// lets send a hello from neighborA that contains the other router
	eth0.SendFromRemote(neighborA.mac, []byte{
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
		240,        // Type
		15,         // Length
		1,          // Adj State down
		0, 0, 0, 0, // extended local circuit id
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
	})

	clock.Add(time.Second)
	// checking if the adjancency is up
	for _, a := range s.GetAdjacencies() {
		assert.Equal(t, neighborA.mac.String(), a.Address.String())
		assert.Equal(t, "eth0", a.InterfaceName)
		assert.Equal(t, neighborA.systemID.String(), a.SystemID.String())
		assert.Equal(t, packet.P2PAdjStateUp, int(a.Status))
	}

	clock.Add(time.Second * 3)
	pkt = readNext(packet.L2_LS_PDU_TYPE, eth0)
	hostname, err := os.Hostname()
	if err != nil {
		panic("no hostname")
	}
	expected := []byte{
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
		0, 49, // Length
		7, 7, // Remaining Lifetime
		12, 12, 12, 13, 13, 13, 0, 0, // LSP ID
		0, 0, 0, 1, // Sequence number
		229, 53, // Checksum
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
		137, // Hostname
		6,   // Length
	}
	expected = append(expected, []byte(hostname)...)
	if !assert.Equal(t, expected, pkt) {
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
