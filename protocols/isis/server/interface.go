package server

import (
	"syscall"
	"bytes"
	"fmt"
	"time"
	"net"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/biosyscall"
)

var (
	AllL1ISS = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x14}
	AllL2ISS = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x15}
	AllP2PISS = [6]byte{0x09, 0x00, 0x2b, 0x00, 0x00, 0x5b}
	AllISS = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}
	AllESS = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x04}
	// tcpdump -n -i XX isis -dd
	filter = []syscall.SockFilter{
		{ 0x28, 0, 0, 0x0000000c },
		{ 0x25, 5, 0, 0x000005dc },
		{ 0x28, 0, 0, 0x0000000e },
		{ 0x15, 0, 3, 0x0000fefe },
		{ 0x30, 0, 0, 0x00000011 },
		{ 0x15, 0, 1, 0x00000083 },
		{ 0x6, 0, 0, 0x00040000 },
		{ 0x6, 0, 0, 0x00000000 },
	}
)

type netIf struct {
	isisServer *ISISServer
	name       string
	ifa        *net.Interface
	passive    bool
	p2p        bool
	l1         *level
	l2         *level
	socket     int
}

type level struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	Priority      uint8
	neighbors     map[types.SystemID]*neighbor
}

func newNetIf(srv *ISISServer, c config.ISISInterfaceConfig) (*netIf, error) {
	nif := netIf{
		isisServer: srv,
		passive:    c.Passive,
		p2p:        c.P2P,
	}

	if c.ISISLevel1Config != nil {
		nif.l1 = &level{}
		nif.l1.HelloInterval = c.ISISLevel1Config.HelloInterval
		nif.l1.HoldTime = c.ISISLevel1Config.HoldTime
		nif.l1.Metric = c.ISISLevel1Config.Metric
		nif.l1.Priority = c.ISISLevel1Config.Priority
		nif.l1.neighbors = make(map[types.SystemID]*neighbor)
	}

	if c.ISISLevel2Config != nil {
		nif.l2 = &level{}
		nif.l2.HelloInterval = c.ISISLevel2Config.HelloInterval
		nif.l2.HoldTime = c.ISISLevel2Config.HoldTime
		nif.l2.Metric = c.ISISLevel2Config.Metric
		nif.l2.Priority = c.ISISLevel2Config.Priority
		nif.l2.neighbors = make(map[types.SystemID]*neighbor)
	}

	ifa, err := net.InterfaceByName(c.Name)
	if err != nil {
		return nil, fmt.Errorf("Unable to get interface %q: %v", c.Name, err)
	}
	nif.ifa = ifa

	err = nif.openPacketSocket()
	if err != nil {
		return nil, fmt.Errorf("Failed to open packet socket: %v", err)
	}

	err = nif.mcastJoin(AllP2PISS)
	if err != nil {
		return nil, fmt.Errorf("Failed to join multicast group: %v", err)
	}

	return &nif, nil
}

func (n *netIf) mcastJoin(addr [6]byte) error {
	if biosyscall.JoinISISMcast(n.socket, n.ifa.Index) != 0 {
		return fmt.Errorf("setsockopt failed")
	}

	return nil
}

func (n *netIf) openPacketSocket() error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("socket() failed: %v", err)
	}
	n.socket = socket

	if biosyscall.SetBPFFilter(n.socket) != 0 {
		return fmt.Errorf("Unable to set BPF filter")
	}

	if biosyscall.BindToInterface(n.socket, n.ifa.Index) != 0 {
		return fmt.Errorf("Unable to bind to interface")
	}

	return nil
}

func (n *netIf) readPacket() {
	buffer := make([]byte, 1500)
	fmt.Printf("Waiting for packet...\n")
	_, err := syscall.Read(n.socket, buffer)
	if err != nil {
		fmt.Printf("read() failed\n")
		return
	}

	fmt.Printf("Recevied: %v\n", buffer)
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
