package server

import (
	"encoding/binary"
	"fmt"
	"syscall"

	"github.com/bio-routing/bio-rd/biosyscall"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

func (n *netIf) openPacketSocket() error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("socket() failed: %v", err)
	}
	n.socket = socket

	if biosyscall.SetBPFFilter(n.socket) != 0 {
		return fmt.Errorf("Unable to set BPF filter")
	}

	if biosyscall.BindToInterface(n.socket, int(n.device.Index)) != 0 {
		return fmt.Errorf("Unable to bind to interface")
	}

	return nil
}

func (n *netIf) closePacketSocket() error {
	return syscall.Close(n.socket)
}

func (n *netIf) mcastJoin(addr [6]byte) error {
	if biosyscall.JoinISISMcast(n.socket, int(n.device.Index)) != 0 {
		return fmt.Errorf("setsockopt failed")
	}

	return nil
}

func (n *netIf) recvPacket() (pkt []byte, src types.SystemID, err error) {
	buf := make([]byte, 1500)
	nBytes, from, err := syscall.Recvfrom(n.socket, buf, 0)
	if err != nil {
		return nil, types.SystemID{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:6])

	return buf[:nBytes], src, nil
}

func (n *netIf) sendPacket(pkt []byte, dst [6]byte) error {
	ll := syscall.SockaddrLinklayer{
		//Protocol: htons(uint16(len(pkt) + 3)),
		Ifindex: int(n.device.Index),
		Halen:   6, // MAC address length
	}

	for i := uint8(0); i < ll.Halen; i++ {
		ll.Addr[i] = dst[i]
	}

	newPkt := []byte{
		0xfe, 0xfe, 0x03,
	}

	newPkt = append(newPkt, pkt...)

	ll.Protocol = htons(uint16(len(newPkt)))

	err := syscall.Sendto(n.socket, newPkt, 0, &ll)
	if err != nil {
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}

func htons(input uint16) uint16 {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, input)

	return uint16(data[1])*256 + uint16(data[0])
}
