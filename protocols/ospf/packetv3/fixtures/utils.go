package fixtures

import (
	"net"
	"os"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func PacketReader(t *testing.T, path string) (*pcapgo.Reader, *os.File) {
	f, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}

	r, err := pcapgo.NewReader(f)
	if err != nil {
		t.Error(err)
	}
	return r, f
}

func Payload(raw []byte) (pl []byte, src, dst net.IP, err error) {
	packet := gopacket.NewPacket(raw, layers.LayerTypeEthernet, gopacket.Default)
	if perr := packet.ErrorLayer(); perr != nil {
		// fallback to handling of FrameRelay (cut-off header)
		packet = gopacket.NewPacket(raw[4:], layers.LayerTypeIPv6, gopacket.Default)
		if perr = packet.ErrorLayer(); perr != nil {
			err = perr.Error()
			return
		}
	}

	flowSrc, flowDst := packet.NetworkLayer().NetworkFlow().Endpoints()
	src = net.IP(flowSrc.Raw())
	dst = net.IP(flowDst.Raw())

	pl = packet.NetworkLayer().LayerPayload()
	return
}
