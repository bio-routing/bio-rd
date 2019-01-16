package server

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/pkg/errors"
)

func (d *dev) receiverRoutine() {
	defer d.wg.Done()

	for {
		select {
		case <-d.done:
			return
		default:
			err := d.self.processIngressPacket()
			if err != nil {
				d.srv.log.Errorf("%v", err)
				return
			}
		}
	}
}

func (d *dev) processIngressPacket() error {
	rawPkt, src, err := d.sys.recvPacket()
	if err != nil {
		return errors.Wrap(err, "receiving packet failed")
	}

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
