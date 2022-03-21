package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decoder"
)

// StatsReport represents a stats report message
type StatsReport struct {
	CommonHeader  *CommonHeader
	PerPeerHeader *PerPeerHeader
	StatsCount    uint32
	Stats         []*InformationTLV
}

// MsgType returns the type of this message
func (s *StatsReport) MsgType() uint8 {
	return s.CommonHeader.MsgType
}

func decodeStatsReport(buf *bytes.Buffer, ch *CommonHeader) (Msg, error) {
	sr := &StatsReport{
		CommonHeader: ch,
	}

	pph, err := decodePerPeerHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("unable to decode per peer header: %w", err)
	}

	sr.PerPeerHeader = pph

	fields := []interface{}{
		&sr.StatsCount,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return sr, err
	}

	sr.Stats = make([]*InformationTLV, sr.StatsCount)
	for i := uint32(0); i < sr.StatsCount; i++ {
		infoTLV, err := decodeInformationTLV(buf)
		if err != nil {
			return sr, fmt.Errorf("unable to decode information TLV: %w", err)
		}

		sr.Stats[i] = infoTLV
	}

	return sr, nil
}
