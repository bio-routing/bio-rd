package packetv3_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	ospf "github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
	"github.com/bio-routing/bio-rd/protocols/ospf/packetv3/fixtures"
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
