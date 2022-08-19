package server

import (
	"fmt"
	"io"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/util/log"
)

func serializeAndSendUpdate(out io.Writer, update serializeAbleUpdate, opt *packet.EncodeOptions) error {
	updateBytes, err := update.SerializeUpdate(opt)
	if err != nil {
		log.Errorf("unable to serialize BGP Update: %v", err)
		return nil
	}

	_, err = out.Write(updateBytes)
	if err != nil {
		return fmt.Errorf("failed sending Update: %w", err)
	}
	return nil
}

type serializeAbleUpdate interface {
	SerializeUpdate(opt *packet.EncodeOptions) ([]byte, error)
}
