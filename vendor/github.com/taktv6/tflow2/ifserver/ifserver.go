// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ifserver provides IPFIX collection services via UDP and passes flows into annotator layer
package ifserver

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/config"
	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/ipfix"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/srcache"
	"github.com/taktv6/tflow2/stats"
)

// fieldMap describes what information is at what index in the slice
// that we get from decoding a netflow packet
type fieldMap struct {
	srcAddr                int
	dstAddr                int
	protocol               int
	packets                int
	size                   int
	intIn                  int
	intOut                 int
	nextHop                int
	family                 int
	vlan                   int
	ts                     int
	srcAsn                 int
	dstAsn                 int
	srcPort                int
	dstPort                int
	samplingPacketInterval int
}

// IPFIXServer represents a Netflow Collector instance
type IPFIXServer struct {
	// tmplCache is used to save received flow templates
	// for later lookup in order to decode netflow packets
	tmplCache *templateCache

	// receiver is the channel used to receive flows from the annotator layer
	Output chan *netflow.Flow

	// con is the UDP socket
	conn *net.UDPConn

	wg sync.WaitGroup

	sampleRateCache *srcache.SamplerateCache

	config *config.Config
}

// New creates and starts a new `IPFIXServer` instance
func New(numReaders int, config *config.Config, sampleRateCache *srcache.SamplerateCache) *IPFIXServer {
	ifs := &IPFIXServer{
		tmplCache:       newTemplateCache(),
		Output:          make(chan *netflow.Flow),
		sampleRateCache: sampleRateCache,
		config:          config,
	}

	addr, err := net.ResolveUDPAddr("udp", *ifs.config.IPFIX.Listen)
	if err != nil {
		panic(fmt.Sprintf("ResolveUDPAddr: %v", err))
	}

	con, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(fmt.Sprintf("Listen: %v", err))
	}

	// Create goroutines that read netflow packet and process it
	for i := 0; i < numReaders; i++ {
		ifs.wg.Add(numReaders)
		go func(num int) {
			ifs.packetWorker(num, con)
		}(i)
	}

	return ifs
}

// Close closes the socket and stops the workers
func (ifs *IPFIXServer) Close() {
	ifs.conn.Close()
	ifs.wg.Wait()
}

// validateSource checks if src is a configured agent
func (ifs *IPFIXServer) validateSource(src net.IP) bool {
	if _, ok := ifs.config.AgentsNameByIP[src.String()]; ok {
		return true
	}
	return false
}

// packetWorker reads netflow packet from socket and handsoff processing to processFlowSets()
func (ifs *IPFIXServer) packetWorker(identity int, conn *net.UDPConn) {
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
		atomic.AddUint64(&stats.GlobalStats.IPFIXpackets, 1)
		atomic.AddUint64(&stats.GlobalStats.IPFIXbytes, uint64(length))

		if !ifs.validateSource(remote.IP) {
			glog.Errorf("Unknown source: %s", remote.IP.String())
		}

		ifs.processPacket(remote.IP, buffer[:length])
	}
	ifs.wg.Done()
}

// processPacket takes a raw netflow packet, send it to the decoder, updates template cache
// (if there are templates in the packet) and passes the decoded packet over to processFlowSets()
func (ifs *IPFIXServer) processPacket(remote net.IP, buffer []byte) {
	length := len(buffer)
	packet, err := ipfix.Decode(buffer[:length], remote)
	if err != nil {
		glog.Errorf("ipfix.Decode: %v", err)
		return
	}

	ifs.updateTemplateCache(remote, packet)
	ifs.processFlowSets(remote, packet.Header.DomainID, packet.DataFlowSets(), int64(packet.Header.ExportTime), packet)
}

// processFlowSets iterates over flowSets and calls processFlowSet() for each flow set
func (ifs *IPFIXServer) processFlowSets(remote net.IP, domainID uint32, flowSets []*ipfix.Set, ts int64, packet *ipfix.Packet) {
	addr := remote.String()
	keyParts := make([]string, 3, 3)
	for _, set := range flowSets {
		template := ifs.tmplCache.get(convert.Uint32(remote), domainID, set.Header.SetID)

		if template == nil {
			templateKey := makeTemplateKey(addr, domainID, set.Header.SetID, keyParts)
			if *ifs.config.Debug > 0 {
				glog.Warningf("Template for given FlowSet not found: %s", templateKey)
			}
			continue
		}

		records := template.DecodeFlowSet(*set)
		if records == nil {
			glog.Warning("Error decoding FlowSet")
			continue
		}
		ifs.processFlowSet(template, records, remote, ts, packet)
	}
}

// process generates Flow elements from records and pushes them into the `receiver` channel
func (ifs *IPFIXServer) processFlowSet(template *ipfix.TemplateRecords, records []ipfix.FlowDataRecord, agent net.IP, ts int64, packet *ipfix.Packet) {
	fm := generateFieldMap(template)

	for _, r := range records {
		/*if template.OptionScopes != nil {
			if fm.samplingPacketInterval >= 0 {
				ifs.sampleRateCache.Set(agent, uint64(convert.Uint32(r.Values[fm.samplingPacketInterval])))
			}
			continue
		}*/

		if fm.family >= 0 {
			if fm.family == 4 {
				atomic.AddUint64(&stats.GlobalStats.Flows4, 1)
			} else if fm.family == 6 {
				atomic.AddUint64(&stats.GlobalStats.Flows6, 1)
			} else {
				glog.Warning("Unknown address family")
				continue
			}
		}

		var fl netflow.Flow
		fl.Router = agent
		fl.Timestamp = ts

		if fm.family >= 0 {
			fl.Family = uint32(fm.family)
		}

		if fm.packets >= 0 {
			fl.Packets = convert.Uint32(r.Values[fm.packets])
		}

		if fm.size >= 0 {
			fl.Size = uint64(convert.Uint32(r.Values[fm.size]))
		}

		if fm.protocol >= 0 {
			fl.Protocol = convert.Uint32(r.Values[fm.protocol])
		}

		if fm.intIn >= 0 {
			fl.IntIn = convert.Uint32(r.Values[fm.intIn])
		}

		if fm.intOut >= 0 {
			fl.IntOut = convert.Uint32(r.Values[fm.intOut])
		}

		if fm.srcPort >= 0 {
			fl.SrcPort = convert.Uint32(r.Values[fm.srcPort])
		}

		if fm.dstPort >= 0 {
			fl.DstPort = convert.Uint32(r.Values[fm.dstPort])
		}

		if fm.srcAddr >= 0 {
			fl.SrcAddr = convert.Reverse(r.Values[fm.srcAddr])
		}

		if fm.dstAddr >= 0 {
			fl.DstAddr = convert.Reverse(r.Values[fm.dstAddr])
		}

		if fm.nextHop >= 0 {
			fl.NextHop = convert.Reverse(r.Values[fm.nextHop])
		}

		if !*ifs.config.BGPAugmentation.Enabled {
			if fm.srcAsn >= 0 {
				fl.SrcAs = convert.Uint32(r.Values[fm.srcAsn])
			}

			if fm.dstAsn >= 0 {
				fl.DstAs = convert.Uint32(r.Values[fm.dstAsn])
			}
		}

		fl.Samplerate = ifs.sampleRateCache.Get(agent)

		if *ifs.config.Debug > 2 {
			Dump(&fl)
		}

		ifs.Output <- &fl
	}
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

// DumpTemplate dumps a template on the screen
func DumpTemplate(tmpl *ipfix.TemplateRecords) {
	fmt.Printf("Template %d\n", tmpl.Header.TemplateID)
	for rec, i := range tmpl.Records {
		fmt.Printf("%d: %v\n", i, rec)
	}
}

// generateFieldMap processes a TemplateRecord and populates a fieldMap accordingly
// the FieldMap can then be used to read fields from a flow
func generateFieldMap(template *ipfix.TemplateRecords) *fieldMap {
	fm := fieldMap{
		srcAddr:                -1,
		dstAddr:                -1,
		protocol:               -1,
		packets:                -1,
		size:                   -1,
		intIn:                  -1,
		intOut:                 -1,
		nextHop:                -1,
		family:                 -1,
		vlan:                   -1,
		ts:                     -1,
		srcAsn:                 -1,
		dstAsn:                 -1,
		srcPort:                -1,
		dstPort:                -1,
		samplingPacketInterval: -1,
	}

	i := -1
	for _, f := range template.Records {
		i++

		switch f.Type {
		case ipfix.IPv4SrcAddr:
			fm.srcAddr = i
			fm.family = 4
		case ipfix.IPv6SrcAddr:
			fm.srcAddr = i
			fm.family = 6
		case ipfix.IPv4DstAddr:
			fm.dstAddr = i
		case ipfix.IPv6DstAddr:
			fm.dstAddr = i
		case ipfix.InBytes:
			fm.size = i
		case ipfix.Protocol:
			fm.protocol = i
		case ipfix.InPkts:
			fm.packets = i
		case ipfix.InputSnmp:
			fm.intIn = i
		case ipfix.OutputSnmp:
			fm.intOut = i
		case ipfix.IPv4NextHop:
			fm.nextHop = i
		case ipfix.IPv6NextHop:
			fm.nextHop = i
		case ipfix.L4SrcPort:
			fm.srcPort = i
		case ipfix.L4DstPort:
			fm.dstPort = i
		case ipfix.SrcAs:
			fm.srcAsn = i
		case ipfix.DstAs:
			fm.dstAsn = i
		case ipfix.SamplingPacketInterval:
			fm.samplingPacketInterval = i
		}
	}

	return &fm
}

// updateTemplateCache updates the template cache
func (ifs *IPFIXServer) updateTemplateCache(remote net.IP, p *ipfix.Packet) {
	templRecs := p.GetTemplateRecords()
	for _, tr := range templRecs {
		ifs.tmplCache.set(convert.Uint32(remote), tr.Packet.Header.DomainID, tr.Header.TemplateID, *tr)
	}
}

// makeTemplateKey creates a string of the 3 tuple router address, source id and template id
func makeTemplateKey(addr string, sourceID uint32, templateID uint16, keyParts []string) string {
	keyParts[0] = addr
	keyParts[1] = strconv.Itoa(int(sourceID))
	keyParts[2] = strconv.Itoa(int(templateID))
	return strings.Join(keyParts, "|")
}
