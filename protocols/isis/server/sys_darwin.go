package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

func (b *bioSys) openPacketSocket() error {
	return fmt.Errorf("Unsupported platform")
}

func (b *bioSys) closePacketSocket() error {
	return fmt.Errorf("Unsupported platform")
}

func (b *bioSys) mcastJoin(addr [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (b *bioSys) sendPacket(pkt []byte, dst [6]byte) error {
	return fmt.Errorf("Unsupported platform")
}

func (b *bioSys) recvPacket() (pkt []byte, src types.MACAddress, err error) {
	return nil, types.MACAddress{}, fmt.Errorf("Unsupported platform")
}
