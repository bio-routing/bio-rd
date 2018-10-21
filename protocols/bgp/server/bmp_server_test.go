package server

import (
	"net"
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	biotesting "github.com/bio-routing/bio-rd/testing"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	assert.Equal(t, &BMPServer{
		routers: map[string]*router{},
	}, s)
}

func TestIntegrationPeerUpRouteMonitor(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

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
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv6IPv4(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
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
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
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
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

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

		// THIRD MESSAGE
		// Common Header
		3,            // Version
		0, 0, 0, 113, // Message Length
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
		0, 65, // Length
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

		0, // /0
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv4 route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv4IPv6(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
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
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 113, // Message Length
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
		0, 65, // Length
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

		0, // /0

		// THIRD MESSAGE

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
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
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

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
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv4 route count. Expected: 1 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationPeerUpRouteMonitorIPv6(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	conA, conB := net.Pipe()

	go r.serve(conB)

	// Peer Up Notification
	_, err := conA.Write([]byte{
		// Common Header
		3,            // Version
		0, 0, 0, 142, // Message Length
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
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 200, // AS
		0, 180, // Hold Time
		1, 0, 0, 1, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// Received OPEN message
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 37, // Length
		1,      // Open message type
		4,      // BGP Version
		0, 100, // AS
		0, 180, // Hold Time
		1, 0, 0, 255, // BGP Identifier
		8, // Opt param length
		2, 6,
		1, 4, // MP BGP
		0, 2, // IPv6
		0, 1, // Unicast

		// SECOND MESSAGE:

		// Common Header
		3,            // Version
		0, 0, 0, 138, // Message Length
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
		0, 90, // Length
		2, // Update

		0, 0, // Withdrawn Routes Length
		0, 67, // Total Path Attribute Length

		255,
		14,    // MP REACH NLRI
		0, 22, // Length
		0, 2, // IPv6
		1,  // Unicast
		16, // IPv6 Next Hop
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, // Reserved
		0, // Pfxlen /0

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
	})

	if err != nil {
		panic("write failed")
	}

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib6.RouteCount()
	if count != 1 {
		t.Errorf("Unexpected IPv6 route count. Expected: 1 Got: %d", count)
	}

	count = rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected IPv4 route count. Expected: 0 Got: %d", count)
	}

	conA.Close()
}

func TestIntegrationIncompleteBMPMsg(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	// Peer Up Notification with invalid version number (4)
	con.Write([]byte{
		// Common Header
		4,            // Version
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

	})

	r.serve(con)

	if !con.Closed {
		t.Errorf("Connection not closed although failure should have happened")
	}
}

func TestBMPFullRunWithWithdraw(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	go r.serve(con)

	con.Write([]byte{
		// #####################################################################
		// Initiation Message
		// #####################################################################
		// Common Header
		3,
		0, 0, 0, 23,
		4,
		0, 1, // sysDescr
		0, 4, // Length
		42, 42, 42, 42, // AAAA
		0, 2, //sysName
		0, 5, // Length
		43, 43, 43, 43, 43, // BBBBB

		// #####################################################################
		// Peer UP Notification
		// #####################################################################
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

		// #####################################################################
		// Route Monitoring Message #1
		// #####################################################################
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

		// #####################################################################
		// Route Monitoring Message #2 (withdraw of prefix from #1)
		// #####################################################################
		// Common Header
		3,           // Version
		0, 0, 0, 75, // Message Length
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
		0, 27, // Length
		2, // Update

		0, 4, // Withdrawn Routes Length
		24,
		192, 168, 0,
		0, 0, // Total Path Attribute Length
	})

	time.Sleep(time.Millisecond * 50)
	assert.NotEmpty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected route count. Expected: 0 Got: %d", count)
	}

}

func TestBMPFullRunWithPeerDownNotification(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	go r.serve(con)

	con.Write([]byte{
		// #####################################################################
		// Initiation Message
		// #####################################################################
		// Common Header
		3,
		0, 0, 0, 23,
		4,
		0, 1, // sysDescr
		0, 4, // Length
		42, 42, 42, 42, // AAAA
		0, 2, //sysName
		0, 5, // Length
		43, 43, 43, 43, 43, // BBBBB

		// #####################################################################
		// Peer UP Notification
		// #####################################################################
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

		// #####################################################################
		// Route Monitoring Message #1
		// #####################################################################
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

		// #####################################################################
		// Peer Down Notification
		// #####################################################################
		// Common Header
		3,           // Version
		0, 0, 0, 49, // Message Length
		2, // Msg Type = Peer Down Notification

		// Per Peer Header
		0,                      // Peer Type
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 20, 30, 40, // 10.20.30.40 peer address
		0, 0, 0, 100, // Peer AS
		0, 0, 0, 255, // Peer BGP ID
		0, 0, 0, 0, // Timestamp s
		0, 0, 0, 0, // Timestamp µs

		4, // Reason
	})

	time.Sleep(time.Millisecond * 50)
	assert.Empty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected route count. Expected: 0 Got: %d", count)
	}
}

func TestBMPFullRunWithTerminationMessage(t *testing.T) {
	addr := net.IP{10, 20, 30, 40}
	port := uint16(12346)

	rib4 := locRIB.New()
	rib6 := locRIB.New()

	r := newRouter(addr, port, rib4, rib6)
	con := biotesting.NewMockConn()

	go r.serve(con)

	con.Write([]byte{
		// #####################################################################
		// Initiation Message
		// #####################################################################
		// Common Header
		3,
		0, 0, 0, 23,
		4,
		0, 1, // sysDescr
		0, 4, // Length
		42, 42, 42, 42, // AAAA
		0, 2, //sysName
		0, 5, // Length
		43, 43, 43, 43, 43, // BBBBB

		// #####################################################################
		// Peer UP Notification
		// #####################################################################
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

		// #####################################################################
		// Route Monitoring Message #1
		// #####################################################################
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

		// #####################################################################
		// Termination Message
		// #####################################################################
		// Common Header
		3,           // Version
		0, 0, 0, 12, // Message Length
		5, // Msg Type = Termination Message

		// TLV
		0, 0, 0, 2,
		42, 42,
	})

	time.Sleep(time.Millisecond * 50)
	assert.Empty(t, r.neighbors)

	time.Sleep(time.Millisecond * 50)

	count := rib4.RouteCount()
	if count != 0 {
		t.Errorf("Unexpected route count. Expected: 0 Got: %d", count)
	}
}
