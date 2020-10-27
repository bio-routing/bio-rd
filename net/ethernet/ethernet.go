package ethernet

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/pkg/errors"
)

const (
	ethALen         = 6
	ethPAll         = 0x0300
	maxMTU          = 9216
	maxLLCLen       = 0x5ff
	ethertypeExtLLC = 0x8870
)

var (
	wordWidth  uint8
	wordLength uintptr
)

func init() {
	wordWidth = uint8(unsafe.Sizeof(int(0)))
	wordLength = unsafe.Sizeof(uintptr(0))
}

// MACAddr represens a MAC address
type MACAddr [ethALen]byte

func (m MACAddr) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
}

// Handler is an Ethernet handler
type Handler struct {
	socket  int
	devName string
	ifIndex uint32
}

// HandlerInterface is an handler interface
type HandlerInterface interface {
	NewConn(dest MACAddr) net.Conn
	RecvPacket() (pkt []byte, src MACAddr, err error)
	SendPacket(pkt []byte, dst MACAddr) error
	MCastJoin(addr MACAddr) error
	GetMTU() int
}

// NewHandler creates a new Ethernet handler
func NewHandler(devName string, bpf *BPF) (*Handler, error) {
	ifa, err := net.InterfaceByName(devName)
	if err != nil {
		return nil, errors.Wrapf(err, "net.InterfaceByName failed")
	}

	h := &Handler{
		devName: devName,
		ifIndex: uint32(ifa.Index),
	}

	err = h.init(bpf)
	if err != nil {
		return nil, errors.Wrap(err, "init failed")
	}

	return h, nil
}

func (e *Handler) init(b *BPF) error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("socket() failed: %v", err)
	}
	e.socket = socket

	err = e.loadBPF(b)
	if err != nil {
		return errors.Wrap(err, "Unable to load BPF")
	}

	err = syscall.Bind(e.socket, &syscall.SockaddrLinklayer{
		Protocol: ethPAll,
		Ifindex:  int(e.ifIndex),
	})
	if err != nil {
		return errors.Wrap(err, "Bind failed")
	}

	return nil
}

func (e *Handler) closePacketSocket() error {
	return syscall.Close(e.socket)
}

// RecvPacket receives a packet on the ethernet handler
func (e *Handler) RecvPacket() (pkt []byte, src MACAddr, err error) {
	buf := make([]byte, maxMTU)
	nBytes, from, err := syscall.Recvfrom(e.socket, buf, 0)
	if err != nil {
		return nil, MACAddr{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:ethALen])

	return buf[:nBytes], src, nil
}

// SendPacket sends a packet
func (e *Handler) SendPacket(pkt []byte, dst MACAddr) error {
	newPkt := []byte{
		0xfe, 0xfe, 0x03, // LLC
	}
	newPkt = append(newPkt, pkt...)

	sall := &syscall.SockaddrLinklayer{
		Protocol: bnet.Htons(uint16(ethertype802dot3(len(newPkt)))),
		Ifindex:  int(e.ifIndex),
		Halen:    ethALen,
	}

	for i := uint8(0); i < sall.Halen; i++ {
		sall.Addr[i] = dst[i]
	}

	err := syscall.Sendto(e.socket, newPkt, 0, sall)
	if err != nil {
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}

func ethertype802dot3(len int) int {
	if len > maxLLCLen {
		return ethertypeExtLLC
	}

	return len
}

// GetMTU gets the interfaces MTU
func (e *Handler) GetMTU() int {
	netIfa, err := net.InterfaceByIndex(int(e.ifIndex))
	if err != nil {
		return -1
	}

	return netIfa.MTU
}
