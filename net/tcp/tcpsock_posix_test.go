package tcp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestNetTCPAddrToSockAddr(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected unix.Sockaddr
	}{
		{
			name: "IPv4 / 0.0.0.0",
			addr: "0.0.0.0:179",
			expected: &unix.SockaddrInet4{
				Port: 179,
				Addr: [4]byte{0, 0, 0, 0},
			},
		},
		{
			name: "IPv4 / 192.0.2.42",
			addr: "192.0.2.42:179",
			expected: &unix.SockaddrInet4{
				Port: 179,
				Addr: [4]byte{192, 0, 2, 42},
			},
		},
		{
			name: "IPv6 / 2001:db8::$2",
			addr: "[2001:db8::42]:179",
			expected: &unix.SockaddrInet6{
				Port: 179,
				Addr: [16]byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x42},
			},
		},
	}

	for _, test := range tests {
		tcpaddr, err := net.ResolveTCPAddr("tcp", test.addr)
		if err != nil {
			t.Fatalf("Failed to resolve TCPAddr: %v", err)
		}
		assert.Equal(t, test.expected, netTCPAddrToSockAddr(tcpaddr), test.name)
	}
}
