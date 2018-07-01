package packet

import (
	"bytes"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestSerializeMultiProtocolReachNLRI(t *testing.T) {
	tests := []struct {
		name     string
		nlri     MultiProtocolReachNLRI
		expected []byte
	}{
		{
			name: "Simple IPv6 prefix",
			nlri: MultiProtocolReachNLRI{
				AFI:     IPv6AFI,
				SAFI:    UnicastSAFI,
				NextHop: bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0x2),
				Prefix:  bnet.NewPfx(bnet.IPv6FromBlocks(0x2600, 0x6, 0xff05, 0, 0, 0, 0, 0), 48),
			},
			expected: []byte{
				0x00, 0x02, // AFI
				0x01,                                                                                                 // SAFI
				0x10, 0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // NextHop
				0x00,                                     // RESERVED
				0x30, 0x26, 0x00, 0x00, 0x06, 0xff, 0x05, // Prefix
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			test.nlri.serialize(buf)
			assert.Equal(t, test.expected, buf.Bytes())
		})
	}
}
