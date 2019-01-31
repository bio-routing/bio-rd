package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/stretchr/testify/assert"
)

func TestEnableDisable(t *testing.T) {
	tests := []struct {
		name     string
		dev      *dev
		wantFail bool
	}{
		{
			name: "Failed open() for socket",
			dev: &dev{
				sys: &mockSys{
					wantFailOpenPacketSocket: true,
				},
			},
			wantFail: true,
		},
		{
			name: "Failed mcast join",
			dev: &dev{
				sys: &mockSys{
					wantFailMcastJoin: true,
				},
			},
			wantFail: true,
		},
		{
			name: "Success",
			dev: &dev{
				sys: &mockSys{},
			},
			wantFail: false,
		},
	}

	for _, test := range tests {
		test.dev.self = newMockDev()
		err := test.dev.enable()
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

		err = test.dev.disable()
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

		assert.Equal(t, true, test.dev.sys.(*mockSys).closePacketSocketCalled)
	}
}

func TestDeviceUpdate(t *testing.T) {
	tests := []struct {
		name     string
		dev      *dev
		update   *device.Device
		expected bool
	}{
		{
			name: "Enable",
			dev: &dev{
				up:  false,
				sys: &mockSys{},
			},
			update: &device.Device{
				OperState: device.IfOperUp,
			},
			expected: true,
		},
		{
			name: "Disable #1",
			dev: &dev{
				done: make(chan struct{}),
				up:   true,
				sys:  &mockSys{},
			},
			update: &device.Device{
				OperState: device.IfOperLowerLayerDown,
			},
			expected: false,
		},
		{
			name: "Disable #2",
			dev: &dev{
				done: make(chan struct{}),
				up:   true,
				sys:  &mockSys{},
			},
			update: &device.Device{
				OperState: device.IfOperDown,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		test.dev.self = newMockDev()
		test.dev.DeviceUpdate(test.update)
		assert.Equal(t, test.expected, test.dev.up, test.name)
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
