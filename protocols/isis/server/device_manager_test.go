package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/stretchr/testify/assert"
)

func TestRemoveDevice(t *testing.T) {
	tests := []struct {
		name           string
		dm             *devicesManager
		removeName     string
		wantFail       bool
		expected       *devicesManager
		wantUnregister bool
	}{
		{
			name: "Remove existing",
			dm: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: []*dev{
					{
						done: make(chan struct{}),
						sys:  &mockSys{},
						name: "foobar",
					},
				},
			},
			removeName: "foobar",
			expected: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{
						UnsubscribeCalled: true,
						UnsubscribeName:   "foobar",
					},
				},
				db: []*dev{},
			},
			wantUnregister: true,
		},
		{
			name: "Remove non-existing",
			dm: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: []*dev{
					{
						done: make(chan struct{}),
						sys:  &mockSys{},
						name: "foobar",
					},
				},
			},
			removeName: "baz",
			expected: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: []*dev{
					{
						done: make(chan struct{}),
						sys:  &mockSys{},
						name: "foobar",
					},
				},
			},
			wantUnregister: false,
			wantFail:       true,
		},
		{
			name: "Remove existing - disable fails",
			dm: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: []*dev{
					{
						done: make(chan struct{}),
						sys: &mockSys{
							wantFailClosedPacketSocket: true,
						},
						name: "foobar",
					},
				},
			},
			removeName: "foobar",
			expected: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{
						UnsubscribeCalled: true,
						UnsubscribeName:   "foobar",
					},
				},
				db: []*dev{},
			},
			wantUnregister: true,
			wantFail:       true,
		},
	}

	for _, test := range tests {
		err := test.dm.removeDevice(test.removeName)

		assert.Equal(t, test.wantUnregister, test.dm.srv.ds.(*device.MockServer).UnsubscribeCalled, test.name)

		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
		}

		// Ignore some attributes
		for i := range test.dm.db {
			test.dm.db[i].self = nil
			test.dm.db[i].srv = nil
			test.dm.db[i].done = nil
		}

		assert.Equal(t, test.expected, test.dm, test.name)
	}
}

func TestDeviceAddDevice(t *testing.T) {
	tests := []struct {
		name         string
		dm           *devicesManager
		addIfCfg     *config.ISISInterfaceConfig
		wantFail     bool
		expected     *devicesManager
		wantRegister bool
	}{
		{
			name: "Test #1",
			dm: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: []*dev{
					{
						name: "foobar",
					},
				},
			},
			addIfCfg: &config.ISISInterfaceConfig{
				Name:    "baz",
				Passive: true,
				ISISLevel2Config: &config.ISISLevelConfig{
					HelloInterval: 5,
				},
			},
			expected: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{
						Called: true,
						Name:   "baz",
					},
				},
				db: []*dev{
					{
						name: "foobar",
					},
					{
						name:    "baz",
						passive: true,
						level2: &level{
							HelloInterval: 5,
						},
					},
				},
			},
			wantRegister: true,
		},
		{
			name: "Test #2",
			dm: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{},
				},
				db: []*dev{
					{
						name: "foobar",
					},
				},
			},
			addIfCfg: &config.ISISInterfaceConfig{
				Name:    "foobar",
				Passive: true,
			},
			expected: &devicesManager{
				srv: &Server{
					ds: &device.MockServer{
						Called: true,
						Name:   "baz",
					},
				},
				db: []*dev{
					{
						name: "foobar",
					},
				},
			},
			wantRegister: false,
			wantFail:     true,
		},
	}

	for _, test := range tests {
		err := test.dm.addDevice(test.addIfCfg)

		assert.Equal(t, test.wantRegister, test.dm.srv.ds.(*device.MockServer).Called, test.name)

		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
		}

		// Ignore some attributes
		for i := range test.dm.db {
			test.dm.db[i].self = nil
			test.dm.db[i].srv = nil
			test.dm.db[i].done = nil
		}

		assert.Equal(t, test.expected, test.dm, test.name)
	}
}

func TestValidateNeighborAddresses(t *testing.T) {
	tests := []struct {
		name     string
		d        *dev
		addrs    []uint32
		expected []uint32
	}{
		{
			name: "Test #1",
			d: &dev{
				phy: &device.Device{
					Addrs: []net.Prefix{
						net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 24),
					},
				},
			},
			addrs: []uint32{
				net.IPv4FromOctets(10, 0, 0, 2).ToUint32(),
			},
			expected: []uint32{
				net.IPv4FromOctets(10, 0, 0, 2).ToUint32(),
			},
		},
		{
			name: "Test #2",
			d: &dev{
				phy: &device.Device{
					Addrs: []net.Prefix{
						net.NewPfx(net.IPv4FromOctets(10, 0, 0, 0), 30),
						net.NewPfx(net.IPv4FromOctets(10, 0, 0, 4), 30),
						net.NewPfx(net.IPv4FromOctets(192, 168, 100, 0), 22),
					},
				},
			},
			addrs: []uint32{
				net.IPv4FromOctets(100, 100, 100, 100).ToUint32(),
				net.IPv4FromOctets(10, 0, 0, 5).ToUint32(),
				net.IPv4FromOctets(10, 0, 0, 9).ToUint32(),
				net.IPv4FromOctets(192, 168, 101, 22).ToUint32(),
				net.IPv4FromOctets(10, 0, 0, 22).ToUint32(),
			},
			expected: []uint32{
				net.IPv4FromOctets(10, 0, 0, 5).ToUint32(),
				net.IPv4FromOctets(192, 168, 101, 22).ToUint32(),
			},
		},
	}

	for _, test := range tests {
		res := test.d.validateNeighborAddresses(test.addrs)
		assert.Equal(t, test.expected, res, test.name)
	}
}
