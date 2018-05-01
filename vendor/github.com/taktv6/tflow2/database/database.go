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

// Package database keeps track of flow information
package database

import (
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/taktv6/tflow2/iana"
	"github.com/taktv6/tflow2/intfmapper"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/taktv6/tflow2/avltree"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/nfserver"
)

// FlowsByTimeRtr holds all keys (and thus is the only way) to our flows
type FlowsByTimeRtr map[int64]map[string]*TimeGroup

// FlowDatabase represents a flow database object
type FlowDatabase struct {
	flows          FlowsByTimeRtr
	lock           sync.RWMutex
	maxAge         int64
	aggregation    int64
	lastDump       int64
	compLevel      int
	samplerate     int
	storage        *string
	debug          int
	anonymize      bool
	Input          chan *netflow.Flow
	intfMapper     intfmapper.IntfMapperInterface
	agentsNameByIP map[string]string
	iana           *iana.IANA
}

const anyIndex = uint8(0)

// New creates a new FlowDatabase and returns a pointer to it
func New(aggregation int64, maxAge int64, numAddWorker int, debug int, compLevel int, storage *string, anonymize bool, intfMapper intfmapper.IntfMapperInterface, agentsNameByIP map[string]string, iana *iana.IANA) *FlowDatabase {
	flowDB := &FlowDatabase{
		maxAge:         maxAge,
		aggregation:    aggregation,
		compLevel:      compLevel,
		Input:          make(chan *netflow.Flow),
		lastDump:       time.Now().Unix(),
		storage:        storage,
		debug:          debug,
		flows:          make(FlowsByTimeRtr),
		anonymize:      anonymize,
		intfMapper:     intfMapper,
		agentsNameByIP: agentsNameByIP,
		iana:           iana,
	}

	for i := 0; i < numAddWorker; i++ {
		go func() {
			for {
				fl := <-flowDB.Input
				flowDB.Add(fl)
			}
		}()

		go func() {
			for {
				// Set a timer and wait for our next run
				event := time.NewTimer(time.Duration(flowDB.aggregation) * time.Second)
				<-event.C
				flowDB.CleanUp()
			}
		}()

		if flowDB.storage != nil {
			go func() {
				for {
					// Set a timer and wait for our next run
					event := time.NewTimer(time.Duration(flowDB.aggregation) * time.Second)
					<-event.C
					flowDB.Dumper()
				}
			}()
		}
	}
	return flowDB
}

func (fdb *FlowDatabase) getTimeGroup(fl *netflow.Flow, rtr string) *TimeGroup {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	// Check if timestamp entry exists already. If not, create it.
	flows, ok := fdb.flows[fl.Timestamp]
	if !ok {
		flows = make(map[string]*TimeGroup)
		fdb.flows[fl.Timestamp] = flows
	}

	// Check if router entry exists already. If not, create it.
	timeGroup, ok := flows[rtr]
	if !ok {
		timeGroup = &TimeGroup{
			Any:               newMapTree(),
			SrcAddr:           newMapTree(),
			DstAddr:           newMapTree(),
			Protocol:          newMapTree(),
			IntIn:             newMapTree(),
			IntOut:            newMapTree(),
			NextHop:           newMapTree(),
			SrcAs:             newMapTree(),
			DstAs:             newMapTree(),
			NextHopAs:         newMapTree(),
			SrcPfx:            newMapTree(),
			DstPfx:            newMapTree(),
			SrcPort:           newMapTree(),
			DstPort:           newMapTree(),
			InterfaceIDByName: fdb.intfMapper.GetInterfaceIDByName(rtr),
		}
		flows[rtr] = timeGroup
	}

	return timeGroup
}

// Add adds flow `fl` to database fdb
func (fdb *FlowDatabase) Add(fl *netflow.Flow) {
	// build indices for map access
	rtrip := net.IP(fl.Router)

	if _, ok := fdb.agentsNameByIP[rtrip.String()]; !ok {
		glog.Warningf("Unknown flow source: %s", rtrip.String())
		return
	}

	rtrName := fdb.agentsNameByIP[rtrip.String()]
	timeGroup := fdb.getTimeGroup(fl, rtrName)

	fdb.lock.RLock()
	defer fdb.lock.RUnlock()
	if _, ok := fdb.flows[fl.Timestamp]; !ok {
		glog.Warningf("stopped adding data for %d: already deleted", fl.Timestamp)
		return
	}

	// Insert into indices
	timeGroup.Any.Insert(anyIndex, fl)
	timeGroup.SrcAddr.Insert(net.IP(fl.SrcAddr), fl)
	timeGroup.DstAddr.Insert(net.IP(fl.DstAddr), fl)
	timeGroup.Protocol.Insert(byte(fl.Protocol), fl)
	timeGroup.IntIn.Insert(uint16(fl.IntIn), fl)
	timeGroup.IntOut.Insert(uint16(fl.IntOut), fl)
	timeGroup.NextHop.Insert(net.IP(fl.NextHop), fl)
	timeGroup.SrcAs.Insert(fl.SrcAs, fl)
	timeGroup.DstAs.Insert(fl.DstAs, fl)
	timeGroup.NextHopAs.Insert(fl.NextHopAs, fl)
	timeGroup.SrcPfx.Insert(fl.SrcPfx.String(), fl)
	timeGroup.DstPfx.Insert(fl.DstPfx.String(), fl)
	timeGroup.SrcPort.Insert(fl.SrcPort, fl)
	timeGroup.DstPort.Insert(fl.DstPort, fl)
}

// CurrentTimeslot returns the beginning of the current timeslot
func (fdb *FlowDatabase) CurrentTimeslot() int64 {
	now := time.Now().Unix()
	return now - now%fdb.aggregation
}

// AggregationPeriod returns the configured aggregation period
func (fdb *FlowDatabase) AggregationPeriod() int64 {
	return fdb.aggregation
}

// CleanUp deletes all flows from database `fdb` that are older than `maxAge` seconds
func (fdb *FlowDatabase) CleanUp() {
	now := fdb.CurrentTimeslot()

	fdb.lock.Lock()
	defer fdb.lock.Unlock()
	for ts := range fdb.flows {
		if ts < now-fdb.maxAge {
			delete(fdb.flows, ts)
		}
	}
}

// Dumper dumps all flows in `fdb` to hard drive that haven't been dumped yet
func (fdb *FlowDatabase) Dumper() {
	fdb.lock.RLock()
	defer fdb.lock.RUnlock()

	min := atomic.LoadInt64(&fdb.lastDump)
	max := fdb.CurrentTimeslot() - 2*fdb.aggregation
	atomic.StoreInt64(&fdb.lastDump, max)

	for ts := range fdb.flows {
		if ts < min || ts > max {
			continue
		}
		for router := range fdb.flows[ts] {
			go fdb.dumpToDisk(ts, router)
		}
		atomic.StoreInt64(&fdb.lastDump, ts)
	}
}

func (fdb *FlowDatabase) dumpToDisk(ts int64, router string) {
	if fdb.storage == nil {
		return
	}

	fdb.lock.RLock()
	tg := fdb.flows[ts][router]
	tree := fdb.flows[ts][router].Any.Get(anyIndex)
	fdb.lock.RUnlock()

	// Create flow proto buffer
	flows := &netflow.Flows{}

	// Populate interface mapping
	for name, id := range tg.InterfaceIDByName {
		flows.InterfaceMapping = append(flows.InterfaceMapping, &netflow.Intf{
			Id:   uint32(id),
			Name: name,
		})
	}

	// Write flows into `flows` proto buffer
	tree.Each(dump, fdb.anonymize, flows)

	if fdb.debug > 1 {
		glog.Warningf("flows contains %d flows", len(flows.Flows))
	}

	// Marshal flows into proto buffer
	buffer, err := proto.Marshal(flows)
	if err != nil {
		glog.Errorf("unable to marshal flows into pb: %v", err)
		return
	}

	// Create dir if doesn't exist
	ymd := fmt.Sprintf("%04d-%02d-%02d", time.Unix(ts, 0).Year(), time.Unix(ts, 0).Month(), time.Unix(ts, 0).Day())
	os.Mkdir(fmt.Sprintf("%s/%s", *fdb.storage, ymd), 0700)

	// Create file
	fh, err := os.Create(fmt.Sprintf("%s/%s/nf-%d-%s.tflow2.pb.gzip", *fdb.storage, ymd, ts, router))
	if err != nil {
		glog.Errorf("couldn't create file: %v", err)
	}
	defer fh.Close()

	// Compress data before writing it out to the disk
	gz, err := gzip.NewWriterLevel(fh, fdb.compLevel)
	if err != nil {
		glog.Errorf("invalud gzip compression level: %v", err)
		return
	}

	// Compress and write file
	_, err = gz.Write(buffer)
	gz.Close()

	if err != nil {
		glog.Errorf("failed to write file: %v", err)
	}
}

func dump(node *avltree.TreeNode, vals ...interface{}) {
	anonymize := vals[0].(bool)
	flows := vals[1].(*netflow.Flows)

	for _, f := range node.Values {
		flow := f.(*netflow.Flow)
		flowcopy := *flow

		if anonymize {
			// Remove information about particular IP addresses for privacy reason
			flowcopy.SrcAddr = []byte{0, 0, 0, 0}
			flowcopy.DstAddr = []byte{0, 0, 0, 0}
		}

		flows.Flows = append(flows.Flows, &flowcopy)
	}
}

// ptrIsSmaller checks if uintptr c1 is smaller than uintptr c2
func ptrIsSmaller(c1 interface{}, c2 interface{}) bool {
	x := uintptr(unsafe.Pointer(c1.(*netflow.Flow)))
	y := uintptr(unsafe.Pointer(c2.(*netflow.Flow)))

	return x < y
}

// uint64IsSmaller checks if uint64 c1 is smaller than uint64 c2
func uint64IsSmaller(c1 interface{}, c2 interface{}) bool {
	return c1.(uint64) < c2.(uint64)
}

// uint64IsSmaller checks if int64 c1 is small than int64 c2
func int64IsSmaller(c1 interface{}, c2 interface{}) bool {
	return c1.(int64) < c2.(int64)
}

// dumpFlows dumps all flows a tree `tree`
func dumpFlows(tree *avltree.TreeNode) {
	tree.Each(printNode)
}

// printNode dumps the flow of `node` on the screen
func printNode(node *avltree.TreeNode, vals ...interface{}) {
	for _, fl := range node.Values {
		nfserver.Dump(fl.(*netflow.Flow))
	}
}
