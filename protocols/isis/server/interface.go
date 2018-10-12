package server

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"

	log "github.com/sirupsen/logrus"
)

var (
	AllL1ISS  = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x14}
	AllL2ISS  = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x15}
	AllP2PISS = [6]byte{0x09, 0x00, 0x2b, 0x00, 0x00, 0x05}
	AllISS    = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}
	AllESS    = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x04}
)

const (
	NLPID_IPv4 = uint8(0xcc)
	NLPID_IPv6 = uint8(0x8e)
)

type netIf struct {
	isisServer         *ISISServer
	name               string
	ifa                *net.Interface
	passive            bool
	p2p                bool
	l1                 *level
	l2                 *level
	socket             int
	supportedProtocols []uint8
	stop               chan struct{}
}

type level struct {
	HelloInterval uint16
	HoldTime      uint16
	Metric        uint32
	Priority      uint8
	neighbors     map[types.SystemID]*neighbor
	neighborsMu   sync.RWMutex
}

func newNetIf(srv *ISISServer, c config.ISISInterfaceConfig) (*netIf, error) {
	nif := netIf{
		isisServer:         srv,
		passive:            c.Passive,
		p2p:                c.P2P,
		supportedProtocols: []uint8{NLPID_IPv4, NLPID_IPv6},
		stop:               make(chan struct{}),
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

func (ifa *netIf) compareSupportedProtocols(protocols []uint8) bool {
	if len(ifa.supportedProtocols) != len(protocols) {
		return false
	}

	for _, p := range protocols {
		found := false
		for _, q := range ifa.supportedProtocols {
			if p == q {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (ifa *netIf) startReceiver() {
	go func(ifa *netIf) {
		for {
			select {
			case <-ifa.stop:
				return
			default:
				rawPkt, src, err := ifa.recvPacket()
				if err != nil {
					log.Errorf("recvPacket() failed: %v", err)
					return
				}

				ifa.processIngressPacket(rawPkt, src)
			}
		}
	}(n)
}

func (ifa *netIf) processIngressPacket(rawPkt []byte, src types.SystemID) {
	pkt, err := packet.Decode(bytes.NewBuffer(rawPkt))
	if err != nil {
		log.Errorf("Unable to decode packet from %v: %v: %v", src, rawPkt, err)
		return
	}

	switch pkt.Header.PDUType {
	case packet.P2P_HELLO:
		log.Infof("Received P2P hello: %v", rawPkt)
		ifa.processIngressP2PHello(pkt)
	case packet.L1_LAN_HELLO_TYPE:
		// TODO: Implement LAN support for L1
		log.Errorf("L1 LAN support is not implemented yet")
	case packet.L2_LAN_HELLO_TYPE:
		// TODO: Implement LAN support for L2
		log.Errorf("L2 LAN support is not implemented yet")
	default:

		log.Errorf("Unknown packet received from %v: %v", src, rawPkt)
	}
}

func (ifa *netIf) processIngressP2PHello(pkt *packet.ISISPacket) {
	hello := pkt.Body.(*packet.P2PHello)

	for _, tlv := range hello.TLVs {
		fmt.Printf("TLV Type: %d\n", tlv.Type())
	}

	switch hello.CircuitType {
	case 1:
		// TODO: Implement P2P L1 support
		return
	case 2:
		ifa.l2.neighborsMu.RLock()
		if _, ok := ifa.l2.neighbors[hello.SystemID]; !ok {
			neighbor := &neighbor{
				systemID:       hello.SystemID,
				ifa:            n,
				holdingTime:    hello.HoldingTimer,
				localCircuitID: hello.LocalCircuitID,
			}
			ifa.l2.neighborsMu.RUnlock()
			ifa.l2.neighborsMu.Lock()
			ifa.l2.neighbors[hello.SystemID] = neighbor
			ifa.l2.neighborsMu.Unlock()

			ifa.l2.neighborsMu.RLock()

			fsm := newFSM(ifa.l2.neighbors[hello.SystemID])
			ifa.l2.neighbors[hello.SystemID].fsm = fsm
			go fsm.run()

			return
		}

		neighbor := ifa.l2.neighbors[hello.SystemID]
		ifa.l2.neighborsMu.RUnlock()

		neighbor.fsm.receive(pkt)
	case 3:
		// TODO
	}
}

func (ifa *netIf) helloSender() {
	ifa.p2pHelloSender()
}

func (ifa *netIf) p2pHelloSender() {
	t := time.NewTicker(time.Duration(ifa.l2.HelloInterval) * time.Second)
	for {
		select {
		case <-t.C:
			err := ifa.sendP2PHello()
			if err != nil {
				log.Errorf("Unable to send hello packet: %v", err)
			}
		case <-ifa.stop:
			return
		}
	}
}

func (ifa *netIf) sendP2PHello() error {
	p := packet.P2PHello{
		CircuitType:  packet.L2CircuitType,
		SystemID:     ifa.isisServer.systemID(),
		HoldingTimer: ifa.l2.HoldTime,
		PDULength:    packet.P2PHelloMinSize,
		TLVs:         ifa.p2pHelloTLVs(),
	}

	for _, TLV := range p.TLVs {
		p.PDULength += 2
		p.PDULength += uint16(TLV.Length())
	}

	helloBuf := bytes.NewBuffer(nil)
	p.Serialize(helloBuf)

	hdr := packet.ISISHeader{
		ProtoDiscriminator:  0x83,
		LengthIndicator:     20,
		ProtocolIDExtension: 1,
		IDLength:            0,
		PDUType:             packet.P2P_HELLO,
		Version:             1,
		MaxAreaAddresses:    0,
	}

	hdrBuf := bytes.NewBuffer(nil)
	hdr.Serialize(hdrBuf)

	hdrBuf.Write(helloBuf.Bytes())

	fmt.Printf("Sending Hello: %v\n", hdrBuf.Bytes())

	err := ifa.sendPacket(hdrBuf.Bytes(), AllISS)
	if err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

func (ifa *netIf) p2pHelloTLVs() []packet.TLV {

	l2AdjacencyState, neighborSystemID, neighborExtendedLocalCircuitID := ifa.p2pL2AdjacencyState()
	p2pAdjStateTLV := packet.NewP2PAdjacencyStateTLV(l2AdjacencyState, 1234)

	switch l2AdjacencyState {
	case packet.INITIALIZING_STATE:
		p2pAdjStateTLV.TLVLength = 15
		p2pAdjStateTLV.NeighborSystemID = neighborSystemID
		p2pAdjStateTLV.NeighborExtendedLocalCircuitID = neighborExtendedLocalCircuitID
	}

	protocolsSupportedTLV := packet.NewProtocolsSupportedTLV(ifa.supportedProtocols)
	areaAddressesTLV := packet.NewAreaAddressesTLV(ifa.getAreas())

	ipInterfaceAddressesTLV := packet.NewIPInterfaceAddressTLV(3232236033) //FIXME: Insert address automatically

	return []packet.TLV{
		p2pAdjStateTLV,
		protocolsSupportedTLV,
		areaAddressesTLV,
		ipInterfaceAddressesTLV,
	}
}

func (ifa *netIf) getAreas() []types.AreaID {
	areas := make([]types.AreaID, len(ifa.isisServer.config.NETs))
	for i, NET := range ifa.isisServer.config.NETs {
		a := []byte{NET.AFI}
		a = append(a, NET.AreaID...)
		areas[i] = a
	}

	return areas
}

func (ifa *netIf) p2pL2AdjacencyState() (state uint8, neighbor types.SystemID, neighborExtendedLocalCircuitID uint32) {
	ifa.l2.neighborsMu.RLock()
	defer ifa.l2.neighborsMu.RUnlock()

	if len(ifa.l2.neighbors) == 0 {
		return packet.DOWN_STATE, types.SystemID{}, 0
	}

	for systemID, neighbor := range ifa.l2.neighbors {
		return neighbor.fsm.state.getState(), systemID, neighbor.extendedLocalCircuitID
	}

	panic("This is impossible: Length of map is != 0 while map is empty")
}
