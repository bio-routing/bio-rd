package server

import (
	"sync"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btime "github.com/bio-routing/bio-rd/util/time"

	log "github.com/sirupsen/logrus"
)

/*
* TODOs:
* What to do when TLVs in Hellos change (e.g. IP address is added?)
 */

type neighbor struct {
	name                   string
	sysID                  types.SystemID
	extendedLocalCircuitID uint32
	nm                     *neighborManager
	state                  uint8
	stateMu                sync.RWMutex
	timeout                time.Time
	timeoutMu              sync.Mutex
	priority               uint8
	ipAddresses            []bnet.IP
	protocols              []uint8
	areas                  []types.AreaID
	adjCheckTicker         btime.Ticker
	wg                     sync.WaitGroup
	done                   chan struct{}
}

func (nm *neighborManager) neighborFromP2PHello(hello *packet.P2PHello) *neighbor {
	n := &neighbor{
		sysID:       hello.SystemID,
		nm:          nm,
		state:       packet.P2PAdjStateInit,
		timeout:     time.Now().Add(time.Duration(hello.HoldingTimer) * time.Second),
		ipAddresses: make([]bnet.IP, 0),
		protocols:   make([]uint8, 0),
		areas:       make([]types.AreaID, 0),
		done:        make(chan struct{}),
	}

	for _, tlv := range hello.TLVs {
		switch tlv.Type() {
		case packet.ProtocolsSupportedTLVType:
			x := tlv.Value().(packet.ProtocolsSupportedTLV)
			for _, p := range x.NetworkLayerProtocolIDs {
				n.protocols = append(n.protocols, p)
			}
		case packet.IPInterfaceAddressesTLVType:
			ipIntAddrs := tlv.Value().(packet.IPInterfaceAddressesTLV)
			for _, a := range ipIntAddrs.IPv4Addresses {
				n.ipAddresses = append(n.ipAddresses, bnet.IPv4(a))
			}
		case packet.AreaAddressesTLVType:
			x := tlv.Value().(*packet.AreaAddressesTLV)
			for _, a := range x.AreaIDs {
				n.areas = append(n.areas, a)
			}
		case packet.P2PAdjacencyStateTLVType:
			x := tlv.Value().(packet.P2PAdjacencyStateTLV)
			n.extendedLocalCircuitID = x.ExtendedLocalCircuitID
		}
	}

	return n
}

func (n *neighbor) dispose() {
	close(n.done)
}

func (n *neighbor) fields() log.Fields {
	return log.Fields{
		"protocol":  "IS-IS",
		"component": "neighbor",
		"interface": n.nm.netIfa.name,
		"level":     n.nm.level,
		"sysID":     n.sysID,
	}
}

// adjChecker checks if a timeout has occured
func (n *neighbor) adjChecker() {
	defer n.wg.Done()
	defer log.WithFields(n.fields()).Debug("Stopping adjacency timeout checker")

	n.adjCheckTicker = btime.NewBIOTicker(time.Second)
	defer n.adjCheckTicker.Stop()

	log.WithFields(n.fields()).Debug("Starting adjacency timeout checker")
	for {
		select {
		case <-n.done:
			return
		case <-n.adjCheckTicker.C():
			if n.timedOut() {
				log.WithFields(n.fields()).Infof("Timeout event")
				return
			}
		}
	}
}

func (n *neighbor) timedOut() bool {
	n.timeoutMu.Lock()
	defer n.timeoutMu.Unlock()

	return n.timeout.Before(time.Now())
}

func (n *neighbor) updateTimeout(to time.Time) {
	log.WithFields(n.fields()).Debug("Timeout updated")
	n.timeoutMu.Lock()
	defer n.timeoutMu.Unlock()

	n.timeout = to
}

func (n *neighbor) getState() uint8 {
	n.stateMu.Lock()
	defer n.stateMu.Unlock()

	return n.state
}

func (n *neighbor) setState(s uint8) {
	n.stateMu.Lock()
	defer n.stateMu.Unlock()

	n.state = s
}

func (n *neighbor) processP2PHello(hello *packet.P2PHello) error {
	n.updateTimeout(time.Now().Add(time.Second * time.Duration(hello.HoldingTimer)))

	p2pAdjState := getP2PAdjTLV(hello.TLVs)
	if p2pAdjState != nil {
		if n.p2pAdjTLVContainsSelf(p2pAdjState) {
			if n.getState() != packet.P2PAdjStateUp {
				log.WithFields(n.fields()).Infof("Adjacency reaches up state")
				n.setState(packet.P2PAdjStateUp)

				n.nm.server.regenerateL2LSP()
				n.getLSDB().sendCSNPs(n.nm.netIfa) // TODO: Make this a go routinge that sends periodically
				n.getLSDB().setSRMAllLSPs(n.nm.netIfa)
			}
		}
	}

	return nil
}

func (n *neighbor) getLSDB() *lsdb {
	if n.nm.level == 1 {
		return n.nm.server.lsdbL1
	}

	return n.nm.server.lsdbL2
}

func getP2PAdjTLV(tlvs []packet.TLV) *packet.P2PAdjacencyStateTLV {
	for _, tlv := range tlvs {
		if tlv.Type() == packet.P2PAdjacencyStateTLVType {
			x := tlv.Value().(packet.P2PAdjacencyStateTLV)
			return &x
		}
	}

	return nil
}

func (n *neighbor) p2pAdjTLVContainsSelf(t *packet.P2PAdjacencyStateTLV) bool {
	return t.NeighborSystemID == n.nm.netIfa.srv.nets[0].SystemID && t.NeighborExtendedLocalCircuitID == uint32(n.nm.netIfa.devStatus.GetIndex())
}
