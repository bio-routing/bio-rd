package main

import (
	"fmt"
	"os"

	"github.com/bio-routing/bio-rd/protocols/device"
	log "github.com/sirupsen/logrus"
)

// Client is a device protocol client
type Client struct {
}

// DeviceUpdate is a callback to get updated device information
func (c *Client) DeviceUpdate(d *device.Device) {
	fmt.Printf("Device Update! %s\n", d.Name)
	fmt.Printf("New State: %v\n", d.OperState)
}

func main() {
	s, err := device.New()
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	err = s.Start()
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	c := &Client{}
	intf := "virbr0"
	err = s.Subscribe(c, intf)
	if err != nil {
		fmt.Printf("Error while subscribing to interface %s: %v", intf, err)
		os.Exit(1)
	}

	select {}
}
