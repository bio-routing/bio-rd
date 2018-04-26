package database

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/taktv6/tflow2/convert"
	"github.com/taktv6/tflow2/iana"
	"github.com/taktv6/tflow2/intfmapper"
	"github.com/taktv6/tflow2/netflow"
)

type intfMapper struct {
}

func (m *intfMapper) GetInterfaceIDByName(agent string) intfmapper.InterfaceIDByName {
	return intfmapper.InterfaceIDByName{
		"xe-0/0/1": 1,
		"xe-0/0/2": 2,
		"xe-0/0/3": 3,
	}
}

func (m *intfMapper) GetInterfaceNameByID(agent string) intfmapper.InterfaceNameByID {
	return intfmapper.InterfaceNameByID{
		1: "xe-0/0/1",
		2: "xe-0/0/2",
		3: "xe-0/0/3",
	}
}

func TestQuery(t *testing.T) {
	minute := int64(60)
	hour := int64(3600)

	ts1 := int64(3600)
	ts1 = ts1 - ts1%minute

	tests := []struct {
		name           string
		flows          []*netflow.Flow
		query          *Query
		expectedResult Result
	}{
		{
			// Testcase: 2 flows from AS100 to AS300 and back (TCP session).
			name: "Test 1",
			flows: []*netflow.Flow{
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{10, 0, 0, 1},
					DstAddr:    []byte{30, 0, 0, 1},
					Protocol:   6,
					SrcPort:    12345,
					DstPort:    443,
					Packets:    2,
					Size:       1000,
					IntIn:      1,
					IntOut:     3,
					NextHop:    []byte{30, 0, 0, 100},
					SrcAs:      100,
					DstAs:      300,
					NextHopAs:  300,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{10, 0, 0, 1},
					DstAddr:    []byte{30, 0, 0, 2},
					Protocol:   6,
					SrcPort:    12345,
					DstPort:    443,
					Packets:    2,
					Size:       1000,
					IntIn:      1,
					IntOut:     3,
					NextHop:    []byte{30, 0, 0, 100},
					SrcAs:      100,
					DstAs:      300,
					NextHopAs:  300,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{30, 0, 0, 1},
					DstAddr:    []byte{10, 0, 0, 1},
					Protocol:   6,
					SrcPort:    443,
					DstPort:    12345,
					Packets:    5,
					Size:       10000,
					IntIn:      3,
					IntOut:     1,
					NextHop:    []byte{10, 0, 0, 100},
					SrcAs:      300,
					DstAs:      100,
					NextHopAs:  100,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{30, 0, 0, 2},
					DstAddr:    []byte{10, 0, 0, 1},
					Protocol:   6,
					SrcPort:    443,
					DstPort:    12345,
					Packets:    5,
					Size:       10000,
					IntIn:      3,
					IntOut:     1,
					NextHop:    []byte{10, 0, 0, 100},
					SrcAs:      300,
					DstAs:      100,
					NextHopAs:  100,
					Samplerate: 4,
					Timestamp:  ts1,
				},
			},
			query: &Query{
				Cond: []Condition{
					{
						Field:    FieldAgent,
						Operator: OpEqual,
						Operand:  []byte("test01.pop01"),
					},
					{
						Field:    FieldTimestamp,
						Operator: OpGreater,
						Operand:  convert.Uint64Byte(uint64(ts1 - 3*minute)),
					},
					{
						Field:    FieldTimestamp,
						Operator: OpSmaller,
						Operand:  convert.Uint64Byte(uint64(ts1 + minute)),
					},
					{
						Field:    FieldIntOut,
						Operator: OpEqual,
						Operand:  convert.Uint16Byte(uint16(1)),
					},
				},
				Breakdown: BreakdownFlags{
					SrcAddr: true,
					DstAddr: true,
				},
				TopN: 100,
			},
			expectedResult: Result{
				TopKeys: map[BreakdownKey]void{
					BreakdownKey{
						FieldSrcAddr: "30.0.0.1",
						FieldDstAddr: "10.0.0.1",
					}: void{},
					BreakdownKey{
						FieldSrcAddr: "30.0.0.2",
						FieldDstAddr: "10.0.0.1",
					}: void{},
				},
				Timestamps: []int64{
					ts1,
				},
				Data: map[int64]BreakdownMap{
					ts1: BreakdownMap{
						BreakdownKey{
							FieldSrcAddr: "30.0.0.1",
							FieldDstAddr: "10.0.0.1",
						}: 40000,
						BreakdownKey{
							FieldSrcAddr: "30.0.0.2",
							FieldDstAddr: "10.0.0.1",
						}: 40000,
					},
				},
				Aggregation: minute,
			},
		},

		{
			// Testcase: 2 flows from AS100 to AS300 and back (TCP session).
			// Opposite direction of Test 1
			name: "Test 2",
			flows: []*netflow.Flow{
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{10, 0, 0, 1},
					DstAddr:    []byte{30, 0, 0, 1},
					Protocol:   6,
					SrcPort:    12345,
					DstPort:    443,
					Packets:    2,
					Size:       1000,
					IntIn:      1,
					IntOut:     3,
					NextHop:    []byte{30, 0, 0, 100},
					SrcAs:      100,
					DstAs:      300,
					NextHopAs:  300,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{10, 0, 0, 1},
					DstAddr:    []byte{30, 0, 0, 2},
					Protocol:   6,
					SrcPort:    12345,
					DstPort:    443,
					Packets:    2,
					Size:       1000,
					IntIn:      1,
					IntOut:     3,
					NextHop:    []byte{30, 0, 0, 100},
					SrcAs:      100,
					DstAs:      300,
					NextHopAs:  300,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{30, 0, 0, 1},
					DstAddr:    []byte{10, 0, 0, 1},
					Protocol:   6,
					SrcPort:    443,
					DstPort:    12345,
					Packets:    5,
					Size:       10000,
					IntIn:      3,
					IntOut:     1,
					NextHop:    []byte{10, 0, 0, 100},
					SrcAs:      300,
					DstAs:      100,
					NextHopAs:  100,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{30, 0, 0, 2},
					DstAddr:    []byte{10, 0, 0, 1},
					Protocol:   6,
					SrcPort:    443,
					DstPort:    12345,
					Packets:    5,
					Size:       10000,
					IntIn:      3,
					IntOut:     1,
					NextHop:    []byte{10, 0, 0, 100},
					SrcAs:      300,
					DstAs:      100,
					NextHopAs:  100,
					Samplerate: 4,
					Timestamp:  ts1,
				},
			},
			query: &Query{
				Cond: []Condition{
					{
						Field:    FieldAgent,
						Operator: OpEqual,
						Operand:  []byte("test01.pop01"),
					},
					{
						Field:    FieldTimestamp,
						Operator: OpGreater,
						Operand:  convert.Uint64Byte(uint64(ts1 - 3*minute)),
					},
					{
						Field:    FieldTimestamp,
						Operator: OpSmaller,
						Operand:  convert.Uint64Byte(uint64(ts1 + minute)),
					},
					{
						Field:    FieldIntOut,
						Operator: OpEqual,
						Operand:  convert.Uint16Byte(uint16(3)),
					},
				},
				Breakdown: BreakdownFlags{
					SrcAddr: true,
					DstAddr: true,
				},
				TopN: 100,
			},
			expectedResult: Result{
				TopKeys: map[BreakdownKey]void{
					BreakdownKey{
						FieldSrcAddr: "10.0.0.1",
						FieldDstAddr: "30.0.0.1",
					}: void{},
					BreakdownKey{
						FieldSrcAddr: "10.0.0.1",
						FieldDstAddr: "30.0.0.2",
					}: void{},
				},
				Timestamps: []int64{
					ts1,
				},
				Data: map[int64]BreakdownMap{
					ts1: BreakdownMap{
						BreakdownKey{
							FieldSrcAddr: "10.0.0.1",
							FieldDstAddr: "30.0.0.1",
						}: 4000,
						BreakdownKey{
							FieldSrcAddr: "10.0.0.1",
							FieldDstAddr: "30.0.0.2",
						}: 4000,
					},
				},
				Aggregation: minute,
			},
		},

		{
			// Testcase: 2 flows from AS100 to AS300 and back (TCP session).
			// Test TopN function
			name: "Test 3",
			flows: []*netflow.Flow{
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{10, 0, 0, 1},
					DstAddr:    []byte{30, 0, 0, 1},
					Protocol:   6,
					SrcPort:    12345,
					DstPort:    443,
					Packets:    2,
					Size:       1001,
					IntIn:      1,
					IntOut:     3,
					NextHop:    []byte{30, 0, 0, 100},
					SrcAs:      100,
					DstAs:      300,
					NextHopAs:  300,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{10, 0, 0, 1},
					DstAddr:    []byte{30, 0, 0, 2},
					Protocol:   6,
					SrcPort:    12345,
					DstPort:    443,
					Packets:    2,
					Size:       1000,
					IntIn:      1,
					IntOut:     3,
					NextHop:    []byte{30, 0, 0, 100},
					SrcAs:      100,
					DstAs:      300,
					NextHopAs:  300,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{30, 0, 0, 1},
					DstAddr:    []byte{10, 0, 0, 1},
					Protocol:   6,
					SrcPort:    443,
					DstPort:    12345,
					Packets:    5,
					Size:       10000,
					IntIn:      3,
					IntOut:     1,
					NextHop:    []byte{10, 0, 0, 100},
					SrcAs:      300,
					DstAs:      100,
					NextHopAs:  100,
					Samplerate: 4,
					Timestamp:  ts1,
				},
				&netflow.Flow{
					Router:     []byte{1, 2, 3, 4},
					Family:     4,
					SrcAddr:    []byte{30, 0, 0, 2},
					DstAddr:    []byte{10, 0, 0, 1},
					Protocol:   6,
					SrcPort:    443,
					DstPort:    12345,
					Packets:    5,
					Size:       10000,
					IntIn:      3,
					IntOut:     1,
					NextHop:    []byte{10, 0, 0, 100},
					SrcAs:      300,
					DstAs:      100,
					NextHopAs:  100,
					Samplerate: 4,
					Timestamp:  ts1,
				},
			},
			query: &Query{
				Cond: []Condition{
					{
						Field:    FieldAgent,
						Operator: OpEqual,
						Operand:  []byte("test01.pop01"),
					},
					{
						Field:    FieldTimestamp,
						Operator: OpGreater,
						Operand:  convert.Uint64Byte(uint64(ts1 - 3*minute)),
					},
					{
						Field:    FieldTimestamp,
						Operator: OpSmaller,
						Operand:  convert.Uint64Byte(uint64(ts1 + minute)),
					},
					{
						Field:    FieldIntOut,
						Operator: OpEqual,
						Operand:  convert.Uint16Byte(uint16(3)),
					},
				},
				Breakdown: BreakdownFlags{
					SrcAddr: true,
					DstAddr: true,
				},
				TopN: 1,
			},
			expectedResult: Result{
				TopKeys: map[BreakdownKey]void{
					BreakdownKey{
						FieldSrcAddr: "10.0.0.1",
						FieldDstAddr: "30.0.0.1",
					}: void{},
				},
				Timestamps: []int64{
					ts1,
				},
				Data: map[int64]BreakdownMap{
					ts1: BreakdownMap{
						BreakdownKey{
							FieldSrcAddr: "10.0.0.1",
							FieldDstAddr: "30.0.0.1",
						}: 4004,
						BreakdownKey{
							FieldSrcAddr: "10.0.0.1",
							FieldDstAddr: "30.0.0.2",
						}: 4000,
					},
				},
				Aggregation: minute,
			},
		},
	}

	for _, test := range tests {
		fdb := New(minute, hour, 1, 0, 6, nil, false, &intfMapper{}, map[string]string{
			net.IP([]byte{1, 2, 3, 4}).String(): "test01.pop01",
		}, iana.New())

		for _, flow := range test.flows {
			fdb.Input <- flow
		}

		time.Sleep(time.Second)

		result, err := fdb.RunQuery(test.query)
		if err != nil {
			t.Errorf("Unexpected error on RunQuery: %v", err)
		}

		assert.Equal(t, test.expectedResult, *result, test.name)
	}
}

func dumpRes(res Result) {
	for ts := range res.Data {
		for k, v := range res.Data[ts] {
			fmt.Printf("TS: %d\tKey: %v\t %d\n", ts, k, v)
		}
	}
}
