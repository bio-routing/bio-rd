package packet

import (
	"bytes"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestSerializeMultiProtocolUnreachNLRI(t *testing.T) {
	tests := []struct {
		name     string
		nlri     MultiProtocolUnreachNLRI
		expected []byte
		addPath  bool
	}{
		{
			name: "Simple IPv6 prefix",
			nlri: MultiProtocolUnreachNLRI{
				AFI:  IPv6AFI,
				SAFI: UnicastSAFI,
				NLRI: &NLRI{
					Prefix: bnet.NewPfx(bnet.IPv6FromBlocks(0x2620, 0x110, 0x9000, 0, 0, 0, 0, 0).Dedup(), 44).Dedup(),
				},
			},
			expected: []byte{
				0x00, 0x02, // AFI
				0x01,                                     // SAFI
				0x2c, 0x26, 0x20, 0x01, 0x10, 0x90, 0x00, // Prefix
			},
		},
		{
			name: "IPv6 prefix with ADD-PATH",
			nlri: MultiProtocolUnreachNLRI{
				AFI:  IPv6AFI,
				SAFI: UnicastSAFI,
				NLRI: &NLRI{
					PathIdentifier: 100,
					Prefix:         bnet.NewPfx(bnet.IPv6FromBlocks(0x2620, 0x110, 0x9000, 0, 0, 0, 0, 0).Dedup(), 44).Dedup(),
				},
			},
			expected: []byte{
				0x00, 0x02, // AFI
				0x01,                  // SAFI
				0x00, 0x00, 0x00, 100, // PathID
				0x2c, 0x26, 0x20, 0x01, 0x10, 0x90, 0x00, // Prefix
			},
			addPath: true,
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
