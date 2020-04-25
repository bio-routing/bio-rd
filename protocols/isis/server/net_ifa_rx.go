package server

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (nifa *netIfa) receiver() {
	for {
		err := nifa.receive()
		if err != nil {
			log.Errorf("IS-IS: receive(): %v", err)
		}
	}
}

func (nifa *netIfa) receive() error {
	fmt.Printf("received: Waiting for incoming packet\n")
	pkt, src, err := nifa.ethHandler.RecvPacket()
	panic(pkt)
	if err != nil {
		panic(err)
		return errors.Wrap(err, "Read failed")
	}

	fmt.Printf("Received %v from %v\n", pkt, src)

	return nil
}
