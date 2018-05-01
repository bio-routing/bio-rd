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

// Package nfserver provides netflow collection services via UDP and passes flows into annotator layer
package nfserver

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/taktv6/tflow2/config"
	"github.com/taktv6/tflow2/srcache"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/nf9"
	"github.com/taktv6/tflow2/stats"
)

// fieldMap describes what information is at what index in the slice
// that we get from decoding a netflow packet
type fieldMap struct {
	srcAddr                   int
	dstAddr                   int
	protocol                  int
	packets                   int
	size                      int
	intIn                     int
	intOut                    int
	nextHop                   int
	family                    int
	vlan                      int
	ts                        int
	srcAsn                    int
	dstAsn                    int
	srcPort                   int
	dstPort                   int
	flowSamplerID             int
	samplingInterval          int
	flowSamplerRandomInterval int
}

// NetflowServer represents a Netflow Collector instance
type NetflowServer struct {
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

// New creates and starts a new `NetflowServer` instance
func New(numReaders int, config *config.Config, sampleRateCache *srcache.SamplerateCache) *NetflowServer {
	nfs := &NetflowServer{
		tmplCache:       newTemplateCache(),
		Output:          make(chan *netflow.Flow),
		sampleRateCache: sampleRateCache,
		config:          config,
	}

	addr, err := net.ResolveUDPAddr("udp", *nfs.config.NetflowV9.Listen)
	if err != nil {
		panic(fmt.Sprintf("ResolveUDPAddr: %v", err))
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(fmt.Sprintf("Listen: %v", err))
	}
	nfs.conn = conn

	// Create goroutines that read netflow packet and process it
	nfs.wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func(num int) {
			nfs.packetWorker(num)
		}(i)
	}

	return nfs
}

// Close closes the socket and stops the workers
func (nfs *NetflowServer) Close() {
	nfs.conn.Close()
	nfs.wg.Wait()
}

// validateSource checks if src is a configured agent
func (nfs *NetflowServer) validateSource(src net.IP) bool {
	if _, ok := nfs.config.AgentsNameByIP[src.String()]; ok {
		return true
	}
	return false
}

// packetWorker reads netflow packet from socket and handsoff processing to processFlowSets()
func (nfs *NetflowServer) packetWorker(identity int) {
	buffer := make([]byte, 8960)
	for {
		length, remote, err := nfs.conn.ReadFromUDP(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			glog.Errorf("Error reading from socket: %v", err)
			continue
		}
		atomic.AddUint64(&stats.GlobalStats.Netflow9packets, 1)
		atomic.AddUint64(&stats.GlobalStats.Netflow9bytes, uint64(length))

		if !nfs.validateSource(remote.IP) {
			glog.Errorf("Unknown source: %s", remote.IP.String())
		}

		nfs.processPacket(remote.IP, buffer[:length])
	}
	nfs.wg.Done()
}

// processPacket takes a raw netflow packet, send it to the decoder, updates template cache
// (if there are templates in the packet) and passes the decoded packet over to processFlowSets()
func (nfs *NetflowServer) processPacket(remote net.IP, buffer []byte) {
	length := len(buffer)
	packet, err := nf9.Decode(buffer[:length], remote)
	if err != nil {
		glog.Errorf("nf9packet.Decode: %v", err)
		return
	}

	nfs.updateTemplateCache(remote, packet)
	nfs.processFlowSets(remote, packet.Header.SourceID, packet.DataFlowSets(), int64(packet.Header.UnixSecs), packet)
}

// processFlowSets iterates over flowSets and calls processFlowSet() for each flow set
func (nfs *NetflowServer) processFlowSets(remote net.IP, sourceID uint32, flowSets []*nf9.FlowSet, ts int64, packet *nf9.Packet) {
	addr := remote.String()
	keyParts := make([]string, 3, 3)
	for _, set := range flowSets {
		template := nfs.tmplCache.get(convert.Uint32(remote), sourceID, set.Header.FlowSetID)

		if template == nil {
			templateKey := makeTemplateKey(addr, sourceID, set.Header.FlowSetID, keyParts)
			if *nfs.config.Debug > 0 {
				glog.Warningf("Template for given FlowSet not found: %s", templateKey)
			}
			continue
		}

		records := nf9.DecodeFlowSet(template.Records, *set)
		if records == nil {
			glog.Warning("Error decoding FlowSet")
			continue
		}
		nfs.processFlowSet(template, records, remote, ts, packet)
	}
}

// process generates Flow elements from records and pushes them into the `receiver` channel
func (nfs *NetflowServer) processFlowSet(template *nf9.TemplateRecords, records []nf9.FlowDataRecord, agent net.IP, ts int64, packet *nf9.Packet) {
	fm := generateFieldMap(template)

	for _, r := range records {
		if template.OptionScopes != nil {
			if fm.samplingInterval >= 0 {
				nfs.sampleRateCache.Set(agent, uint64(convert.Uint32(r.Values[fm.samplingInterval])))
			}

			if fm.flowSamplerRandomInterval >= 0 {
				nfs.sampleRateCache.Set(agent, uint64(convert.Uint32(r.Values[fm.flowSamplerRandomInterval])))
			}
			continue
		}

		if fm.family >= 0 {
			switch fm.family {
			case 4:
				atomic.AddUint64(&stats.GlobalStats.Flows4, 1)
			case 6:
				atomic.AddUint64(&stats.GlobalStats.Flows6, 1)
			default:
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

		if !*nfs.config.BGPAugmentation.Enabled {
			if fm.srcAsn >= 0 {
				fl.SrcAs = convert.Uint32(r.Values[fm.srcAsn])
			}

			if fm.dstAsn >= 0 {
				fl.DstAs = convert.Uint32(r.Values[fm.dstAsn])
			}
		}

		fl.Samplerate = nfs.sampleRateCache.Get(agent)

		if *nfs.config.Debug > 2 {
			Dump(&fl)
		}

		nfs.Output <- &fl
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
func DumpTemplate(tmpl *nf9.TemplateRecords) {
	fmt.Printf("Template %d\n", tmpl.Header.TemplateID)
	for rec, i := range tmpl.Records {
		fmt.Printf("%d: %v\n", i, rec)
	}
}

// generateFieldMap processes a TemplateRecord and populates a fieldMap accordingly
// the FieldMap can then be used to read fields from a flow
func generateFieldMap(template *nf9.TemplateRecords) *fieldMap {
	fm := fieldMap{
		srcAddr:                   -1,
		dstAddr:                   -1,
		protocol:                  -1,
		packets:                   -1,
		size:                      -1,
		intIn:                     -1,
		intOut:                    -1,
		nextHop:                   -1,
		family:                    -1,
		vlan:                      -1,
		ts:                        -1,
		srcAsn:                    -1,
		dstAsn:                    -1,
		srcPort:                   -1,
		dstPort:                   -1,
		flowSamplerID:             -1,
		samplingInterval:          -1,
		flowSamplerRandomInterval: -1,
	}

	i := -1
	for _, f := range template.Records {
		i++

		switch f.Type {
		case nf9.IPv4SrcAddr:
			fm.srcAddr = i
			fm.family = 4
		case nf9.IPv6SrcAddr:
			fm.srcAddr = i
			fm.family = 6
		case nf9.IPv4DstAddr:
			fm.dstAddr = i
		case nf9.IPv6DstAddr:
			fm.dstAddr = i
		case nf9.InBytes:
			fm.size = i
		case nf9.Protocol:
			fm.protocol = i
		case nf9.InPkts:
			fm.packets = i
		case nf9.InputSnmp:
			fm.intIn = i
		case nf9.OutputSnmp:
			fm.intOut = i
		case nf9.IPv4NextHop:
			fm.nextHop = i
		case nf9.IPv6NextHop:
			fm.nextHop = i
		case nf9.L4SrcPort:
			fm.srcPort = i
		case nf9.L4DstPort:
			fm.dstPort = i
		case nf9.SrcAs:
			fm.srcAsn = i
		case nf9.DstAs:
			fm.dstAsn = i
		case nf9.SamplingInterval:
			fm.samplingInterval = i
		case nf9.FlowSamplerRandomInterval:
			fm.flowSamplerRandomInterval = i
		}
	}
	return &fm
}

// updateTemplateCache updates the template cache
func (nfs *NetflowServer) updateTemplateCache(remote net.IP, p *nf9.Packet) {
	templRecs := p.GetTemplateRecords()
	for _, tr := range templRecs {
		nfs.tmplCache.set(convert.Uint32(remote), tr.Packet.Header.SourceID, tr.Header.TemplateID, *tr)
	}
}

// makeTemplateKey creates a string of the 3 tuple router address, source id and template id
func makeTemplateKey(addr string, sourceID uint32, templateID uint16, keyParts []string) string {
	keyParts[0] = addr
	keyParts[1] = strconv.Itoa(int(sourceID))
	keyParts[2] = strconv.Itoa(int(templateID))
	return strings.Join(keyParts, "|")
}
