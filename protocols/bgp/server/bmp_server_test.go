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
					"20.30.40.50": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
				},
			},
			expected: &BMPServer{
				routers: make(map[string]*router),
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
					"10.20.30.40": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
				},
			},
		},
		{
			name: "Test with routers no clients",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": &router{
						ribClients: make(map[afiClient]struct{}),
					},
				},
				ribClients: map[string]map[afiClient]struct{}{},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"10.20.30.40": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
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
					"10.20.30.40": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
					"20.30.40.50": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
					"10.20.30.40": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
				},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": &router{
						ribClients: map[afiClient]struct{}{},
					},
					"20.30.40.50": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
					"10.20.30.40": map[afiClient]struct{}{
					},
				},
			},
		},
		{
			name: "Unsubscribe existing from non-router",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.60": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
					"20.30.40.50": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
					"10.20.30.60": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
				},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.60": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
					"20.30.40.50": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.50": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
					"10.20.30.60": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
				},
			},
		},
		{
			name: "Unsubscribe existing from non-existing client",
			srv: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": &router{
						ribClients: map[afiClient]struct{}{},
					},
					"20.30.40.50": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.40": map[afiClient]struct{}{},
					"10.20.30.60": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
					},
				},
			},
			expected: &BMPServer{
				routers: map[string]*router{
					"10.20.30.40": &router{
						ribClients: map[afiClient]struct{}{},
					},
					"20.30.40.50": &router{
						ribClients: map[afiClient]struct{}{
							afiClient{
								afi:    packet.IPv4AFI,
								client: nil,
							}: struct{}{},
						},
					},
				},
				ribClients: map[string]map[afiClient]struct{}{
					"20.30.40.40": map[afiClient]struct{}{},
					"10.20.30.60": map[afiClient]struct{}{
						afiClient{
							afi:    packet.IPv4AFI,
							client: nil,
						}: struct{}{},
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
