package server

import (
	"testing"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestGetAreas(t *testing.T) {
	tests := []struct {
		name     string
		s        *Server
		expected []types.AreaID
	}{
		{
			name: "Test #1",
			s: &Server{
				config: &config.ISISConfig{
					NETs: []config.NET{
						{
							AreaID: types.AreaID{10, 10, 10, 10},
						},
						{
							AreaID: types.AreaID{10, 10, 10, 20},
						},
					},
				},
			},
			expected: []types.AreaID{
				types.AreaID{10, 10, 10, 10},
				types.AreaID{10, 10, 10, 12},
			},
		},
	}

	for _, test := range tests {
		res := test.s.getAreas()
		assert.Equal(t, test.expected, res, test.name)
	}
}
