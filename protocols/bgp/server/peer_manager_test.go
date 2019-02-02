package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestAdd(t *testing.T) {
	ip := bnet.IPv4FromOctets(192, 168, 0, 1)
	p := &peer{
		addr: ip,
	}

	m := newPeerManager()
	m.add(p)

	found, _ := m.peers[ip]
	assert.Exactly(t, p, found)
}

func TestRemove(t *testing.T) {
	ip := bnet.IPv4FromOctets(192, 168, 0, 1)
	p := &peer{
		addr: ip,
	}

	m := newPeerManager()
	m.peers[ip] = p

	m.remove(ip)

	assert.Empty(t, m.peers)
}

func TestGet(t *testing.T) {
	ip := bnet.IPv4FromOctets(192, 168, 0, 1)
	p := &peer{
		addr: ip,
	}

	m := newPeerManager()
	m.peers[ip] = p

	found := m.get(ip)
	assert.Exactly(t, p, found)
}

func TestList(t *testing.T) {
	p1 := &peer{
		addr: bnet.IPv4FromOctets(192, 168, 0, 1),
	}
	p2 := &peer{
		addr: bnet.IPv4FromOctets(192, 168, 0, 2),
	}

	m := newPeerManager()
	m.peers[p1.GetAddr()] = p1
	m.peers[p2.GetAddr()] = p2

	list := m.list()
	assert.Contains(t, list, p1)
	assert.Contains(t, list, p2)
}
