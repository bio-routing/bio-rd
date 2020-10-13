package server

import (
	"fmt"
	"net"
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/net/ethernet"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type neighborManager struct {
	server      *Server
	netIfa      *netIfa
	level       uint8
	neighbors   map[ethernet.MACAddr]*neighbor
	neighborsMu sync.RWMutex
}

func newNeighborManager(server *Server, netIfa *netIfa, level uint8) *neighborManager {
	return &neighborManager{
		server:    server,
		netIfa:    netIfa,
		level:     level,
		neighbors: make(map[ethernet.MACAddr]*neighbor),
	}
}

func (nm *neighborManager) fields() log.Fields {
	return log.Fields{
		"protocol":  "IS-IS",
		"component": "nm",
		"interface": nm.netIfa.name,
		"level":     nm.level,
	}
}

func (nm *neighborManager) getNeighbors() []*neighbor {
	ret := make([]*neighbor, 0)
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	for _, v := range nm.neighbors {
		ret = append(ret, v)
	}

	return ret
}

func (nm *neighborManager) neighborUp(src ethernet.MACAddr) bool {
	nm.neighborsMu.RLock()
	defer nm.neighborsMu.RUnlock()

	if _, found := nm.neighbors[src]; !found {
		return false
	}

	return nm.neighbors[src].getState() == packet.UP_STATE
}

// TODO: Catch if P2P Adj. State is DOWN. What to do then? Drop the neighbor?
func (nm *neighborManager) processP2PHello(src ethernet.MACAddr, hello *packet.P2PHello) error {
	nm.neighborsMu.Lock()
	defer nm.neighborsMu.Unlock()

	err := nm.validateP2PHello(hello)
	if err != nil {
		return errors.Wrap(err, "Invalid p2p hello msg")
	}

	if nm.level == 1 {
		areaAddrsTLV := hello.GetAreaAddressesTLV()
		if areaAddrsTLV == nil {
			return fmt.Errorf("Area Addresses TLV not found")
		}
		if !nm.netIfa.validateAreasL1(areaAddrsTLV.AreaIDs) {
			log.WithFields(nm.fields()).Debug("Rejecting L1 adjacency due to area mismatch")
			return nil
		}
	}

	if _, found := nm.neighbors[src]; !found {
		n := nm.neighborFromP2PHello(hello)
		nm.neighbors[src] = n

		n.wg.Add(1)
		go n.adjChecker()

		log.WithFields(nm.fields()).Infof("Adding new neighbor %q", hello.SystemID.String())
		return nil
	}

	n := nm.neighbors[src]
	return n.processP2PHello(hello)
}

// validateP2PHello validates p2p hello messages
func (nm *neighborManager) validateP2PHello(hello *packet.P2PHello) error {
	areaAddrsTLV := hello.GetAreaAddressesTLV()
	if areaAddrsTLV == nil {
		return fmt.Errorf("Area Addresses TLV missing")
	}

	if len(areaAddrsTLV.AreaIDs) == 0 {
		return fmt.Errorf("No area(s) given in Area Addresses TLV")
	}

	p2pAdjTLV := hello.GetP2PAdjTLV()
	if p2pAdjTLV == nil {
		return fmt.Errorf("P2P Adjacency TLV missing")
	}

	protoSupportTLV := hello.GetProtocolsSupportedTLV()
	if protoSupportTLV == nil {
		return fmt.Errorf("Protocol Supported TLV missing")
	}

	if !validateProtocolsSupported(protoSupportTLV.NetworkLayerProtocolIDs) {
		return fmt.Errorf("Protocol supported mismatch (IPv4 + IPv6 required)")
	}

	ipAddrsTLV := hello.GetIPInterfaceAddressesesTLV()
	if ipAddrsTLV == nil {
		return fmt.Errorf("IP Interface Addresses TLV missing")
	}

	if !nm.netIfa.validateIPv4Addresses(ipAddrsTLV.IPv4Addresses) {
		return fmt.Errorf("IPv4 addressing mismatch")
	}

	return nil
}

// validateAreasL1 checks if any of the received areas matches with a localy configured area
func (nifa *netIfa) validateAreasL1(receivedAreas []types.AreaID) bool {
	localAreas := make([]types.AreaID, 0)
	for _, net := range nifa.srv.nets {
		localAreas = append(localAreas, append([]byte{net.AFI}, net.AreaID...))
	}

	for _, needle := range receivedAreas {
		for _, local := range localAreas {
			if needle.Equal(local) {
				return true
			}
		}
	}

	return false
}

func validateProtocolsSupported(protocols []uint8) bool {
	needed := []byte{
		packet.NLPIDIPv4,
		packet.NLPIDIPv6,
	}

	for _, needle := range needed {
		found := false
		for _, p := range protocols {
			if p == needle {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (nifa *netIfa) validateIPv4Addresses(addrs []uint32) bool {
	localAddrs := nifa.devStatus.GetAddrs()

	for _, a := range addrs {
		b := bnet.IPv4(a)
		for _, localAddr := range localAddrs {
			if localAddr.Contains(bnet.NewPfx(b, net.IPv4len*8).Ptr()) {
				return true
			}
		}
	}

	return false
}
