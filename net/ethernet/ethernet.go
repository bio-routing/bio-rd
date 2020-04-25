package ethernet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"

	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

const (
	// EthALen is the length of an ethernet address
	EthALen = 6
)

var (
	// AllL1ISs is All Level 1 Intermediate Systems
	AllL1ISs = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x14}

	// AllL2ISs is All Level 2 Intermediate Systems
	AllL2ISs = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x15}

	// AllP2PISs is All Point-to-Point Intermediate Systems
	AllP2PISs = [6]byte{0x09, 0x00, 0x2b, 0x00, 0x00, 0x5b}

	// ISp2pHello is Intermediate System Point-to-Point Hello
	ISp2pHello = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}

	AllISS = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}
	AllESS = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x04}
)

// Handler is an Ethernet handler
type Handler struct {
	socket  int
	devName string
	ifIndex uint32
}

// HandlerInterface is an handler interface
type HandlerInterface interface {
	NewConn(dest [EthALen]byte) net.Conn
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
	fj   uint8
	k    uint32
}

func (s sockFprog) serialize() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(convert.Uint16Byte(uint16(len(s.filters) + 2)))
	for _, sf := range s.filters {
		buf.Write(convert.Uint16Byte(sf.code))
		buf.WriteByte(sf.jt)
		buf.WriteByte(sf.fj)
		buf.Write(convert.Uint32Byte(sf.k))
	}

	return buf.Bytes()
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
				fj:   3,
				k:    0x0000fefe,
			},
			{
				code: 0x30,
				k:    0x00000011 - 14,
			},
			{
				code: 0x15,
				fj:   1,
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

	/*err = syscall.SetsockoptString(socket, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, string(getISISFilter().serialize()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed (SO_ATTACH_FILTER)")
	}*/

	err = syscall.BindToDevice(e.socket, e.devName)
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
	buf.Write(convert.Uint32Byte(p.mrIfIndex))
	buf.Write(convert.Uint16Byte(p.mrType))
	buf.Write(convert.Uint16Byte(p.mrAlen))
	buf.Write(p.mrAddress[:])
	return buf.Bytes()
}

func (e *Handler) mcastJoin(addr [EthALen]byte) error {
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

func (e *Handler) recvPacket() (pkt []byte, src types.MACAddress, err error) {
	buf := make([]byte, 1500)
	nBytes, from, err := syscall.Recvfrom(e.socket, buf, 0)
	if err != nil {
		return nil, types.MACAddress{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:6])

	return buf[:nBytes], src, nil
}

func (e *Handler) sendPacket(pkt []byte, dst [EthALen]byte) error {
	fmt.Printf("Sending packet: %v\n", pkt)
	fmt.Printf("Pkt len: %d\n", len(pkt))
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
	fmt.Printf("len(newPkt: %d\n", len(newPkt))

	length := []byte{0, 0}
	binary.BigEndian.PutUint16(length, uint16(len(newPkt)))
	ll.Protocol = convert.Uint16(length)

	err := syscall.Sendto(e.socket, newPkt, 0, &ll)
	if err != nil {
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}
