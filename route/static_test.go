package route

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	apinet "github.com/bio-routing/bio-rd/net/api"
	api "github.com/bio-routing/bio-rd/route/api"
	"github.com/stretchr/testify/assert"
)

func TestStaticToProto(t *testing.T) {
	tests := []struct {
		name     string
		s        *StaticPath
		expected *api.StaticPath
	}{
		{
			name: "Some static path",
			s: &StaticPath{
				NextHop: bnet.IPv4(123),
			},
			expected: &api.StaticPath{
				NextHop: &apinet.IP{
					Version: apinet.IP_IPv4,
					Lower:   123,
				},
			},
		},
	}

	for _, test := range tests {
		res := test.s.ToProto()
		assert.Equal(t, test.expected, res, test.name)
	}
}
