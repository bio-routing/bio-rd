package server

import (
	"github.com/bio-routing/bio-rd/protocols/device"

	bnet "github.com/bio-routing/bio-rd/net"
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

func (m *mockDeviceUpdater) Start() error {
	return nil
}
