package server

import (
	"net"
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	assert.Equal(t, &BMPServer{
		routers: map[string]*router{},
	}, s)
}

func TestIntegration(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()
	r.con = conB

	go r.serve()

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 126, // Message Length
		3, // Msg Type = Peer Up Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// Peer Up Notification
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 41, // 10.20.30.41 local address
		0, 123, // Local Port
		0, 234, // Remote Port

		// Sent OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		0, // Opt param length

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 29, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		0, // Opt param length

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 116, // Message Length
		0, // Msg Type = Route Monitoring Message

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		// BGP Update
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 68, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 41, // Total Path Attribute Length

		255,  // Attribute flags
		1,    // Attribute Type code (ORIGIN)
		0, 1, // Length
		2, // INCOMPLETE

		0,      // Attribute flags
		2,      // Attribute Type code (AS Path)
		12,     // Length
		2,      // Type = AS_SEQUENCE
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320
		1,      // Type = AS_SET
		2,      // Path Segment Length
		59, 65, // AS15169
		12, 248, // AS3320

		0,              // Attribute flags
		3,              // Attribute Type code (Next Hop)
		4,              // Length
		10, 11, 12, 13, // Next Hop

		0,          // Attribute flags
		4,          // Attribute Type code (MED)
		4,          // Length
		0, 0, 1, 0, // MED 256

		0,          // Attribute flags
		5,          // Attribute Type code (Local Pref)
		4,          // Length
		0, 0, 1, 0, // Local Pref 256

		// NLRI
		24,
		192, 168, 0,
	})

	if err != nil {
		panic("write #1 failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	if err != nil {
		panic("Write #2 failed")
	}

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}
