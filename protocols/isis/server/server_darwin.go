package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

func (n *netIf) openPacketSocket() error {
	return fmt.Errorf("Unsupported platform")
}

func (n *netIf) mcastJoin(addr [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (n *netIf) sendPacket(pkt []byte, dst [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (n *netIf) recvPacket() (pkt []byte, src types.SystemID, err error) {
	return nil, types.SystemID{}, fmt.Errorf("Unsupported platform")
}
