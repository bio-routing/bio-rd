package packetv3_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/bio-routing/bio-rd/net"
	ospf "github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
	"github.com/bio-routing/bio-rd/protocols/ospf/packetv3/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var files = []string{
	"OSPFv3_multipoint_adjacencies.cap",
	"OSPFv3_broadcast_adjacency.cap",
	"OSPFv3_NBMA_adjacencies.cap",
}

var dir string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir = cwd + "/fixtures/"
}

func TestDecodeDumps(t *testing.T) {
	for _, path := range files {
		t.Run(path, func(t *testing.T) {
			testDecodeFile(t, dir+path)
		})
	}
}

func testDecodeFile(t *testing.T, path string) {
	fmt.Printf("Testing on file: %s\n", path)
	r, f := fixtures.PacketReader(t, path)
	defer f.Close()

	var packetCount int
	for {
		data, _, err := r.ReadPacketData()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			return
		}

		t.Run(fmt.Sprintf("Packet_%03d", packetCount+1), func(t *testing.T) {
			payload, src, dst, err := fixtures.Payload(data)
			if err != nil {
				t.Error(err)
				return
			}

			buf := bytes.NewBuffer(payload)
			if _, _, err := ospf.DeserializeOSPFv3Message(buf, src, dst); err != nil {
				t.Error(err)
			}
		})
		packetCount++
	}
}

type test struct {
	name     string
	input    []byte
	wantFail bool
	expected interface{}
}

func runTest(t *testing.T, testCase test, src, dst net.IP) {
	t.Run(testCase.name, func(t *testing.T) {
		buf := bytes.NewBuffer(testCase.input)
		msg, _, err := ospf.DeserializeOSPFv3Message(buf, src, dst)
		if testCase.wantFail {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, testCase.expected, msg)
	})
}

func routerID(o1, o2, o3, o4 uint8) ospf.ID {
	return ospf.ID(net.IPv4FromOctets(o1, o2, o3, o4).Ptr().ToUint32())
}

func TestDecodeHello(t *testing.T) {
	tests := []test{
		{
			name: "Default",
			input: []byte{
				// Header
				3,     // Version
				1,     // Type: Hello
				0, 36, // Length
				3, 3, 3, 3, // Router ID
				0, 0, 0, 0, // Area ID
				0x94, 0x1c, // Checksum
				0, // Instance ID
				0, // Reserved

				// Payload (Hello)
				0, 0, 0, 6, // Interface ID
				100,     // Router Prio
				0,       // Reserved
				0, 0x13, // Options: R, E, V6
				0, 30, // Hello Interval
				0, 120, // Dead Interval
				0, 0, 0, 0, // Designated Router
				0, 0, 0, 0, // Backup Designated Router
			},
			expected: &ospf.OSPFv3Message{
				Version:      3,
				Type:         ospf.MsgTypeHello,
				Checksum:     0x941c,
				PacketLength: 36,
				RouterID:     routerID(3, 3, 3, 3),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:        6,
					RouterPriority:     100,
					HelloInterval:      30,
					RouterDeadInterval: 120,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
				},
			},
		},
		{
			name: "InvalidLength",
			input: []byte{
				// Header
				3,     // Version
				1,     // Type: Hello
				0, 38, // Length (invalid, expecting 36)
				3, 3, 3, 3, // Router ID
				0, 0, 0, 0, // Area ID
				0x94, 0x1a, // Checksum
				0, // Instance ID
				0, // Reserved

				// Payload (20 bytes)
				0, 0, 0, 6, 100, 0, 0, 0x13, 0, 30, 0, 120, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			wantFail: true,
		},
		{
			name: "InvalidChecksum",
			input: []byte{
				// Header
				3,     // Version
				1,     // Type: Hello
				0, 36, // Length
				3, 3, 3, 3, // Router ID
				0, 0, 0, 0, // Area ID
				0x94, 0x1d, // Checksum (invalid)
				0, // Instance ID
				0, // Reserved

				// Payload (20 bytes)
				0, 0, 0, 6, 100, 0, 0, 0x13, 0, 30, 0, 120, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			wantFail: true,
		},
		{
			name: "WithNeighbors",
			input: []byte{
				// Header
				3,     // Version
				1,     // Type: Hello
				0, 44, // Length
				3, 3, 3, 3, // Router ID
				0, 0, 0, 0, // Area ID
				0x8e, 0x06, // Checksum
				0, // Instance ID
				0, // Reserved

				// Payload (Hello)
				0, 0, 0, 6, // Interface ID
				100,     // Router Prio
				0,       // Reserved
				0, 0x13, // Options: R, E, V6
				0, 30, // Hello Interval
				0, 120, // Dead Interval
				0, 0, 0, 0, // Designated Router
				0, 0, 0, 0, // Backup Designated Router
				1, 1, 1, 1, // Neighbor 1
				2, 2, 2, 2, // Neighbor 2
			},
			expected: &ospf.OSPFv3Message{
				Version:      3,
				Type:         ospf.MsgTypeHello,
				Checksum:     0x8e06,
				PacketLength: 44,
				RouterID:     routerID(3, 3, 3, 3),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:        6,
					RouterPriority:     100,
					HelloInterval:      30,
					RouterDeadInterval: 120,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					Neighbors: []ospf.ID{
						routerID(1, 1, 1, 1),
						routerID(2, 2, 2, 2),
					},
				},
			},
		},
		{
			name: "WithDR",
			input: []byte{
				// Header
				3,     // Version
				1,     // Type: Hello
				0, 44, // Length
				3, 3, 3, 3, // Router ID
				0, 0, 0, 0, // Area ID
				0x8c, 0x04, // Checksum
				0, // Instance ID
				0, // Reserved

				// Payload (Hello)
				0, 0, 0, 6, // Interface ID
				100,     // Router Prio
				0,       // Reserved
				0, 0x13, // Options: R, E, V6
				0, 30, // Hello Interval
				0, 120, // Dead Interval
				1, 1, 1, 1, // Designated Router
				0, 0, 0, 0, // Backup Designated Router
				1, 1, 1, 1, // Neighbor 1
				2, 2, 2, 2, // Neighbor 2
			},
			expected: &ospf.OSPFv3Message{
				Version:      3,
				Type:         ospf.MsgTypeHello,
				Checksum:     0x8c04,
				PacketLength: 44,
				RouterID:     routerID(3, 3, 3, 3),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:        6,
					RouterPriority:     100,
					HelloInterval:      30,
					RouterDeadInterval: 120,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					DesignatedRouterID: routerID(1, 1, 1, 1),
					Neighbors: []ospf.ID{
						routerID(1, 1, 1, 1),
						routerID(2, 2, 2, 2),
					},
				},
			},
		},
		{
			name: "WithBDR",
			input: []byte{
				// Header
				3,     // Version
				1,     // Type: Hello
				0, 44, // Length
				3, 3, 3, 3, // Router ID
				0, 0, 0, 0, // Area ID
				0x88, 0x00, // Checksum
				0, // Instance ID
				0, // Reserved

				// Payload (Hello)
				0, 0, 0, 6, // Interface ID
				100,     // Router Prio
				0,       // Reserved
				0, 0x13, // Options: R, E, V6
				0, 30, // Hello Interval
				0, 120, // Dead Interval
				1, 1, 1, 1, // Designated Router
				2, 2, 2, 2, // Backup Designated Router
				1, 1, 1, 1, // Neighbor 1
				2, 2, 2, 2, // Neighbor 2
			},
			expected: &ospf.OSPFv3Message{
				Version:      3,
				Type:         ospf.MsgTypeHello,
				Checksum:     0x8800,
				PacketLength: 44,
				RouterID:     routerID(3, 3, 3, 3),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:              6,
					RouterPriority:           100,
					HelloInterval:            30,
					RouterDeadInterval:       120,
					Options:                  ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					DesignatedRouterID:       routerID(1, 1, 1, 1),
					BackupDesignatedRouterID: routerID(2, 2, 2, 2),
					Neighbors: []ospf.ID{
						routerID(1, 1, 1, 1),
						routerID(2, 2, 2, 2),
					},
				},
			},
		},
	}

	src, err := net.IPFromString("fe80::3")
	require.NoError(t, err)
	dst, err := net.IPFromString("ff02::5")
	require.NoError(t, err)

	for _, test := range tests {
		runTest(t, test, src, dst)
	}
}

func TestDecodeDBDesc(t *testing.T) {
	tests := []test{
		{
			name: "Default",
			input: []byte{
				// Header
				0x03,       // Version
				0x02,       // Type: Database Description
				0x00, 0x1c, // Length
				0x03, 0x03, 0x03, 0x03, // Router ID
				0x00, 0x00, 0x00, 0x00, // Area ID
				0xe7, 0xad, // Checksum
				0x00, // Instance ID
				0x00, // Reserved

				// Payload
				0x00,             // Reserved
				0x00, 0x00, 0x13, // Options
				0x05, 0xdc, // MTU
				0x00,                   // Reserved
				0x07,                   // Description Flags
				0x00, 0x00, 0x0b, 0xbd, // Sequence Number
			},
			expected: &ospf.OSPFv3Message{
				Version:      3,
				Type:         ospf.MsgTypeDatabaseDescription,
				Checksum:     0xe7ad,
				PacketLength: 28,
				RouterID:     routerID(3, 3, 3, 3),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.DatabaseDescription{
					Options:          ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					InterfaceMTU:     1500,
					DBFlags:          ospf.DBFlagInit | ospf.DBFlagMore | ospf.DBFlagMS,
					DDSequenceNumber: 3005,
				},
			},
		},
		{
			name: "WithLSAs",
			input: []byte{
				// Header
				0x03,       // Version
				0x02,       // Type: Database Description
				0x00, 0xbc, // Length
				0x01, 0x01, 0x01, 0x01, // Router ID
				0x00, 0x00, 0x00, 0x00, // Area ID
				0xb6, 0xd0, // Checksum
				0x00, // Instance ID
				0x00, // Reserved

				// Payload
				0x00,             // Reserved
				0x00, 0x00, 0x13, // Options
				0x05, 0xdc, // Link MTU
				0x00,                   // Reserved
				0x02,                   // Flags
				0x00, 0x00, 0x0b, 0xbd, // Seq Num

				// LSA Header
				0x00, 0x1d,
				0x20, 0x01, // Type: Router-LSA
				0x00, 0x00, 0x00, 0x00, // LS ID
				0x01, 0x01, 0x01, 0x01, // Router ID
				0x80, 0x00, 0x00, 0x12, // Seq Num
				0xb1, 0x4a, // Checksum
				0x00, 0x18, // Length

				// LSA Header
				0x01, 0xb4,
				0x20, 0x01, // Type: Router-LSA
				0x00, 0x00, 0x00, 0x00,
				0x02, 0x02, 0x02, 0x02,
				0x80, 0x00, 0x00, 0x0f,
				0x02, 0x8e,
				0x00, 0x28,

				// LSA Header: Network-LSA
				0x01, 0xdc, 0x20, 0x02, 0x00, 0x00, 0x00, 0x06,
				0x03, 0x03, 0x03, 0x03, 0x80, 0x00, 0x00, 0x02,
				0x6d, 0x6c, 0x00, 0x24,

				// LSA-Header: Inter-Area-Prefix-LSA
				0x00, 0x1e, 0x20, 0x03, 0x00, 0x00, 0x00, 0x05,
				0x01, 0x01, 0x01, 0x01, 0x80, 0x00, 0x00, 0x01,
				0xdb, 0x0f, 0x00, 0x24,

				// LSA-Header: Inter-Area-Prefix-LSA
				0x03, 0x2a, 0x20, 0x03, 0x00, 0x00, 0x00, 0x04,
				0x02, 0x02, 0x02, 0x02, 0x80, 0x00, 0x00, 0x01,
				0xc7, 0x20, 0x00, 0x24,

				// LSA-Header: Link-LSA
				0x00, 0x1d, 0x00, 0x08, 0x00, 0x00, 0x00, 0x06,
				0x01, 0x01, 0x01, 0x01, 0x80, 0x00, 0x00, 0x01,
				0x86, 0xd0, 0x00, 0x38,

				// LSA-Header: Intra-Area-Prefix-LSA
				0x00, 0x1d, 0x20, 0x09, 0x00, 0x00, 0x00, 0x00,
				0x01, 0x01, 0x01, 0x01, 0x80, 0x00, 0x00, 0x01,
				0x74, 0x18, 0x00, 0x34,

				// LSA-Header: Unknown type
				0x00, 0x1d,
				0x20, 0x22, // Type: Unknown
				0x00, 0x00, 0x00, 0x00,
				0x01, 0x01, 0x01, 0x01,
				0x80, 0x00, 0x00, 0x01,
				0x74, 0x18, 0x00, 0x34,
			},
			expected: &ospf.OSPFv3Message{
				Version:      3,
				Type:         ospf.MsgTypeDatabaseDescription,
				Checksum:     0xb6d0,
				PacketLength: 188,
				RouterID:     routerID(1, 1, 1, 1),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.DatabaseDescription{
					Options:          ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					InterfaceMTU:     1500,
					DBFlags:          ospf.DBFlagMore,
					DDSequenceNumber: 3005,
					LSAHeaders: []*ospf.LSA{
						{
							Type:              ospf.LSATypeRouter,
							Age:               29,
							ID:                0,
							AdvertisingRouter: routerID(1, 1, 1, 1),
							SequenceNumber:    0x80000012,
							Checksum:          0xb14a,
							Length:            0x18,
						},
						{
							Type:              ospf.LSATypeRouter,
							Age:               436,
							ID:                0,
							AdvertisingRouter: routerID(2, 2, 2, 2),
							SequenceNumber:    0x8000000f,
							Checksum:          0x028e,
							Length:            0x28,
						},
						{
							Type:              ospf.LSATypeNetwork,
							Age:               476,
							ID:                6,
							AdvertisingRouter: routerID(3, 3, 3, 3),
							SequenceNumber:    0x80000002,
							Checksum:          0x6d6c,
							Length:            0x24,
						},
						{
							Type:              ospf.LSATypeInterAreaPrefix,
							Age:               30,
							ID:                5,
							AdvertisingRouter: routerID(1, 1, 1, 1),
							SequenceNumber:    0x80000001,
							Checksum:          0xdb0f,
							Length:            0x24,
						},
						{
							Type:              ospf.LSATypeInterAreaPrefix,
							Age:               0x032a,
							ID:                4,
							AdvertisingRouter: routerID(2, 2, 2, 2),
							SequenceNumber:    0x80000001,
							Checksum:          0xc720,
							Length:            0x24,
						},
						{
							Type:              ospf.LSATypeLink,
							Age:               0x001d,
							ID:                6,
							AdvertisingRouter: routerID(1, 1, 1, 1),
							SequenceNumber:    0x80000001,
							Checksum:          0x86d0,
							Length:            0x38,
						},
						{
							Type:              ospf.LSATypeIntraAreaPrefix,
							Age:               0x001d,
							ID:                0,
							AdvertisingRouter: routerID(1, 1, 1, 1),
							SequenceNumber:    0x80000001,
							Checksum:          0x7418,
							Length:            0x34,
						},
						{
							Type:              0x2022, // Unknown
							Age:               0x001d,
							ID:                0,
							AdvertisingRouter: routerID(1, 1, 1, 1),
							SequenceNumber:    0x80000001,
							Checksum:          0x7418,
							Length:            0x34,
						},
					},
				},
			},
		},
	}

	src, err := net.IPFromString("fe80::3")
	require.NoError(t, err)
	dst, err := net.IPFromString("fe80::1")
	require.NoError(t, err)

	for _, test := range tests {
		runTest(t, test, src, dst)
	}
}
