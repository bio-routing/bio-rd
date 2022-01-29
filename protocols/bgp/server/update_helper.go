package server

import (
	"io"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

func serializeAndSendUpdate(out io.Writer, update serializeAbleUpdate, opt *packet.EncodeOptions, safi uint8) error {
	updateBytes, err := update.SerializeUpdate(opt, safi)
	if err != nil {
		log.Errorf("Unable to serialize BGP Update: %v", err)
		return nil
	}

	_, err = out.Write(updateBytes)
	if err != nil {
		return errors.Wrap(err, "Failed sending Update")
	}
	return nil
}

type serializeAbleUpdate interface {
	SerializeUpdate(opt *packet.EncodeOptions, safi uint8) ([]byte, error)
}
