package ethernet

import (
	"bytes"
	"syscall"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

type packetMreq struct {
	mrIfIndex uint32
	mrType    uint16
	mrAlen    uint16
	mrAddress [8]byte
}

func (p packetMreq) serialize() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(bnet.BigEndianToLocal(convert.Uint32Byte(p.mrIfIndex)))
	buf.Write(bnet.BigEndianToLocal(convert.Uint16Byte(p.mrType)))
	buf.Write(bnet.BigEndianToLocal(convert.Uint16Byte(p.mrAlen)))
	buf.Write(p.mrAddress[:])
	return buf.Bytes()
}

// MCastJoin joins a multicast group
func (e *Handler) MCastJoin(addr MACAddr) error {
	mreq := packetMreq{
		mrIfIndex: uint32(e.ifIndex),
		mrType:    syscall.PACKET_MR_MULTICAST,
		mrAlen:    ethALen,
		mrAddress: [8]byte{addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]},
	}

	err := syscall.SetsockoptString(e.socket, syscall.SOL_PACKET, syscall.PACKET_ADD_MEMBERSHIP, string(mreq.serialize()))
	if err != nil {
		return errors.Wrap(err, "Setsockopt failed")
	}

	return nil
}
