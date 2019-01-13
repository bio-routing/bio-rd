package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/stretchr/testify/assert"
)

func (d *dev) mockRecv() {
	<-d.done
	d.wg.Done()
}

func (d *dev) mockHello() {
	<-d.done
	d.wg.Done()
}

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
		test.dev.receiverMethod = test.dev.mockRecv
		test.dev.helloMethod = test.dev.mockHello

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
		{
			name: "Enable failed",
			dev: &dev{
				up: false,
				sys: &mockSys{
					wantFailOpenPacketSocket: true,
				},
			},
			update: &device.Device{
				OperState: device.IfOperUp,
			},
			expected: false,
			wantFail: true,
		},
		{
			name: "Disable failed",
			dev: &dev{
				done: make(chan struct{}),
				up:   true,
				sys: &mockSys{
					wantFailClosedPacketSocket: true,
				},
			},
			update: &device.Device{
				OperState: device.IfOperLowerLayerDown,
			},
			expected: true,
			wantFail: true,
		},
	}

	for _, test := range tests {
		test.dev.receiverMethod = test.dev.mockRecv
		test.dev.helloMethod = test.dev.mockHello

		test.dev.DeviceUpdate(test.update)

		assert.Equal(t, test.expected, test.dev.up, test.name)
	}
}
