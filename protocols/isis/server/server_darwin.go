package server

import (
	"fmt"
)

func (n *netIf) openPacketSocket() error {
	return fmt.Errorf("Unsupported platform")
}

func (n *netIf) mcastJoin(addr [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (n *netIf) sendPacket(pkt []byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (n *netIf) recvPacket() (pkt []byte, src [6]byte, err error) {
	return nil, [6]byte{}, fmt.Errorf("Unsupported platform")
}
