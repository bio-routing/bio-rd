package ethernet

import (
	"bytes"
	"fmt"
	"net"
	"syscall"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	ethALen         = 6
	ethPAll         = 0x0300
	maxMTU          = 9216
	maxLLCLen       = 0x5ff
	ethertypeExtLLC = 0x8870
)

// MACAddr represens a MAC address
type MACAddr [ethALen]byte

func (m MACAddr) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
}

type EthernetInterface struct {
	socket  int
	devName string
	ifIndex uint32
	llc     LLC
}

type EthernetInterfaceI interface {
	RecvPacket() (pkt []byte, src MACAddr, err error)
	SendPacket(dst MACAddr, pkt []byte) error
	MCastJoin(addr MACAddr) error
	GetMTU() int
	Close()
}

func NewEthernetInterface(devName string, bpf *BPF, llc LLC) (*EthernetInterface, error) {
	ifa, err := net.InterfaceByName(devName)
	if err != nil {
		return nil, fmt.Errorf("net.InterfaceByName failed: %w", err)
	}

	h := &EthernetInterface{
		devName: devName,
		ifIndex: uint32(ifa.Index),
		llc:     llc,
	}

	err = h.init(bpf)
	if err != nil {
		return nil, fmt.Errorf("init failed: %w", err)
	}

	return h, nil
}

func (e *EthernetInterface) init(b *BPF) error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("socket() failed: %v", err)
	}
	e.socket = socket

	err = e.loadBPF(b)
	if err != nil {
		return fmt.Errorf("unable to load BPF: %w", err)
	}

	err = syscall.Bind(e.socket, &syscall.SockaddrLinklayer{
		Protocol: ethPAll,
		Ifindex:  int(e.ifIndex),
	})
	if err != nil {
		return fmt.Errorf("bind failed: %w", err)
	}

	return nil
}

// Close closes the handler
func (e *EthernetInterface) Close() {
	syscall.Close(e.socket)
}

func (e *EthernetInterface) RecvPacket() (pkt []byte, src MACAddr, err error) {
	buf := make([]byte, maxMTU)
	nBytes, from, err := syscall.Recvfrom(e.socket, buf, 0)
	if err != nil {
		return nil, MACAddr{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:ethALen])

	return buf[:nBytes], src, nil
}

// SendPacket sends an 802.3 ethernet packet (LLC)
func (e *EthernetInterface) SendPacket(dst MACAddr, pkt []byte) error {
	pkt = craftLLCPacket(e.llc, pkt)

	err := syscall.Sendto(e.socket, pkt, 0, e.getSockaddrLinklayer(len(pkt), dst))
	if err != nil {
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}

func craftLLCPacket(llc LLC, pkt []byte) []byte {
	buf := bytes.NewBuffer(nil)
	llc.serialize(buf)
	buf.Write(pkt)
	return buf.Bytes()
}

func (e *EthernetInterface) getSockaddrLinklayer(pktLen int, dst MACAddr) *syscall.SockaddrLinklayer {
	sall := &syscall.SockaddrLinklayer{
		Protocol: bnet.Htons(ethertype802dot3(pktLen)),
		Ifindex:  int(e.ifIndex),
		Halen:    ethALen,
	}

	for i := uint8(0); i < sall.Halen; i++ {
		sall.Addr[i] = dst[i]
	}

	return sall
}

func ethertype802dot3(len int) uint16 {
	if len > maxLLCLen {
		return ethertypeExtLLC
	}

	return uint16(len)
}

// GetMTU gets the interfaces MTU
func (e *EthernetInterface) GetMTU() int {
	netIfa, err := net.InterfaceByIndex(int(e.ifIndex))
	if err != nil {
		return -1
	}

	return netIfa.MTU
}
