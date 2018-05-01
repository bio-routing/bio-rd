package database

import (
	"bytes"
	"fmt"
	"net"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/avltree"
	"github.com/taktv6/tflow2/iana"
	"github.com/taktv6/tflow2/intfmapper"
	"github.com/taktv6/tflow2/netflow"
)

// BreakdownKey is the key used for the brakedown map
type BreakdownKey [FieldMax]string

// BreakdownMap maps breakdown keys to values
type BreakdownMap map[BreakdownKey]uint64

// BreakdownFlags defines by what fields data should be broken down in a query
type BreakdownFlags struct {
	Family     bool
	SrcAddr    bool
	DstAddr    bool
	Protocol   bool
	IntIn      bool
	IntOut     bool
	NextHop    bool
	SrcAsn     bool
	DstAsn     bool
	NextHopAsn bool
	SrcPfx     bool
	DstPfx     bool
	SrcPort    bool
	DstPort    bool
	IntInName  bool
	IntOutName bool
}

var breakdownLabels = map[int]string{
	FieldFamily:     "Family",
	FieldSrcAddr:    "SrcAddr",
	FieldDstAddr:    "DstAddr",
	FieldProtocol:   "Protocol",
	FieldIntIn:      "IntIn",
	FieldIntOut:     "IntOut",
	FieldNextHop:    "NextHop",
	FieldSrcAs:      "SrcAsn",
	FieldDstAs:      "DstAsn",
	FieldNextHopAs:  "NextHopAsn",
	FieldSrcPfx:     "SrcPfx",
	FieldDstPfx:     "DstPfx",
	FieldSrcPort:    "SrcPort",
	FieldDstPort:    "DstPort",
	FieldIntInName:  "IntInName",
	FieldIntOutName: "IntOutName",
}

// GetBreakdownLabels returns a sorted list of known breakdown labels
func GetBreakdownLabels() []string {
	return []string{
		breakdownLabels[FieldFamily],
		breakdownLabels[FieldSrcAddr],
		breakdownLabels[FieldDstAddr],
		breakdownLabels[FieldProtocol],
		breakdownLabels[FieldIntIn],
		breakdownLabels[FieldIntOut],
		breakdownLabels[FieldNextHop],
		breakdownLabels[FieldSrcAs],
		breakdownLabels[FieldDstAs],
		breakdownLabels[FieldNextHopAs],
		breakdownLabels[FieldSrcPfx],
		breakdownLabels[FieldDstPfx],
		breakdownLabels[FieldSrcPort],
		breakdownLabels[FieldDstPort],
		breakdownLabels[FieldIntInName],
		breakdownLabels[FieldIntOutName],
	}
}

// Join formats the keys and joins them with commas
func (bk *BreakdownKey) Join(format string) string {
	var buffer bytes.Buffer
	for i, value := range bk {
		if value == "" {
			continue
		}
		if buffer.Len() > 0 {
			buffer.WriteRune(',')
		}
		buffer.WriteString(fmt.Sprintf(format, breakdownLabels[i], value))
	}

	return buffer.String()
}

// Set enables the flags in the given list
func (bf *BreakdownFlags) Set(keys []string) error {
	for _, key := range keys {
		switch key {
		case breakdownLabels[FieldFamily]:
			bf.Family = true
		case breakdownLabels[FieldSrcAddr]:
			bf.SrcAddr = true
		case breakdownLabels[FieldDstAddr]:
			bf.DstAddr = true
		case breakdownLabels[FieldProtocol]:
			bf.Protocol = true
		case breakdownLabels[FieldIntIn]:
			bf.IntIn = true
		case breakdownLabels[FieldIntOut]:
			bf.IntOut = true
		case breakdownLabels[FieldNextHop]:
			bf.NextHop = true
		case breakdownLabels[FieldSrcAs]:
			bf.SrcAsn = true
		case breakdownLabels[FieldDstAs]:
			bf.DstAsn = true
		case breakdownLabels[FieldNextHopAs]:
			bf.NextHopAsn = true
		case breakdownLabels[FieldSrcPfx]:
			bf.SrcPfx = true
		case breakdownLabels[FieldDstPfx]:
			bf.DstPfx = true
		case breakdownLabels[FieldSrcPort]:
			bf.SrcPort = true
		case breakdownLabels[FieldDstPort]:
			bf.DstPort = true
		case breakdownLabels[FieldIntInName]:
			bf.IntInName = true
		case breakdownLabels[FieldIntOutName]:
			bf.IntOutName = true

		default:
			return fmt.Errorf("invalid breakdown key: %s", key)
		}
	}
	return nil
}

// Count returns the number of enabled breakdown flags
func (bf *BreakdownFlags) Count() (count int) {

	if bf.Family {
		count++
	}
	if bf.SrcAddr {
		count++
	}
	if bf.DstAddr {
		count++
	}
	if bf.Protocol {
		count++
	}
	if bf.IntIn {
		count++
	}
	if bf.IntOut {
		count++
	}
	if bf.NextHop {
		count++
	}
	if bf.SrcAsn {
		count++
	}
	if bf.DstAsn {
		count++
	}
	if bf.NextHopAsn {
		count++
	}
	if bf.SrcPfx {
		count++
	}
	if bf.DstPfx {
		count++
	}
	if bf.SrcPort {
		count++
	}
	if bf.DstPort {
		count++
	}
	if bf.IntInName {
		count++
	}
	if bf.IntOutName {
		count++
	}

	return
}

// breakdown build all possible relevant keys of flows for flows in tree `node`
// and builds sums for each key in order to allow us to find top combinations
func breakdown(node *avltree.TreeNode, vals ...interface{}) {
	if len(vals) != 5 {
		glog.Errorf("lacking arguments")
		return
	}

	intfMap := vals[0].(intfmapper.InterfaceNameByID)
	iana := vals[1].(*iana.IANA)
	bd := vals[2].(BreakdownFlags)
	sums := vals[3].(*concurrentResSum)
	buckets := vals[4].(BreakdownMap)

	for _, flow := range node.Values {
		fl := flow.(*netflow.Flow)

		key := BreakdownKey{}

		if bd.Family {
			key[FieldFamily] = fmt.Sprintf("%d", fl.Family)
		}
		if bd.SrcAddr {
			key[FieldSrcAddr] = net.IP(fl.SrcAddr).String()
		}
		if bd.DstAddr {
			key[FieldDstAddr] = net.IP(fl.DstAddr).String()
		}
		if bd.Protocol {
			protoMap := iana.GetIPProtocolsByID()
			if _, ok := protoMap[uint8(fl.Protocol)]; ok {
				key[FieldProtocol] = fmt.Sprintf("%s", protoMap[uint8(fl.Protocol)])
			} else {
				key[FieldProtocol] = fmt.Sprintf("%d", fl.Protocol)
			}
		}
		if bd.IntIn {
			key[FieldIntIn] = fmt.Sprintf("%d", fl.IntIn)
		}
		if bd.IntOut {
			key[FieldIntOut] = fmt.Sprintf("%d", fl.IntOut)
		}
		if bd.IntInName {
			if _, ok := intfMap[uint16(fl.IntIn)]; ok {
				name := intfMap[uint16(fl.IntIn)]
				key[FieldIntIn] = fmt.Sprintf("%s", name)
			} else {
				key[FieldIntIn] = fmt.Sprintf("%d", fl.IntIn)
			}
		}
		if bd.IntOutName {
			if _, ok := intfMap[uint16(fl.IntOut)]; ok {
				name := intfMap[uint16(fl.IntOut)]
				key[FieldIntOut] = fmt.Sprintf("%s", name)
			} else {
				key[FieldIntOut] = fmt.Sprintf("%d", fl.IntIn)
			}
		}
		if bd.NextHop {
			key[FieldNextHop] = net.IP(fl.NextHop).String()
		}
		if bd.SrcAsn {
			key[FieldSrcAs] = fmt.Sprintf("%d", fl.SrcAs)
		}
		if bd.DstAsn {
			key[FieldDstAs] = fmt.Sprintf("%d", fl.DstAs)
		}
		if bd.NextHopAsn {
			key[FieldNextHopAs] = fmt.Sprintf("%d", fl.NextHopAs)
		}
		if bd.SrcPfx {
			if fl.SrcPfx != nil {
				key[FieldSrcPfx] = fl.SrcPfx.ToIPNet().String()
			} else {
				key[FieldSrcPfx] = "0.0.0.0/0"
			}
		}
		if bd.DstPfx {
			if fl.DstPfx != nil {
				key[FieldDstPfx] = fl.DstPfx.ToIPNet().String()
			} else {
				key[FieldDstPfx] = "0.0.0.0/0"
			}
		}
		if bd.SrcPort {
			key[FieldSrcPort] = fmt.Sprintf("%d", fl.SrcPort)
		}
		if bd.DstPort {
			key[FieldDstPort] = fmt.Sprintf("%d", fl.DstPort)
		}

		// Build sum for key
		buckets[key] += fl.Size * fl.Samplerate

		// Build overall sum
		sums.Lock.Lock()
		sums.Values[key] += fl.Size * fl.Samplerate
		sums.Lock.Unlock()
	}
}
