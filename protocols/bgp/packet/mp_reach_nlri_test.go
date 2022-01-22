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
		addPath  bool
	}{
		{
			name: "Simple IPv6 prefix",
			nlri: MultiProtocolReachNLRI{
				AFI:     IPv6AFI,
				SAFI:    UnicastSAFI,
				NextHop: bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0x2).Dedup(),
				NLRI: &NLRI{
					Prefix: bnet.NewPfx(bnet.IPv6FromBlocks(0x2600, 0x6, 0xff05, 0, 0, 0, 0, 0), 48).Dedup(),
				},
			},
			expected: []byte{
				0x00, 0x02, // AFI
				0x01,                                                                                                 // SAFI
				0x10, 0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // NextHop
				0x00,                                     // RESERVED
				0x30, 0x26, 0x00, 0x00, 0x06, 0xff, 0x05, // Prefix
			},
		},
		{
			name: "IPv6 prefix with ADD-PATH",
			nlri: MultiProtocolReachNLRI{
				AFI:     IPv6AFI,
				SAFI:    UnicastSAFI,
				NextHop: bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 0x2).Dedup(),
				NLRI: &NLRI{
					Prefix:         bnet.NewPfx(bnet.IPv6FromBlocks(0x2600, 0x6, 0xff05, 0, 0, 0, 0, 0), 48).Dedup(),
					PathIdentifier: 100,
				},
			},
			expected: []byte{
				0x00, 0x02, // AFI
				0x01,                                                                                                 // SAFI
				0x10, 0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // NextHop
				0x00,                  // RESERVED
				0x00, 0x00, 0x00, 100, // PathID
				0x30, 0x26, 0x00, 0x00, 0x06, 0xff, 0x05, // Prefix
			},
			addPath: true,
		},
		{
			name: "IPv4 BGP Labeled Unicast",
			nlri: MultiProtocolReachNLRI{
				AFI:     IPv4AFI,
				SAFI:    LabeledUnicastSAFI,
				NextHop: bnet.IPv4FromOctets(192, 0, 2, 0).Dedup(),
				NLRI: &NLRI{
					Prefix: bnet.NewPfx(bnet.IPv4FromOctets(192, 0, 2, 0), 24).Dedup(),
					LabelStack: []Label{
						299824,
					},
				},
			},
			expected: []byte{
				0x00, 0x01, // AFI
				0x04,         // SAFI
				0x04,         // NextHop length
				192, 0, 2, 0, // NextHop
				0x00,             // Reserved
				48,               // Prefix Length + Label Stack size (/24 + 24 bytes MPLS)
				0x49, 0x33, 0x01, // Label (bottom of stack)
				192, 0, 2, // Prefix
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			test.nlri.serialize(buf, &EncodeOptions{
				UseAddPath: test.addPath,
			})
			assert.Equal(t, test.expected, buf.Bytes())
		})
	}
}
