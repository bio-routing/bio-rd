package device

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCopy(t *testing.T) {
	tests := []struct {
		name     string
		dev      *Device
		expected *Device
	}{
		{
			name: "Test #1",
			dev: &Device{
				Name: "Foo",
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
				},
			},
			expected: &Device{
				Name: "Foo",
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
				},
			},
		},
	}

	for _, test := range tests {
		copy := test.dev.copy()
		test.dev.addAddr(bnet.NewPfx(bnet.IPv4(200), 8))
		assert.Equalf(t, test.expected, copy, "Test %q", test.name)
	}
}

func TestDeviceDelAddr(t *testing.T) {
	tests := []struct {
		name     string
		dev      *Device
		delete   *bnet.Prefix
		expected *Device
	}{
		{
			name: "Test #1",
			dev: &Device{
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
					bnet.NewPfx(bnet.IPv4(200), 8),
					bnet.NewPfx(bnet.IPv4(300), 8),
				},
			},
			delete: bnet.NewPfx(bnet.IPv4(200), 8),
			expected: &Device{
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
					bnet.NewPfx(bnet.IPv4(300), 8),
				},
			},
		},
		{
			name: "Test #2",
			dev: &Device{
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
					bnet.NewPfx(bnet.IPv4(200), 8),
					bnet.NewPfx(bnet.IPv4(300), 8),
				},
			},
			delete: bnet.NewPfx(bnet.IPv4(100), 8),
			expected: &Device{
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(200), 8),
					bnet.NewPfx(bnet.IPv4(300), 8),
				},
			},
		},
	}

	for _, test := range tests {
		test.dev.delAddr(test.delete)
		assert.Equalf(t, test.expected, test.dev, "Test %q", test.name)
	}
}

func TestDeviceAddAddr(t *testing.T) {
	tests := []struct {
		name     string
		dev      *Device
		input    *bnet.Prefix
		expected *Device
	}{
		{
			name: "Test #1",
			dev: &Device{
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
				},
			},
			input: bnet.NewPfx(bnet.IPv4(200), 8),
			expected: &Device{
				Addrs: []*bnet.Prefix{
					bnet.NewPfx(bnet.IPv4(100), 8),
					bnet.NewPfx(bnet.IPv4(200), 8),
				},
			},
		},
	}

	for _, test := range tests {
		test.dev.addAddr(test.input)
		assert.Equalf(t, test.expected, test.dev, "Test %q", test.name)
	}
}
