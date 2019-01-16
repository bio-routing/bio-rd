package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

func (d *dev) processPSNP(l *packet.PSNP, srv types.MACAddress) error {
	return fmt.Errorf("Not implemented")
}
