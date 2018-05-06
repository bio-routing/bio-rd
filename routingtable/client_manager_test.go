package routingtable

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

type MockClient struct {
	foo int
}

func (m MockClient) AddPath(net.Prefix, *route.Path) error {
	return nil
}
func (m MockClient) RemovePath(net.Prefix, *route.Path) bool {
	return false
}
func (m MockClient) UpdateNewClient(RouteTableClient) error {
	return nil
}

func TestClients(t *testing.T) {
	tests := []struct {
		name     string
		clients  []MockClient
		expected []RouteTableClient
	}{
		{
			name:     "No clients",
			clients:  []MockClient{},
			expected: []RouteTableClient{},
		},
		{
			name: "No clients",
			clients: []MockClient{
				MockClient{
					foo: 1,
				},
				MockClient{
					foo: 2,
				},
			},
			expected: []RouteTableClient{
				MockClient{
					foo: 1,
				},
				MockClient{
					foo: 2,
				},
			},
		},
	}

	for _, test := range tests {
		cm := NewClientManager(MockClient{})
		for _, client := range test.clients {
			cm.Register(client)
		}
		ret := cm.Clients()
		assert.Equal(t, test.expected, ret)
	}
}
