package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestAdd(t *testing.T) {
	ip := bnet.IPv4FromOctets(192, 168, 0, 1).Ptr()
	p := &peer{
		addr: ip,
	}

	m := newPeerManager()
	m.add(p)

	found := m.peers[p.peerKey()]
	assert.Exactly(t, p, found)
}

func TestRemove(t *testing.T) {
	ip := bnet.IPv4FromOctets(192, 168, 0, 1).Ptr()
	p := &peer{
		addr: ip,
	}

	m := newPeerManager()
	m.peers[p.peerKey()] = p

	m.remove(p.peerKey())

	assert.Empty(t, m.peers)
}

func TestGet(t *testing.T) {
	ip := bnet.IPv4FromOctets(192, 168, 0, 1).Ptr()
	p := &peer{
		addr: ip,
	}

	m := newPeerManager()
	m.peers[p.peerKey()] = p

	found := m.get(nil, ip)
	assert.Exactly(t, p, found)
}

func TestList(t *testing.T) {
	p1 := &peer{
		addr: bnet.IPv4FromOctets(192, 168, 0, 1).Ptr(),
	}
	p2 := &peer{
		addr: bnet.IPv4FromOctets(192, 168, 0, 2).Ptr(),
	}

	m := newPeerManager()
	m.peers[p1.peerKey()] = p1
	m.peers[p2.peerKey()] = p2

	list := m.list()
	assert.Contains(t, list, p1)
	assert.Contains(t, list, p2)
}
