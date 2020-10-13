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
)

const (
	// EthALen is the length of an ethernet address
	EthALen = 6

	// ETH_P_ALL real value
	ETH_P_ALL = 0x0300

	maxMTU = 9216

	MAX_LLC_LEN       = 0x5ff
	ETHERTYPE_EXT_LLC = 0x8870
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

	localEndianness binary.ByteOrder
	wordWidth       uint8
	wordLength      uintptr
)

func init() {
	wordWidth = uint8(unsafe.Sizeof(int(0)))
	wordLength = unsafe.Sizeof(uintptr(0))

	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		localEndianness = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		localEndianness = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

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
	NewConn(dest MACAddr) net.Conn
	RecvPacket() (pkt []byte, src MACAddr, err error)
	SendPacket(pkt []byte, dst MACAddr) error
	MCastJoin(addr MACAddr) error
	GetMTU() int
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
		directives.Write(bigEndianToLocal(convert.Uint16Byte(sf.code)))
		directives.WriteByte(sf.jt)
		directives.WriteByte(sf.jf)
		directives.Write(bigEndianToLocal(convert.Uint32Byte(sf.k)))
	}

	ret := [48]byte{}
	copy(ret[:], directives.Bytes())
	return ret
}

func getISISFilter() sockFprog {
	// { 0x6, 0, 0, 0x00040000 },
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

	bpfProgTermCount := len(f.filters)
	buf.Write(bigEndianToLocal(convert.Uint16Byte(uint16(bpfProgTermCount))))

	// Align to next word
	for i := 0; i < int(wordLength)-int(unsafe.Sizeof(uint16(0))); i++ {
		buf.WriteByte(0)
	}

	p := unsafe.Pointer(&terms)
	switch wordWidth {
	case 4:
		buf.Write(bigEndianToLocal(convert.Uint32Byte(uint32(uintptr(p)))))
	case 8:
		buf.Write(bigEndianToLocal(convert.Uint64Byte(uint64(uintptr(p)))))
	default:
		panic("Unknown word width")
	}

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
	buf.Write(bigEndianToLocal(convert.Uint32Byte(p.mrIfIndex)))
	buf.Write(bigEndianToLocal(convert.Uint16Byte(p.mrType)))
	buf.Write(bigEndianToLocal(convert.Uint16Byte(p.mrAlen)))
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
func (e *Handler) RecvPacket() (pkt []byte, src MACAddr, err error) {
	buf := make([]byte, maxMTU)
	nBytes, from, err := syscall.Recvfrom(e.socket, buf, 0)
	if err != nil {
		return nil, MACAddr{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:EthALen])

	return buf[:nBytes], src, nil
}

// SendPacket sends a packet
func (e *Handler) SendPacket(pkt []byte, dst MACAddr) error {
	newPkt := []byte{
		0xfe, 0xfe, 0x03, // LLC
	}
	newPkt = append(newPkt, pkt...)

	sall := &syscall.SockaddrLinklayer{
		Protocol: htons(uint16(isisEtherType(len(newPkt)))),
		Ifindex:  int(e.ifIndex),
		Halen:    EthALen,
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

func bigEndianToLocal(input []byte) []byte {
	if localEndianness == binary.BigEndian {
		return input
	}

	return convert.Reverse(input)
}

func isisEtherType(len int) int {
	if len > MAX_LLC_LEN {
		return ETHERTYPE_EXT_LLC
	}

	return len
}

func htons(x uint16) uint16 {
	if localEndianness == binary.BigEndian {
		return x
	}

	xp := unsafe.Pointer(&x)
	b := (*[2]byte)(xp)

	tmp := b[0]
	b[0] = b[1]
	b[1] = tmp

	return *(*uint16)(xp)
}

// GetMTU gets the interfaces MTU
func (e *Handler) GetMTU() int {
	netIfa, err := net.InterfaceByIndex(int(e.ifIndex))
	if err != nil {
		return -1
	}

	return netIfa.MTU
}
