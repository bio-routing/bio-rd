package server

import (
	"fmt"
	"syscall"

	"github.com/bio-routing/bio-rd/biosyscall"
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

	if biosyscall.BindToInterface(n.socket, n.ifa.Index) != 0 {
		return fmt.Errorf("Unable to bind to interface")
	}

	return nil
}

func (n *netIf) mcastJoin(addr [6]byte) error {
	if biosyscall.JoinISISMcast(n.socket, n.ifa.Index) != 0 {
		return fmt.Errorf("setsockopt failed")
	}

	return nil
}

func (n *netIf) recvPacket() (pkt []byte, src [6]byte, err error) {
	buf := make([]byte, 1500)
	n, from, err := syscall.Recvfrom(n.socket, buf, 0)
	if err != nil {
		return nil, [6]byte{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := syscall.SockaddrLinklayer(from)
	copy(src[:], ll.Addr[:5])

	return buf[:n-1], src, nil
}

func (n *netIf) sendPacket(pkt []byte, dst [6]byte) error {
	ll := syscall.SockaddrLinklayer{
		Protocol: 0x00FE,
		Ifindex:  uint16(n.ifa.Index),
		Halen:    6, // MAC address length

	}

	err := syscall.Sendto(n.socket, pkt, 0, ll)
	if err != nil {
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}
