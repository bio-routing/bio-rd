package server

import (
	"fmt"
	"io"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	log "github.com/sirupsen/logrus"
)

func serializeAndSendUpdate(out io.Writer, update *packet.BGPUpdate, opt *packet.Options) error {
	updateBytes, err := update.SerializeUpdate(opt)
	if err != nil {
		log.Errorf("Unable to serialize BGP Update: %v", err)
		return nil
	}

	_, err = out.Write(updateBytes)
	if err != nil {
		return fmt.Errorf("Failed sending Update: %v", err)
	}
	return nil
}
