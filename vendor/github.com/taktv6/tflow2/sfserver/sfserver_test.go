// Copyright 2017 EXARING AG. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package sfserver

import (
	"fmt"
	"net"
	"testing"
	"unsafe"

	"github.com/golang/glog"

	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/packet"
	"github.com/taktv6/tflow2/sflow"
)

func TestIntegration(t *testing.T) {
	s := []byte{
		10, 0, 0, 0, // Destination Mask
		32, 0, 0, 0, // Source Mask
		33, 250, 157, 62, // Next-Hop
		1, 0, 0, 0, // Address Family
		16, 0, 0, 0, // Flow Data Length
		234, 3, 0, 0, // Enterprise/Type (Extended router data)

		0, 0, 0, 0, // Priority OUT
		0, 0, 0, 0, // VLAN OUT
		0, 0, 0, 0, // Priority IN
		210, 0, 0, 0, // VLAN IN
		16, 0, 0, 0, // Flow Data Length
		233, 3, 0, 0, // Enterprise/Type (Extended switch data)

		209, 50, 196, 16, 191, 134, 236, 166, 206, 27, 249, 140, 64, 231, 148, 246, 19, 88, 36, 9, 167, 240, 97, 133, 46, 175, 100, 47, 143, 160, 84, 35, 234, 71, 176, 116, 103, 119, 151, 133, 184, 52, 169, 202, 53, 231, 149, 40, 16, 81, 31, 242, 100, 122, 152, 78, 32, 133, 116, 22, 89, 122, 149, 27, 64, 0, 173, 248, 203, 199, 10, 8, 1, 1, 0, 0, 199,
		212, 235, 0, 16,
		128,               // Header Length
		92, 180, 133, 203, // ACK Number
		31, 4, 191, 24, // Sequence Number
		222, 148, // DST port
		80, 0, // SRC port

		19, 131, 191, 87, // DST IP
		238, 153, 37, 185, // SRC IP
		186, 25, // Header Checksum
		6,     // Protocol
		62,    // TTL
		0, 64, // Flags + Fragment offset
		131, 239, // Identifier
		212, 5, // Total Length
		0,  // TOS
		69, // Version + Length

		0, 8, // EtherType
		0xb9, 0x1c, 0x04, 0x71, 0x4e, 0x20, // Source MAC
		0x94, 0x02, 0x7f, 0x1f, 0x71, 0x80, // Destination MAC

		128, 0, 0, 0, // Original Packet length (92 Bytes until here, incl.)
		4, 0, 0, 0, // Payload removed
		230, 5, 0, 0, // Frame length
		1, 0, 0, 0, // Header Protocol
		144, 0, 0, 0, // Flow Data Length
		1, 0, 0, 0, // Enterprise/Type (Raw packet header)

		3, 0, 0, 0, // Flow Record count
		146, 2, 0, 0, // Output interface
		7, 2, 0, 0, // Input interface
		0, 0, 0, 0, // Dropped Packets
		160, 81, 79, 192, // Sampling Pool
		224, 3, 0, 0, // Sampling Rate
		146, 2, 0, 0, // Source ID + Index
		210, 127, 173, 95, // Sequence Number
		232, 0, 0, 0, // sample length
		1, 0, 0, 0, // Enterprise/Type

		1, 0, 0, 0, // NumSamples
		111, 0, 0, 0, // SysUpTime
		222, 0, 0, 0, // Sequence Number
		0, 0, 0, 0, // Sub-AgentID
		14, 19, 205, 10, // Agent Address
		1, 0, 0, 0, // Agent Address Type
		5, 0, 0, 0, // Version
	}
	s = convert.Reverse(s)

	p, err := sflow.Decode(s, net.IP([]byte{1, 1, 1, 1}))
	if err != nil {
		t.Errorf("Decoding packet failed: %v\n", err)
	}

	for _, fs := range p.FlowSamples {
		if fs.RawPacketHeader == nil {
			glog.Infof("Received sflow packet without raw packet header. Skipped.")
			continue
		}

		ether, err := packet.DecodeEthernet(fs.RawPacketHeaderData, fs.RawPacketHeader.OriginalPacketLength)
		if err != nil {
			glog.Infof("Unable to decode ether packet: %v", err)
			continue
		}

		if ether.DstMAC.String() != "80:71:1f:7f:02:94" {
			t.Errorf("Unexpected DST MAC address. Expected %s. Got %s", "80:71:1f:7f:02:94", ether.DstMAC.String())
		}

		if ether.SrcMAC.String() != "20:4e:71:04:1c:b9" {
			t.Errorf("Unexpected SRC MAC address. Expected %s. Got %s", "20:4e:71:04:1c:b9", ether.SrcMAC.String())
		}

		if fs.RawPacketHeader.HeaderProtocol == 1 {
			ipv4Ptr := unsafe.Pointer(uintptr(fs.RawPacketHeaderData) - packet.SizeOfEthernetII)
			ipv4, err := packet.DecodeIPv4(ipv4Ptr, fs.RawPacketHeader.OriginalPacketLength-uint32(packet.SizeOfEthernetII))
			if err != nil {
				t.Errorf("Unable to decode IPv4 packet: %v", err)
			}

			convert.Reverse(ipv4.SrcAddr[:])
			if net.IP(ipv4.SrcAddr[:]).String() != "185.37.153.238" {
				t.Errorf("Wrong IPv4 src address: Got %v. Expected %v", net.IP(convert.Reverse(ipv4.SrcAddr[:])).String(), "185.37.153.238")
			}

			fmt.Printf("IPv4 SRC: %s\n", net.IP(ipv4.SrcAddr[:]).String())

			if ipv4.Protocol == 6 {
				tcpPtr := unsafe.Pointer(uintptr(ipv4Ptr) - packet.SizeOfIPv4Header)
				tcp, err := packet.DecodeTCP(tcpPtr, fs.RawPacketHeader.OriginalPacketLength-uint32(packet.SizeOfEthernetII)-uint32(packet.SizeOfIPv4Header))
				if err != nil {
					t.Errorf("Unable to decode TCP segment: %v", err)
				}
				fmt.Printf("SRC PORT: %d\n", tcp.SrcPort)
				fmt.Printf("DST PORT: %d\n", tcp.DstPort)
			} else {
				t.Errorf("Unknown IP protocol: %d\n", ipv4.Protocol)
			}

		} else {
			t.Errorf("Unknown HeaderProtocol: %d", fs.RawPacketHeader.HeaderProtocol)
		}
	}
}
