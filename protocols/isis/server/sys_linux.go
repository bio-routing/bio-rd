package server

import (
	"fmt"
	"syscall"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/syscallwrappers"
	"github.com/pkg/errors"
)

func (b *bioSys) openPacketSocket() error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return errors.Wrap(err, "socket() failed")
	}
	b.socket = socket

	if syscallwrappers.SetBPFFilter(b.socket) != 0 {
		return fmt.Errorf("Unable to set BPF filter")
	}

	if syscallwrappers.BindToInterface(b.socket, int(b.device.Index)) != 0 {
		return fmt.Errorf("Unable to bind to interface")
	}

	return nil
}

func (b *bioSys) closePacketSocket() error {
	return syscall.Close(b.socket)
}

func (b *bioSys) mcastJoin(addr [6]byte) error {
	if syscallwrappers.JoinISISMcast(b.socket, int(b.device.Index)) != 0 {
		return fmt.Errorf("setsockopt failed")
	}

	return nil
}

func (b *bioSys) recvPacket() (pkt []byte, src types.MACAddress, err error) {
	buf := make([]byte, 1500)
	nBytes, from, err := syscall.Recvfrom(b.socket, buf, 0)
	if err != nil {
		return nil, types.MACAddress{}, errors.Wrap(err, "recvfrom failed")
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:6])

	return buf[:nBytes], src, nil
}

func (b *bioSys) sendPacket(pkt []byte, dst [6]byte) error {
	ll := syscall.SockaddrLinklayer{
		Ifindex: int(b.device.Index),
		Halen:   6, // MAC address length
	}

	for i := uint8(0); i < ll.Halen; i++ {
		ll.Addr[i] = dst[i]
	}

	newPkt := []byte{
		0xfe, 0xfe, 0x03,
	}

	newPkt = append(newPkt, pkt...)

	ll.Protocol = uint16(len(newPkt))

	err := syscall.Sendto(b.socket, newPkt, 0, &ll)
	if err != nil {
		return errors.Wrap(err, "sendto failed")
	}

	return nil
}
