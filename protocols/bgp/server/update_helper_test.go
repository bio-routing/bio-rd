package server

import (
	"testing"

	"bytes"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"

	"github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"
)

func TestSerializeAndSendUpdate(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		testUpdate serializeAbleUpdate
		expected   []byte
	}{
		{
			name: "normal bgp update",
			err:  nil,
			testUpdate: &packet.BGPUpdate{
				WithdrawnRoutesLen: 5,
				WithdrawnRoutes: &packet.NLRI{
					IP:     strAddr("10.0.0.0"),
					Pfxlen: 8,
					Next: &packet.NLRI{
						IP:     strAddr("192.168.0.0"),
						Pfxlen: 16,
					},
				},
			},

			expected: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
				0, 28, // Length
				2,                               // Type = Update
				0, 5, 8, 10, 16, 192, 168, 0, 0, // 2 withdraws
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := serializeAndSendUpdate(buf, test.testUpdate)
			assert.Equal(t, test.err, err)

			assert.Equal(t, test.expected, buf.Bytes())
		})

	}
}

func strAddr(s string) uint32 {
	ret, _ := net.StrToAddr(s)
	return ret
}
