package main

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/device"
)

type Client struct {
}

func (c *Client) LinkUpdate(lu device.LinkUpdate) {
	fmt.Printf("Link Update! %s\n", lu.Name)
}

func main() {
	s := device.New()
	s.Start()

	c := &Client{}
	s.Subscribe(c, "virbr0")

	select {}
}
