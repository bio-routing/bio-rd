package tests

import (
	"testing"

	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/server"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	du := &device.MockServer{}
	s, err := server.New([]*types.NET{
		{
			AFI:      0x49,
			AreaID:   types.AreaID{0x00},
			SystemID: types.SystemID{0xde, 0xad, 0xbe, 0xef, 0xff, 0xff},
			SEL:      0x00,
		},
	}, du, 3600)

	if err != nil {
		t.Errorf("unexpected failure creating IS-IS server: %v", err)
		return
	}

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

	du.DeviceUpEvent("eth0")
	eth0 := s.GetEthernetInterface("eth0").(*ethernet.MockEthernetInterface)
	dst, pkt := eth0.ReceiveAtRemote()
	if !assert.Equal(t, []byte{
		131, 20, 1, 0, 17, 1, 0, 0, 2, 222, 173, 190, 239, 255, 255, 0, 16, 0, 38, 1, 240, 5, 2, 0, 0, 0, 0, 129, 2, 204, 142, 132, 0, 1, 3, 2, 73, 0,
	}, pkt) {
		return
	}

	if !assert.Equal(t, ethernet.MACAddr{
		0x09, 0x00, 0x2b, 0x00, 0x00, 0x05,
	}, dst) {
		return
	}
}
