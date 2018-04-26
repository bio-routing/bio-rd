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

// Package sfserver provides sflow collection services via UDP and passes flows into annotator layer
package sfserver

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/config"
	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/packet"
	"github.com/taktv6/tflow2/sflow"
	"github.com/taktv6/tflow2/srcache"
	"github.com/taktv6/tflow2/stats"
)

// SflowServer represents a sflow Collector instance
type SflowServer struct {
	// Output is the channel used to send flows to the annotator layer
	Output chan *netflow.Flow

	// debug defines the debug level
	debug int

	// bgpAugment is used to decide if ASN information from netflow packets should be used
	bgpAugment bool

	// con is the UDP socket
	conn *net.UDPConn

	wg sync.WaitGroup

	config *config.Config

	sampleRateCache *srcache.SamplerateCache
}

// New creates and starts a new `SflowServer` instance
func New(numReaders int, config *config.Config, sampleRateCache *srcache.SamplerateCache) *SflowServer {
	sfs := &SflowServer{
		Output:          make(chan *netflow.Flow),
		config:          config,
		sampleRateCache: sampleRateCache,
	}

	addr, err := net.ResolveUDPAddr("udp", *sfs.config.Sflow.Listen)
	if err != nil {
		panic(fmt.Sprintf("ResolveUDPAddr: %v", err))
	}

	con, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(fmt.Sprintf("Listen: %v", err))
	}

	// Create goroutines that read netflow packet and process it
	for i := 0; i < numReaders; i++ {
		sfs.wg.Add(numReaders)
		go func(num int) {
			sfs.packetWorker(num, con)
		}(i)
	}

	return sfs
}

// Close closes the socket and stops the workers
func (sfs *SflowServer) Close() {
	sfs.conn.Close()
	sfs.wg.Wait()
}

// packetWorker reads netflow packet from socket and handsoff processing to processFlowSets()
func (sfs *SflowServer) packetWorker(identity int, conn *net.UDPConn) {
	buffer := make([]byte, 8960)
	for {
		length, remote, err := conn.ReadFromUDP(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			glog.Errorf("Error reading from socket: %v", err)
			continue
		}
		atomic.AddUint64(&stats.GlobalStats.SflowPackets, 1)
		atomic.AddUint64(&stats.GlobalStats.SflowBytes, uint64(length))

		remote.IP = remote.IP.To4()
		if remote.IP == nil {
			glog.Errorf("Received IPv6 packet. Dropped.")
			continue
		}

		sfs.processPacket(remote.IP, buffer[:length])
	}
	sfs.wg.Done()
}

// processPacket takes a raw sflow packet, send it to the decoder and passes the decoded packet
func (sfs *SflowServer) processPacket(agent net.IP, buffer []byte) {
	length := len(buffer)
	p, err := sflow.Decode(buffer[:length], agent)
	if err != nil {
		glog.Errorf("sflow.Decode: %v", err)
		return
	}

	for _, fs := range p.FlowSamples {
		if fs.RawPacketHeader == nil {
			glog.Infof("Received sflow packet without raw packet header. Skipped.")
			continue
		}

		if fs.RawPacketHeaderData == nil {
			glog.Infof("Received sflow packet without raw packet header. Skipped.")
			continue
		}

		if fs.RawPacketHeader.HeaderProtocol != 1 {
			glog.Infof("Unknown header protocol: %d", fs.RawPacketHeader.HeaderProtocol)
			continue
		}

		ether, err := packet.DecodeEthernet(fs.RawPacketHeaderData, fs.RawPacketHeader.OriginalPacketLength)
		if err != nil {
			glog.Infof("Unable to decode ether packet: %v", err)
			continue
		}

		fl := &netflow.Flow{
			Router:     agent,
			IntIn:      fs.FlowSampleHeader.InputIf,
			IntOut:     fs.FlowSampleHeader.OutputIf,
			Size:       uint64(fs.RawPacketHeader.FlowDataLength),
			Packets:    uint32(1),
			Timestamp:  time.Now().Unix(),
			Samplerate: uint64(fs.FlowSampleHeader.SamplingRate),
		}

		// We're updating the sampleCache to allow the forntend to show current sampling rates
		sfs.sampleRateCache.Set(agent, uint64(fs.FlowSampleHeader.SamplingRate))

		if fs.ExtendedRouterData != nil {
			fl.NextHop = fs.ExtendedRouterData.NextHop
		}

		if ether.EtherType == packet.EtherTypeIPv4 {
			fl.Family = 4
			ipv4Ptr := unsafe.Pointer(uintptr(fs.RawPacketHeaderData) - packet.SizeOfEthernetII)
			ipv4, err := packet.DecodeIPv4(ipv4Ptr, fs.RawPacketHeader.OriginalPacketLength-uint32(packet.SizeOfEthernetII))
			if err != nil {
				glog.Errorf("Unable to decode IPv4 packet: %v", err)
			}

			fl.SrcAddr = convert.Reverse(ipv4.SrcAddr[:])
			fl.DstAddr = convert.Reverse(ipv4.DstAddr[:])
			fl.Protocol = uint32(ipv4.Protocol)
			switch ipv4.Protocol {
			case packet.TCP:
				tcpPtr := unsafe.Pointer(uintptr(ipv4Ptr) - packet.SizeOfIPv4Header)
				len := fs.RawPacketHeader.OriginalPacketLength - uint32(packet.SizeOfEthernetII) - uint32(packet.SizeOfIPv4Header)
				if err := getTCP(tcpPtr, len, fl); err != nil {
					glog.Errorf("%v", err)
				}
			case packet.UDP:
				udpPtr := unsafe.Pointer(uintptr(ipv4Ptr) - packet.SizeOfIPv4Header)
				len := fs.RawPacketHeader.OriginalPacketLength - uint32(packet.SizeOfEthernetII) - uint32(packet.SizeOfIPv4Header)
				if err := getUDP(udpPtr, len, fl); err != nil {
					glog.Errorf("%v", err)
				}
			}
		} else if ether.EtherType == packet.EtherTypeIPv6 {
			fl.Family = 6
			ipv6Ptr := unsafe.Pointer(uintptr(fs.RawPacketHeaderData) - packet.SizeOfEthernetII)
			ipv6, err := packet.DecodeIPv6(ipv6Ptr, fs.RawPacketHeader.OriginalPacketLength-uint32(packet.SizeOfEthernetII))
			if err != nil {
				glog.Errorf("Unable to decode IPv6 packet: %v", err)
			}

			fl.SrcAddr = convert.Reverse(ipv6.SrcAddr[:])
			fl.DstAddr = convert.Reverse(ipv6.DstAddr[:])
			fl.Protocol = uint32(ipv6.NextHeader)
			switch ipv6.NextHeader {
			case packet.TCP:
				tcpPtr := unsafe.Pointer(uintptr(ipv6Ptr) - packet.SizeOfIPv6Header)
				len := fs.RawPacketHeader.OriginalPacketLength - uint32(packet.SizeOfEthernetII) - uint32(packet.SizeOfIPv6Header)
				if err := getTCP(tcpPtr, len, fl); err != nil {
					glog.Errorf("%v", err)
				}
			case packet.UDP:
				udpPtr := unsafe.Pointer(uintptr(ipv6Ptr) - packet.SizeOfIPv6Header)
				len := fs.RawPacketHeader.OriginalPacketLength - uint32(packet.SizeOfEthernetII) - uint32(packet.SizeOfIPv6Header)
				if err := getUDP(udpPtr, len, fl); err != nil {
					glog.Errorf("%v", err)
				}
			}
		} else if ether.EtherType == packet.EtherTypeARP || ether.EtherType == packet.EtherTypeLACP {
			continue
		} else {
			glog.Errorf("Unknown EtherType: 0x%x", ether.EtherType)
		}

		sfs.Output <- fl
	}
}

func getUDP(udpPtr unsafe.Pointer, length uint32, fl *netflow.Flow) error {
	udp, err := packet.DecodeUDP(udpPtr, length)
	if err != nil {
		return fmt.Errorf("Unable to decode UDP datagram: %v", err)
	}

	fl.SrcPort = uint32(udp.SrcPort)
	fl.DstPort = uint32(udp.DstPort)

	return nil
}

func getTCP(tcpPtr unsafe.Pointer, length uint32, fl *netflow.Flow) error {
	tcp, err := packet.DecodeTCP(tcpPtr, length)
	if err != nil {
		return fmt.Errorf("Unable to decode TCP segment: %v", err)
	}

	fl.SrcPort = uint32(tcp.SrcPort)
	fl.DstPort = uint32(tcp.DstPort)

	return nil
}

// Dump dumps a flow on the screen
func Dump(fl *netflow.Flow) {
	fmt.Printf("--------------------------------\n")
	fmt.Printf("Flow dump:\n")
	fmt.Printf("Router: %d\n", fl.Router)
	fmt.Printf("Family: %d\n", fl.Family)
	fmt.Printf("SrcAddr: %s\n", net.IP(fl.SrcAddr).String())
	fmt.Printf("DstAddr: %s\n", net.IP(fl.DstAddr).String())
	fmt.Printf("Protocol: %d\n", fl.Protocol)
	fmt.Printf("NextHop: %s\n", net.IP(fl.NextHop).String())
	fmt.Printf("IntIn: %d\n", fl.IntIn)
	fmt.Printf("IntOut: %d\n", fl.IntOut)
	fmt.Printf("Packets: %d\n", fl.Packets)
	fmt.Printf("Bytes: %d\n", fl.Size)
	fmt.Printf("--------------------------------\n")
}
