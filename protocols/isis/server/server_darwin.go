package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

func (d *dev) openPacketSocket() error {
	return fmt.Errorf("Unsupported platform")
}

func (d *dev) closePacketSocket() error {
	return fmt.Errorf("Unsupported platform")
}

func (d *dev) mcastJoin(addr [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (d *dev) sendPacket(pkt []byte, dst [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (d *dev) recvPacket() (pkt []byte, src types.MACAddress, err error) {
	return nil, types.MACAddress{}, fmt.Errorf("Unsupported platform")
}
