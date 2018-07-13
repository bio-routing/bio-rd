package server

import (
	"time"
	"fmt"
	"net"
	"syscall"

	"github.com/bio-routing/bio-rd/config"

	"github.com/mdlayher/raw"
)

const (
	maxEtherFrameSize = 9216
)

type ISISServer struct {
	config config.ISISConfig
}

type isisInterface struct {
	name string
	rx chan []byte
}

func NewISISServer(cfg config.ISISConfig) *ISISServer {
	return &ISISServer{
		config: cfg,
	}
}

func newISISInterface(name string) *isisInterface {
	return &isisInterface{
		name: name,
		rx: make(chan []byte),
	}
}

func (isis *ISISServer) Start() error {
	for _, ifs := range isis.config.Interfaces {
		ifa, err := net.InterfaceByName(ifs.Name)
		if err != nil {
			return fmt.Errorf("Unable to get interface: %v", err)
		}

		c, err := raw.ListenPacket(ifa, syscall.ETH_P_802_2, nil)
		if err != nil {
			return fmt.Errorf("Unable to listen: %v", err)
		}

		c.SetPromiscuous(true)

		go func() {
			buffer := make([]byte, maxEtherFrameSize)
			for {
				n, addr, err := c.ReadFrom(buffer)
				if err != nil {
					panic(fmt.Sprintf("Unable to read from socket: %v", err))
				}

				fmt.Printf("Packet from %v: %v\n", addr.String(), buffer[:n])
			}
		}()

		go func() {
			for {
				c.WriteTo([]byte{1, 2, 3, 4}, net.HardwareAddr([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}))
				time.Sleep(time.Second)
			}
		}()
	}

	return nil
}

func (isis *ISISServer) Stop() error {

	return nil
}