package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decoder"
)

// RouteMonitoringMsg represents a Route Monitoring Message
type RouteMonitoringMsg struct {
	CommonHeader  *CommonHeader
	PerPeerHeader *PerPeerHeader
	BGPUpdate     []byte
}

// MsgType returns the type of this message
func (rm *RouteMonitoringMsg) MsgType() uint8 {
	return rm.CommonHeader.MsgType
}

func decodeRouteMonitoringMsg(buf *bytes.Buffer, ch *CommonHeader) (*RouteMonitoringMsg, error) {
	rm := &RouteMonitoringMsg{
		CommonHeader: ch,
	}

	pph, err := decodePerPeerHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode per peer header: %v", err)
	}

	rm.PerPeerHeader = pph

	rm.BGPUpdate = make([]byte, ch.MsgLength-CommonHeaderLen-PerPeerHeaderLen)

	fields := []interface{}{
		&rm.BGPUpdate,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return nil, err
	}

	return rm, nil
}
