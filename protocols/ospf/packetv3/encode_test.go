package packetv3_test

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/net"
	ospf "github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type encodeTest struct {
	name     string
	msg      *ospf.OSPFv3Message
	expected []byte
}

func runEncodeTest(t *testing.T, testCase encodeTest, src, dst net.IP) {
	t.Run(testCase.name, func(t *testing.T) {
		out := new(bytes.Buffer)
		testCase.msg.Serialize(out, src, dst)
		assert.Equal(t, testCase.expected, out.Bytes())
	})
}

func TestEncodeHello(t *testing.T) {
	tests := []encodeTest{
		{
			name: "Init",
			msg: &ospf.OSPFv3Message{
				Version:    3,
				Type:       ospf.MsgTypeHello,
				RouterID:   routerID(5, 5, 5, 5),
				AreaID:     0,
				InstanceID: 0,
				Body: &ospf.Hello{
					InterfaceID:        4,
					RouterPriority:     200,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					HelloInterval:      20,
					RouterDeadInterval: 60,
				},
			},
			expected: []byte{
				0x03,       // Version
				0x01,       // Type: Hello
				0x00, 0x24, // Length
				0x05, 0x05, 0x05, 0x05, // Router ID
				0x00, 0x00, 0x00, 0x00, // Area ID
				0xd0, 0x7a, // Checksum
				0x00, // Instance ID
				0x00, // Reserved

				// Payload
				0x00, 0x00, 0x00, 0x04, // Interface ID
				0xc8,       // prio
				0x00,       // reserved
				0x00, 0x13, // Options
				0x00, 0x14, // Hello Interval
				0x00, 0x3c, // Dead Interval
				0x00, 0x00, 0x00, 0x00, // DR
				0x00, 0x00, 0x00, 0x00, // BDR
			},
		},
		{
			name: "WithDR",
			msg: &ospf.OSPFv3Message{
				Version:    3,
				Type:       ospf.MsgTypeHello,
				RouterID:   routerID(5, 5, 5, 5),
				AreaID:     0,
				InstanceID: 0,
				Body: &ospf.Hello{
					InterfaceID:        4,
					RouterPriority:     200,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					HelloInterval:      20,
					RouterDeadInterval: 60,
					DesignatedRouterID: routerID(3, 3, 3, 3),
					Neighbors: []ospf.ID{
						routerID(3, 3, 3, 3),
						routerID(6, 6, 6, 6),
					},
				},
			},
			expected: []byte{
				0x03,       // Version
				0x01,       // Type: Hello
				0x00, 0x2c, // Length
				0x05, 0x05, 0x05, 0x05, // Router ID
				0x00, 0x00, 0x00, 0x00, // Area ID
				0xb8, 0x52, // Checksum
				0x00, // Instance ID
				0x00, // Reserved

				// Payload
				0x00, 0x00, 0x00, 0x04, // Interface ID
				0xc8,       // prio
				0x00,       // reserved
				0x00, 0x13, // Options
				0x00, 0x14, // Hello Interval
				0x00, 0x3c, // Dead Interval
				0x03, 0x03, 0x03, 0x03, // DR
				0x00, 0x00, 0x00, 0x00, // BDR
				// Neighbors
				0x03, 0x03, 0x03, 0x03,
				0x06, 0x06, 0x06, 0x06,
			},
		},
		{
			name: "WithDRSelf",
			msg: &ospf.OSPFv3Message{
				Version:    3,
				Type:       ospf.MsgTypeHello,
				RouterID:   routerID(5, 5, 5, 5),
				AreaID:     0,
				InstanceID: 0,
				Body: &ospf.Hello{
					InterfaceID:              4,
					RouterPriority:           200,
					Options:                  ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					HelloInterval:            20,
					RouterDeadInterval:       60,
					DesignatedRouterID:       routerID(5, 5, 5, 5),
					BackupDesignatedRouterID: routerID(3, 3, 3, 3),
					Neighbors: []ospf.ID{
						routerID(3, 3, 3, 3),
					},
				},
			},
			expected: []byte{
				0x03,       // Version
				0x01,       // Type: Hello
				0x00, 0x28, // Length
				0x05, 0x05, 0x05, 0x05, // Router ID
				0x00, 0x00, 0x00, 0x00, // Area ID
				0xba, 0x5c, // Checksum
				0x00, // Instance ID
				0x00, // Reserved

				// Payload
				0x00, 0x00, 0x00, 0x04, // Interface ID
				0xc8,       // prio
				0x00,       // reserved
				0x00, 0x13, // Options
				0x00, 0x14, // Hello Interval
				0x00, 0x3c, // Dead Interval
				0x05, 0x05, 0x05, 0x05, // DR
				0x03, 0x03, 0x03, 0x03, // BDR
				// Neighbors
				0x03, 0x03, 0x03, 0x03,
			},
		},
	}

	src, err := net.IPFromString("fe80::c92:6b3f:92b0:d49e")
	require.NoError(t, err)
	dst, err := net.IPFromString("fe80::acac:d4ff:fe15:fd8b")
	require.NoError(t, err)

	for _, test := range tests {
		runEncodeTest(t, test, src, dst)
	}
}
