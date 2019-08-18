package net

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestBIONetIPFromAddr(t *testing.T) {
	tests := []struct {
		name     string
		hostPort string
		expected *bnet.IP
		wantFail bool
	}{
		{
			name:     "IPv6",
			hostPort: "[2001:678:1e0::1]:80",
			expected: bnet.IPv6FromBlocks(0x2001, 0x678, 0x01e0, 0, 0, 0, 0, 1),
		},
		{
			name:     "IPv4",
			hostPort: "192.168.1.1:80",
			expected: bnet.IPv4FromOctets(192, 168, 1, 1),
		},
		{
			name:     "hostname",
			hostPort: "myhost:80",
			wantFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ip, err := BIONetIPFromAddr(test.hostPort)
			if err != nil {
				if test.wantFail {
					return
				}

				t.Fatal(err)
			}

			assert.Equal(t, ip, test.expected)
		})
	}
}
