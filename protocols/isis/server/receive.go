package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/pkg/errors"
)

func (d *dev) receiverRoutine() {
	defer d.wg.Done()

	for {
		select {
		case <-d.done:
			return
		default:
			rawPkt, src, err := d.sys.recvPacket()
			if err != nil {
				d.srv.log.Errorf("recvPacket() failed: %v", err)
				return
			}

			err = d.self.processIngressPacket(rawPkt, src)
			if err != nil {
				d.srv.log.Errorf("Unable to process packet: %v", err)
			}
		}
	}
}

func (d *dev) processIngressPacket(rawPkt []byte, src types.MACAddress) error {
	pkt, err := packet.Decode(bytes.NewBuffer(rawPkt))
	if err != nil {
		return fmt.Errorf("Unable to decode packet from %v on %v: %v: %v", src, d.name, rawPkt, err)
	}

	switch pkt.Header.PDUType {
	case packet.P2P_HELLO:
		err = d.self.processP2PHello(pkt.Body.(*packet.P2PHello), src)
		if err != nil {
			return errors.Wrap(err, "Unable to process P2P Hello")
		}
	case packet.L2_LS_PDU_TYPE:
		err = d.self.processLSPDU(pkt.Body.(*packet.LSPDU), src)
		if err != nil {
			return errors.Wrap(err, "Unable to process LSPDU")
		}
	case packet.L2_CSNP_TYPE:
		err = d.self.processCSNP(pkt.Body.(*packet.CSNP), src)
		if err != nil {
			return errors.Wrap(err, "Unable to process CSNP")
		}
	case packet.L2_PSNP_TYPE:
		err = d.self.processPSNP(pkt.Body.(*packet.PSNP), src)
		if err != nil {
			return errors.Wrap(err, "Unable to process PSNP")
		}
	}

	return nil
}
