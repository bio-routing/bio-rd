package server

import (
	"bytes"
	"fmt"
	"syscall"

	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

const (
	ETH_ALEN = 6
)

var (
	ALL_L1_ISS  = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x14}
	ALL_L2_ISS  = [6]byte{0x01, 0x80, 0xC2, 0x00, 0x00, 0x15}
	ALL_P2P_ISS = [6]byte{0x09, 0x00, 0x2b, 0x00, 0x00, 0x5b}
	ISP2PHELLO  = [6]byte{0x09, 0x00, 0x2b, 0x00, 0x00, 0x05}
	ALL_ISS     = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x05}
	ALL_ESS     = [6]byte{0x09, 0x00, 0x2B, 0x00, 0x00, 0x04}
)

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

func (b *bioSys) openPacketSocket() error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("socket() failed: %v", err)
	}
	b.socket = socket

	err = syscall.SetsockoptString(socket, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, string(getISISFilter().serialize()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed (SO_ATTACH_FILTER)")
	}

	/*if syscallwrappers.SetBPFFilter(b.socket) != 0 {
		return fmt.Errorf("Unable to set BPF filter")
	}*/

	err = syscall.BindToDevice(b.socket, b.device.GetName())
	if err != nil {
		return errors.Wrap(err, "Bind failed")
	}
	/*if syscallwrappers.BindToInterface(b.socket, int(b.device.GetIndex())) != 0 {
		return fmt.Errorf("Unable to bind to interface")
	}*/

	return nil
}

func (b *bioSys) closePacketSocket() error {
	return syscall.Close(b.socket)
}

type PacketMreq struct {
	MrIfIndex uint32
	MrType    uint16
	MrAlen    uint16
	MrAddress [8]byte
}

func (p PacketMreq) serialize() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(convert.Uint32Byte(p.MrIfIndex))
	buf.Write(convert.Uint16Byte(p.MrType))
	buf.Write(convert.Uint16Byte(p.MrAlen))
	buf.Write(p.MrAddress[:])
	return buf.Bytes()
}

func (b *bioSys) mcastJoin(addr [ETH_ALEN]byte) error {
	mreq := PacketMreq{
		MrIfIndex: uint32(b.device.GetIndex()),
		MrType:    syscall.PACKET_MR_MULTICAST,
		MrAlen:    ETH_ALEN,
		MrAddress: [8]byte{addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]},
	}

	err := syscall.SetsockoptString(b.socket, syscall.SOL_PACKET, syscall.PACKET_ADD_MEMBERSHIP, string(mreq.serialize()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed")
	}

	/*if syscallwrappers.JoinISISMcast(b.socket, int(b.device.GetIndex())) != 0 {
		return fmt.Errorf("setsockopt failed")
	}*/

	return nil
}

func (b *bioSys) recvPacket() (pkt []byte, src types.MACAddress, err error) {
	buf := make([]byte, 1500)
	nBytes, from, err := syscall.Recvfrom(b.socket, buf, 0)
	if err != nil {
		return nil, types.MACAddress{}, fmt.Errorf("recvfrom failed: %v", err)
	}

	ll := from.(*syscall.SockaddrLinklayer)
	copy(src[:], ll.Addr[:6])

	return buf[:nBytes], src, nil
}

func (b *bioSys) sendPacket(pkt []byte, dst [ETH_ALEN]byte) error {
	ll := syscall.SockaddrLinklayer{
		Ifindex: int(b.device.GetIndex()),
		Halen:   ETH_ALEN,
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
		return fmt.Errorf("sendto failed: %v", err)
	}

	return nil
}
