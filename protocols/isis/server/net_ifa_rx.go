package server

import (
	"bytes"
	"fmt"
	"net"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (nifa *netIfa) receiver() {
	for {
		err := nifa.receive()
		if err != nil {
			log.Errorf("IS-IS: receive(): %v", err)
		}
	}
}

func (nifa *netIfa) receive() error {
	pkt, _, err := nifa.ethHandler.RecvPacket()
	if err != nil {
		return errors.Wrap(err, "Read failed")
	}

	return nifa.processPkt(pkt)
}

func (nifa *netIfa) processPkt(rawPkt []byte) error {
	buf := bytes.NewBuffer(rawPkt)
	pkt, err := packet.Decode(buf)
	if err != nil {
		panic(err)
	}

	switch pkt.Header.PDUType {
	case packet.P2P_HELLO:
		return nifa.processP2PHello(pkt)
	case packet.L2_LS_PDU_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 LSPDU")
	case packet.L2_CSNP_TYPE:
		log.WithFields(nifa.fields()).Infof("Received L2 CSNP")
	}

	return fmt.Errorf("Unknown PDU type %d", pkt.Header.PDUType)
}

func (nifa *netIfa) processP2PHello(pkt *packet.ISISPacket) error {
	hello := pkt.Body.(*packet.P2PHello)

	if hello.CircuitType == 1 || hello.CircuitType == 3 {
		if nifa.neighborManagerL1 != nil {
			err := nifa.neighborManagerL1.processP2PHello(hello)
			if err != nil {
				return errors.Wrap(err, "neighbor manager L1 failed processing the p2p hello")
			}
		}
	}

	if hello.CircuitType == 2 || hello.CircuitType == 3 {
		if nifa.neighborManagerL2 != nil {
			err := nifa.neighborManagerL2.processP2PHello(hello)
			if err != nil {
				return errors.Wrap(err, "neighbor manager L1 failed processing the p2p hello")
			}
		}
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
