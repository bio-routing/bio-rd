package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	ospf "github.com/bio-routing/bio-rd/protocols/ospfv3/packet"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

var cwd string

func main() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd = dir

	handleFile("OSPFv3_multipoint_adjacencies.cap")
	handleFile("OSPFv3_broadcast_adjacency.cap")
	handleFile("OSPFv3_NBMA_adjacencies.cap")
}

func handleFile(path string) {
	fmt.Printf("Testing on file: %s\n", path)
	f, err := os.Open(cwd + "/examples/ospf/decode/" + path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r, err := pcapgo.NewReader(f)
	if err != nil {
		panic(err)
	}

	var successCount int
	var failedCount int
	for {
		data, _, err := r.ReadPacketData()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		payload := getPayload(data)
		success := handlePacket(payload)
		if success {
			successCount++
		} else {
			failedCount++
		}
	}

	fmt.Printf("Successfully decoded %d packets.\n", successCount)
	if failedCount > 0 {
		fmt.Printf("Failed on %d packets.\n", failedCount)
	}
}

func getPayload(raw []byte) []byte {
	packet := gopacket.NewPacket(raw, layers.LayerTypeEthernet, gopacket.Default)
	if err := packet.ErrorLayer(); err != nil {
		// fallback to handling of FrameRelay (cut-off header)
		packet = gopacket.NewPacket(raw[4:], layers.LayerTypeIPv6, gopacket.Default)
		if err := packet.ErrorLayer(); err != nil {
			panic(fmt.Errorf("Error decoding IPv6 layer of the packet: %v", err))
		}
	}

	return packet.NetworkLayer().LayerPayload()
}

func handlePacket(payload []byte) bool {
	buf := bytes.NewBuffer(payload)
	_, _, err := ospf.DeserializeOSPFv3Message(buf)
	if err != nil {
		fmt.Printf("Error decoding OSPF message: %+v", err)
		return false
	}

	return true
}
