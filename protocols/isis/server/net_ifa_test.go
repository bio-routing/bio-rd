package server

import (
	"fmt"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"

	bnet "github.com/bio-routing/bio-rd/net"
	btesting "github.com/bio-routing/bio-rd/testing"
	btime "github.com/bio-routing/bio-rd/util/time"
)

type mockDevice struct {
	operState uint8
	addrs     []*bnet.Prefix
}

func (m *mockDevice) GetIndex() uint64 {
	return 1337
}

func (m *mockDevice) GetOperState() uint8 {
	return m.operState
}

func (m *mockDevice) GetAddrs() []*bnet.Prefix {
	return m.addrs
}

type mockDeviceUpdater struct {
	interfaces map[string]*mockDevice
}

func newMockDeviceUpdater() *mockDeviceUpdater {
	return &mockDeviceUpdater{
		interfaces: make(map[string]*mockDevice),
	}
}

func (m *mockDeviceUpdater) Subscribe(c device.Client, d string) {
	if _, found := m.interfaces[d]; !found {
		return
	}

	c.DeviceUpdate(m.interfaces[d])
}

func (m *mockDeviceUpdater) Unsubscribe(device.Client, string) {

}

func TestNetIfa(t *testing.T) {
	// In this test we have 2 interface: eth0 and eth1.
	// eth0 is initially up. eth1 is initally down.
	// first eth0 goes down. Then eth1 goes up.
	// After each status change we check if the netIfa behaves as expected.

	du := newMockDeviceUpdater()
	du.interfaces["eth0"] = &mockDevice{
		operState: device.IfOperUp,
	}
	du.interfaces["eth1"] = &mockDevice{
		operState: device.IfOperDown,
	}

	s := &Server{
		nets: []*types.NET{
			{
				AFI: 0x49,
				AreaID: types.AreaID{
					1,
				},
				SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
			},
		},
		ds: du,
	}
	s.netIfaManager = newNetIfaManager(s)
	s.netIfaManager.useMockTicker = true

	err := s.AddInterface(&InterfaceConfig{
		Name:         "eth0",
		Passive:      false,
		PointToPoint: true,
		HoldingTimer: 27,
		Level2: &InterfaceLevelConfig{
			Metric: 100,
		},
	})
	if err != nil {
		t.Errorf("Unexpected failure: %v", err)
		return
	}

	eth0 := s.netIfaManager.getInterface("eth0")
	eth0HelloTicket := eth0.helloTicker.(*btime.MockTicker)
	eth0BidiConn := btesting.NewMockConnBidi(&btesting.MockAddr{}, &btesting.MockAddr{})
	eth0.conn = eth0BidiConn

	// lets tick the hello ticker to trigger sending of first hello PDU
	eth0HelloTicket.Tick()
	fmt.Printf("Tick!\n")

	//time.Sleep(time.Second)
	buf := make([]byte, 1024)
	n, _ := eth0BidiConn.ReadA(buf)
	fmt.Printf("n: %d\n", n)
	fmt.Printf("Buf: %v\n", buf[:n])
	panic("BOOM")

	_ = eth0
}
