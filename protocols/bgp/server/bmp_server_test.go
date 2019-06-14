package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	assert.Equal(t, &BMPServer{
		routers:    map[string]*Router{},
		ribClients: map[string]map[afiClient]struct{}{},
	}, s)
}
