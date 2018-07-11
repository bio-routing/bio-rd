package server

import (
	"fmt"
	"os"
	"syscall"

	"github.com/bio-routing/bio-rd/config"
)

type ISISServer struct {
	config config.ISISConfig
}

func NewISISServer(cfg config.ISISConfig) *ISISServer {
	return &ISISServer{
		config: cfg,
	}
}

func (isis *ISISServer) Start() error {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("Unable to open socket: %v", err)
	}

	err = syscall.Bind(fd, syscall.RawSockaddrLinklayer{
		Protocol: ETH_P_ALL,
		Family:   AF_PACKET,
	})
	if err != nil {
		return fmt.Errorf("Unable to bind: %v", err)
	}

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))

	go func() {
		for {
			buffer := make([]byte, 1500)
			_, err := f.Read(buffer)
			if err != nil {
				panic(fmt.Sprintf("Unable to read from socket: %v", err))
			}

			fmt.Printf("Packet: %v\n", buffer)
		}
	}()

	return nil
}
