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
				RouterID:     ospf.ID(net.IPv4FromOctets(3, 3, 3, 3).Ptr().ToUint32()),
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
				RouterID:     ospf.ID(net.IPv4FromOctets(3, 3, 3, 3).Ptr().ToUint32()),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:        6,
					RouterPriority:     100,
					HelloInterval:      30,
					RouterDeadInterval: 120,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					Neighbors: []ospf.ID{
						ospf.ID(net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32()),
						ospf.ID(net.IPv4FromOctets(2, 2, 2, 2).Ptr().ToUint32()),
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
				RouterID:     ospf.ID(net.IPv4FromOctets(3, 3, 3, 3).Ptr().ToUint32()),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:        6,
					RouterPriority:     100,
					HelloInterval:      30,
					RouterDeadInterval: 120,
					Options:            ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					DesignatedRouterID: ospf.ID(net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32()),
					Neighbors: []ospf.ID{
						ospf.ID(net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32()),
						ospf.ID(net.IPv4FromOctets(2, 2, 2, 2).Ptr().ToUint32()),
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
				RouterID:     ospf.ID(net.IPv4FromOctets(3, 3, 3, 3).Ptr().ToUint32()),
				AreaID:       0,
				InstanceID:   0,
				Body: &ospf.Hello{
					InterfaceID:              6,
					RouterPriority:           100,
					HelloInterval:            30,
					RouterDeadInterval:       120,
					Options:                  ospf.OptionsFromFlags(ospf.RouterOptR, ospf.RouterOptE, ospf.RouterOptV6),
					DesignatedRouterID:       ospf.ID(net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32()),
					BackupDesignatedRouterID: ospf.ID(net.IPv4FromOctets(2, 2, 2, 2).Ptr().ToUint32()),
					Neighbors: []ospf.ID{
						ospf.ID(net.IPv4FromOctets(1, 1, 1, 1).Ptr().ToUint32()),
						ospf.ID(net.IPv4FromOctets(2, 2, 2, 2).Ptr().ToUint32()),
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
