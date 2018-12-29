package server

import (
	"net"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	assert.Equal(t, &BMPServer{
		routers:    map[string]*router{},
		ribClients: map[string]map[afiClient]struct{}{},
	}, s)
}

func TestSubscribeRIBs(t *testing.T) {
	tests := []struct {
		name     string
		srv      *BMPServer
		expected *BMPServer
	}{
		{
			name: "Test without routers with clients",
			srv: &BMPServer{
				routers: make(map[string]*router),
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
			expected: &BMPServer{
				routers: make(map[string]*router),
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
					"10.20.30.40": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
		},
		{
			name: "Test with routers no clients",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": {
						ribClients: make(map[afiClient]struct{}),
					},
				},
				ribClients: map[string]map[afiClient]struct{}{},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"10.20.30.40": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.srv.SubscribeRIBs(nil, net.IP{10, 20, 30, 40}, packet.IPv4AFI)

		assert.Equalf(t, test.expected, test.srv, "Test %q", test.name)

	}
}

func TestUnsubscribeRIBs(t *testing.T) {
	tests := []struct {
		name     string
		srv      *BMPServer
		expected *BMPServer
	}{
		{
			name: "Unsubscribe existing from router",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
					"20.30.40.50": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
					"10.20.30.40": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": {
						ribClients: map[afiClient]struct{}{},
					},
					"20.30.40.50": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
					"10.20.30.40": {},
				},
			},
		},
		{
			name: "Unsubscribe existing from non-router",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.60": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
					"20.30.40.50": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
					"10.20.30.60": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.60": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
					"20.30.40.50": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
					"10.20.30.60": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
		},
		{
			name: "Unsubscribe existing from non-existing client",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": {
						ribClients: map[afiClient]struct{}{},
					},
					"20.30.40.50": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.40": {},
					"10.20.30.60": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": {
						ribClients: map[afiClient]struct{}{},
					},
					"20.30.40.50": {
						ribClients: map[afiClient]struct{}{
							{
								afi:    packet.IPv4AFI,
								client: nil,
							}: {},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.40": {},
					"10.20.30.60": {
						{
							afi:    packet.IPv4AFI,
							client: nil,
						}: {},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.srv.UnsubscribeRIBs(nil, net.IP{10, 20, 30, 40}, packet.IPv4AFI)

		assert.Equalf(t, test.expected, test.srv, "Test %q", test.name)

	}
}
