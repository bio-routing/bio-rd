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

package database

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/taktv6/tflow2/avltree"
	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/intfmapper"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/stats"
)

// These constants are used in communication with the frontend
const (
	OpEqual   = 0
	OpUnequal = 1
	OpSmaller = 2
	OpGreater = 3
)

// These constants are only used internally
const (
	FieldTimestamp = iota
	FieldAgent
	FieldFamily
	FieldSrcAddr
	FieldDstAddr
	FieldProtocol
	FieldIntIn
	FieldIntOut
	FieldNextHop
	FieldSrcAs
	FieldDstAs
	FieldNextHopAs
	FieldSrcPfx
	FieldDstPfx
	FieldSrcPort
	FieldDstPort
	FieldIntInName
	FieldIntOutName
	FieldMax
)

var fieldNames = map[string]int{
	"Timestamp":  FieldTimestamp,
	"Agent":      FieldAgent,
	"Family":     FieldFamily,
	"SrcAddr":    FieldSrcAddr,
	"DstAddr":    FieldDstAddr,
	"Protocol":   FieldProtocol,
	"IntIn":      FieldIntIn,
	"IntOut":     FieldIntOut,
	"NextHop":    FieldNextHop,
	"SrcAs":      FieldSrcAs,
	"DstAs":      FieldDstAs,
	"NextHopAs":  FieldNextHopAs,
	"SrcPfx":     FieldSrcPfx,
	"DstPfx":     FieldDstPfx,
	"SrcPort":    FieldSrcPort,
	"DstPort":    FieldDstPort,
	"IntInName":  FieldIntInName,
	"IntOutName": FieldIntOutName,
}

type void struct{}

// Condition represents a query condition
type Condition struct {
	Field    int
	Operator int
	Operand  []byte
}

// Conditions represents a set of conditions of a query
type Conditions []Condition

// Query is the internal representation of a query
type Query struct {
	Cond      Conditions
	Breakdown BreakdownFlags
	TopN      int
}

type concurrentResSum struct {
	Values BreakdownMap
	Lock   sync.Mutex
}

// GetFieldByName returns the internal number of a field
func GetFieldByName(name string) int {
	if i, found := fieldNames[name]; found {
		return i
	}
	return -1
}

// Includes checks if the given field and operator is included in the list
func (conditions Conditions) Includes(field int, operator int) bool {
	for _, cond := range conditions {
		if cond.Field == field && cond.Operator == operator {
			return true
		}
	}
	return false
}

// loadFromDisc loads netflow data from disk into in memory data structure
func (fdb *FlowDatabase) loadFromDisc(ts int64, agent string, query Query, resSum *concurrentResSum) (BreakdownMap, error) {
	if fdb.storage == nil {
		return nil, fmt.Errorf("Disk storage is disabled")
	}

	res := avltree.New()
	ymd := fmt.Sprintf("%04d-%02d-%02d", time.Unix(ts, 0).Year(), time.Unix(ts, 0).Month(), time.Unix(ts, 0).Day())
	filename := fmt.Sprintf("%s/%s/nf-%d-%s.tflow2.pb.gzip", *fdb.storage, ymd, ts, agent)
	fh, err := os.Open(filename)
	if err != nil {
		if fdb.debug > 0 {
			glog.Errorf("unable to open file: %v", err)
		}
		return nil, err
	}
	if fdb.debug > 1 {
		glog.Infof("successfully opened file: %s", filename)
	}
	defer fh.Close()

	gz, err := gzip.NewReader(fh)
	if err != nil {
		glog.Errorf("unable to create gzip reader: %v", err)
		return nil, err
	}
	defer gz.Close()

	buffer, err := ioutil.ReadAll(gz)
	if err != nil {
		glog.Errorf("unable to gunzip: %v", err)
		return nil, err
	}

	// Unmarshal protobuf
	flows := netflow.Flows{}
	err = proto.Unmarshal(buffer, &flows)
	if err != nil {
		glog.Errorf("unable to unmarshal protobuf: %v", err)
		return nil, err
	}

	// Create interface mapping
	interfaceIDByName := make(intfmapper.InterfaceIDByName)
	for _, m := range flows.InterfaceMapping {
		interfaceIDByName[m.Name] = uint16(m.Id)
	}

	if fdb.debug > 1 {
		glog.Infof("file %s contains %d flows", filename, len(flows.Flows))
	}

	// Validate flows and add them to res tree
	for _, fl := range flows.Flows {
		if validateFlow(fl, query, interfaceIDByName) {
			res.Insert(fl, fl, ptrIsSmaller)
		}
	}

	// Breakdown
	resTime := make(BreakdownMap)
	res.Each(breakdown, fdb.intfMapper.GetInterfaceNameByID(agent), fdb.iana, query.Breakdown, resSum, resTime)

	return resTime, err
}

func validateFlow(fl *netflow.Flow, query Query, interfaceIDByName intfmapper.InterfaceIDByName) bool {
	for _, c := range query.Cond {
		switch c.Field {
		case FieldTimestamp:
			continue
		case FieldAgent:
			continue
		case FieldFamily:
			if fl.Family != uint32(convert.Uint16b(c.Operand)) {
				return false
			}
			continue
		case FieldProtocol:
			if fl.Protocol != uint32(convert.Uint16b(c.Operand)) {
				return false
			}
			continue
		case FieldSrcAddr:
			if !net.IP(fl.SrcAddr).Equal(net.IP(c.Operand)) {
				return false
			}
			continue
		case FieldDstAddr:
			if !net.IP(fl.DstAddr).Equal(net.IP(c.Operand)) {
				return false
			}
			continue
		case FieldIntIn:
			if fl.IntIn != uint32(convert.Uint16b(c.Operand)) {
				return false
			}
			continue
		case FieldIntOut:
			if fl.IntOut != uint32(convert.Uint16b(c.Operand)) {
				return false
			}
			continue
		case FieldNextHop:
			if !net.IP(fl.NextHop).Equal(net.IP(c.Operand)) {
				return false
			}
			continue
		case FieldSrcAs:
			if fl.SrcAs != convert.Uint32b(c.Operand) {
				return false
			}
			continue
		case FieldDstAs:
			if fl.DstAs != convert.Uint32b(c.Operand) {
				return false
			}
			continue
		case FieldNextHopAs:
			if fl.NextHopAs != convert.Uint32b(c.Operand) {
				return false
			}
		case FieldSrcPort:
			if fl.SrcPort != uint32(convert.Uint16b(c.Operand)) {
				return false
			}
			continue
		case FieldDstPort:
			if fl.DstPort != uint32(convert.Uint16b(c.Operand)) {
				return false
			}
			continue
		case FieldSrcPfx:
			if fl.SrcPfx.String() != string(c.Operand) {
				return false
			}
			continue
		case FieldDstPfx:
			if fl.DstPfx.String() != string(c.Operand) {
				return false
			}
			continue
		case FieldIntInName:
			id := interfaceIDByName[string(c.Operand)]
			if uint16(fl.IntIn) != id {
				return false
			}
			continue
		case FieldIntOutName:
			id := interfaceIDByName[string(c.Operand)]
			if uint16(fl.IntOut) != id {
				return false
			}
			continue
		}
	}
	return true
}

func (fdb *FlowDatabase) getAgent(q *Query) (string, error) {
	rtr := ""
	for _, c := range q.Cond {
		if c.Field == FieldAgent {
			rtr = string(c.Operand)
		}
	}
	if rtr == "" {
		glog.Warningf("Agent is mandatory cirteria")
		return "", fmt.Errorf("Agent criteria not found")
	}

	return rtr, nil
}

func (fdb *FlowDatabase) getStartEndTimes(q *Query) (start int64, end int64, err error) {
	end = time.Now().Unix()
	for _, c := range q.Cond {
		if c.Field != FieldTimestamp {
			continue
		}
		switch c.Operator {
		case OpGreater:
			start = int64(convert.Uint64b(c.Operand))
		case OpSmaller:
			end = int64(convert.Uint64b(c.Operand))
		case OpEqual:
			start = int64(convert.Uint64b(c.Operand))
			end = start
		}
	}

	// Align start point to `aggregation` raster
	start = start - (start % fdb.aggregation)

	return
}

func (fdb *FlowDatabase) getResultByTS(resSum *concurrentResSum, ts int64, q *Query, rtr string) BreakdownMap {
	// timeslot in memory?
	fdb.lock.RLock()
	timeGroups, ok := fdb.flows[ts]
	fdb.lock.RUnlock()

	if !ok {
		// not in memory, try to load from disk
		result, _ := fdb.loadFromDisc(ts, rtr, *q, resSum)
		return result
	}

	if timeGroups[rtr] == nil {
		glog.Infof("TG of %s is nil", rtr)
		return map[BreakdownKey]uint64{}
	}

	return timeGroups[rtr].filterAndBreakdown(resSum, q, fdb.iana, fdb.intfMapper.GetInterfaceNameByID(rtr))
}

func (fdb *FlowDatabase) getTopKeys(resSum *concurrentResSum, topN int) map[BreakdownKey]void {
	// Build Tree Bytes -> Key to allow efficient finding of top n flows
	var btree = avltree.New()
	for k, b := range resSum.Values {
		btree.Insert(b, k, uint64IsSmaller)
	}

	// Find top n keys
	topKeysList := btree.TopN(topN)
	topKeys := make(map[BreakdownKey]void)
	for _, v := range topKeysList {
		topKeys[v.(BreakdownKey)] = void{}
	}

	return topKeys
}

// RunQuery executes a query and returns the result
func (fdb *FlowDatabase) RunQuery(q *Query) (*Result, error) {
	queryStart := time.Now()
	stats.GlobalStats.Queries++

	start, end, err := fdb.getStartEndTimes(q)
	if err != nil {
		return nil, fmt.Errorf("Failed to Start/End times: %v", err)
	}

	rtr, err := fdb.getAgent(q)
	if err != nil {
		return nil, fmt.Errorf("Failed to get router: %v", err)
	}

	// resSum holds a sum per breakdown key over all timestamps
	resSum := &concurrentResSum{
		Values: make(BreakdownMap),
	}

	// resTime holds individual sums per breakdown key and ts
	resTime := make(map[int64]BreakdownMap)
	resMtx := sync.Mutex{}
	resWg := sync.WaitGroup{}

	for ts := start; ts <= end; ts += fdb.aggregation {
		glog.Infof("RunQuery: start timeslot %d", ts)
		resWg.Add(1)
		go func(ts int64) {
			result := fdb.getResultByTS(resSum, ts, q, rtr)

			if result != nil {
				glog.Infof("RunQuery: data in timeslot %d", ts)
				resMtx.Lock()
				resTime[ts] = result
				resMtx.Unlock()
			}
			resWg.Done()
		}(ts)
	}

	resWg.Wait()

	// Find all timestamps we have and get them sorted
	tsTree := avltree.New()
	for ts := range resTime {
		tsTree.Insert(ts, ts, int64IsSmaller)
	}

	// Generate topKeys if required
	var topKeys map[BreakdownKey]void
	if q.TopN > 0 {
		topKeys = fdb.getTopKeys(resSum, q.TopN)
	}

	timestamps := make([]int64, 0)
	for _, ts := range tsTree.Dump() {
		timestamps = append(timestamps, ts.(int64))
	}

	glog.Infof("Query %s took %d ns\n", q, time.Since(queryStart))

	return &Result{
		TopKeys:     topKeys,
		Timestamps:  timestamps,
		Data:        resTime,
		Aggregation: fdb.aggregation,
	}, nil
}
