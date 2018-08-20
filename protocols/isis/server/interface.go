package server

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type netIf struct {
	isisServer *ISISServer
	passive    bool
	p2p        bool
	l1         *level
	l2         *level
}

type level struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	Priority      uint8
	neighbors     map[types.SystemID]*neighbor
}

func newNetIf(srv *ISISServer, c config.ISISInterfaceConfig) *netIf {
	nif := netIf{
		isisServer: srv,
		passive:    c.Passive,
		p2p:        c.P2P,
	}

	if c.ISISLevel1Config != nil {
		nif.l1.HelloInterval = c.ISISLevel1Config.HelloInterval
		nif.l1.HoldTime = c.ISISLevel1Config.HoldTime
		nif.l1.Metric = c.ISISLevel1Config.Metric
		nif.l1.Priority = c.ISISLevel1Config.Priority
		nif.l1.neighbors = make(map[types.SystemID]*neighbor)
	}

	if c.ISISLevel2Config != nil {
		nif.l2.HelloInterval = c.ISISLevel2Config.HelloInterval
		nif.l2.HoldTime = c.ISISLevel2Config.HoldTime
		nif.l2.Metric = c.ISISLevel2Config.Metric
		nif.l2.Priority = c.ISISLevel2Config.Priority
		nif.l2.neighbors = make(map[types.SystemID]*neighbor)
	}

	return &nif
}

func (n *netIf) helloSender() {
	t := time.NewTicker(time.Duration(n.l2.HelloInterval) * time.Second)
	for {
		<-t.C
		n.sendHello()
	}
}

func (n *netIf) sendHello() {
	p := packet.L2Hello{
		CircuitType:  packet.L2CircuitType,
		SystemID:     n.isisServer.systemID(),
		HoldingTimer: n.l2.HoldTime,
		Priority:     128,
	}

	buf := bytes.NewBuffer(nil)
	p.Serialize(buf)

	fmt.Printf("Sending Hello: %v\n", buf.Bytes())
}
