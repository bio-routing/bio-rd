package server

import (
	"fmt"
	"io"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	log "github.com/sirupsen/logrus"
)

func serializeAndSendUpdate(out io.Writer, update serializeAbleUpdate, opt *types.Options) error {
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

type serializeAbleUpdate interface {
	SerializeUpdate(opt *types.Options) ([]byte, error)
}
