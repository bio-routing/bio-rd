package ethernet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

const (
	// EthALen is the length of an ethernet address
	EthALen = 6

	// ETH_P_ALL real value
	ETH_P_ALL = 0x0300

	maxMTU = 9216
)

var (
	// AllL1ISs is All Level 1 Intermediate Systems
	AllL1ISs = [EthALen]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x14}

	// AllL2ISs is All Level 2 Intermediate Systems
	AllL2ISs = [EthALen]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x15}

	// AllP2PISs is All Point-to-Point Intermediate Systems
	AllP2PISs = [EthALen]byte{0x09, 0x00, 0x2b, 0x00, 0x00, 0x5b}

	// ISp2pHello is Intermediate System Point-to-Point Hello
	ISp2pHello = [EthALen]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}

	AllISS = [EthALen]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}
	AllESS = [EthALen]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x04}
)

// MACAddr represens a MAC address
type MACAddr [6]byte

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
	NewConn(dest [EthALen]byte) net.Conn
	RecvPacket() (pkt []byte, src types.MACAddress, err error)
	MCastJoin(addr MACAddr) error
}

// NewHandler creates a new Ethernet handler
func NewHandler(devName string) (*Handler, error) {
	ifa, err := net.InterfaceByName(devName)
	if err != nil {
		return nil, errors.Wrapf(err, "net.InterfaceByName failed")
	}

	h := &Handler{
		devName: devName,
		ifIndex: uint32(ifa.Index),
	}

	err = h.init()
	if err != nil {
		return nil, errors.Wrap(err, "init failed")
	}

	return h, nil
}

type sockFprog struct {
	len     uint16
	filters []sockFilter
}

type sockFilter struct {
	code uint16
	jt   uint8
	jf   uint8
	k    uint32
}

func (s sockFprog) serializeTerms() [48]byte {
	directives := bytes.NewBuffer(nil)
	for _, sf := range s.filters {
		directives.Write(convert.Reverse(convert.Uint16Byte(sf.code)))
		directives.WriteByte(sf.jt)
		directives.WriteByte(sf.jf)
		directives.Write(convert.Reverse(convert.Uint32Byte(sf.k)))
	}

	ret := [48]byte{}
	copy(ret[:], directives.Bytes())
	return ret
}

func getISISFilter() sockFprog {
	return sockFprog{
		filters: []sockFilter{
			{
				code: 0x28,
				k:    0x0000000e - 14,
			},
			{
				code: 0x15,
				jf:   3,
				k:    0x0000fefe,
			},
			{
				code: 0x30,
				k:    0x00000011 - 14,
			},
			{
				code: 0x15,
				jf:   1,
				k:    0x00000083,
			},
			{
				code: 0x6,
				k:    0x00040000,
			},
			{
				code: 0x6,
			},
		},
	}
}

func (e *Handler) init() error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("socket() failed: %v", err)
	}
	e.socket = socket

	f := getISISFilter()
	terms := f.serializeTerms()
	buf := bytes.NewBuffer(nil)
	buf.Write(convert.Reverse(convert.Uint16Byte(6)))

	// Align to next 8 byte word
	for i := 0; i < 6; i++ {
		buf.WriteByte(0)
	}

	p := unsafe.Pointer(&terms)
	buf.Write(convert.Reverse(convert.Uint64Byte(uint64(uintptr(p)))))
	err = syscall.SetsockoptString(socket, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, string(buf.Bytes()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed (SO_ATTACH_FILTER)")
	}

	err = syscall.Bind(e.socket, &syscall.SockaddrLinklayer{
		Protocol: ETH_P_ALL,
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

type packetMreq struct {
	mrIfIndex uint32
	mrType    uint16
	mrAlen    uint16
	mrAddress [8]byte
}

func (p packetMreq) serialize() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(convert.Reverse(convert.Uint32Byte(p.mrIfIndex)))
	buf.Write(convert.Reverse(convert.Uint16Byte(p.mrType)))
	buf.Write(convert.Reverse(convert.Uint16Byte(p.mrAlen)))
	buf.Write(p.mrAddress[:])
	return buf.Bytes()
}

// MCastJoin joins a multicast group
func (e *Handler) MCastJoin(addr MACAddr) error {
	mreq := packetMreq{
		mrIfIndex: uint32(e.ifIndex),
		mrType:    syscall.PACKET_MR_MULTICAST,
		mrAlen:    EthALen,
		mrAddress: [8]byte{addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]},
	}

	err := syscall.SetsockoptString(e.socket, syscall.SOL_PACKET, syscall.PACKET_ADD_MEMBERSHIP, string(mreq.serialize()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed")
	}

	return nil
}

// RecvPacket receives a packet on the ethernet handler
func (e *Handler) RecvPacket() (pkt []byte, src types.MACAddress, err error) {
	buf := make([]byte, maxMTU)
	nBytes, from, err := syscall.Recvfrom(e.socket, buf, 0)
	panic("RECV!")
	if err != nil {
		return nil, types.MACAddress{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:EthALen])

	return buf[:nBytes], src, nil
}

func (e *Handler) sendPacket(pkt []byte, dst [EthALen]byte) error {
	fmt.Printf("Sending packet: %v\n", pkt)
	ll := syscall.SockaddrLinklayer{
		Ifindex: int(e.ifIndex),
		Halen:   EthALen,
	}

	for i := uint8(0); i < ll.Halen; i++ {
		ll.Addr[i] = dst[i]
	}

	newPkt := []byte{
		0xfe, 0xfe, 0x03,
	}

	newPkt = append(newPkt, pkt...)

	length := []byte{0, 0}
	binary.BigEndian.PutUint16(length, uint16(len(newPkt)))
	ll.Protocol = convert.Uint16(length)

	err := syscall.Sendto(e.socket, newPkt, 0, &ll)
	if err != nil {
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}
