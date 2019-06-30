package server

import (
	"net"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestBMPServer(t *testing.T) {
	srv := NewServer()

	rtr := newRouter(net.IP{10, 0, 255, 1}, 30119)
	_, pipe := net.Pipe()
	rtr.con = pipe
	srv.addRouter(rtr)

	init := []byte{
		3,           // Version
		0, 0, 0, 22, // Length
		4, // Msg Type (init)

		0, 1, // SysDescr TLV
		0, 4, // Length
		0x42, 0x42, 0x42, 0x42,

		0, 2, // SysName TLV
		0, 4, // Length
		0x41, 0x41, 0x41, 0x41,
	}
	rtr.processMsg(init)

	peerUpA := []byte{
		3,            // Version
		0, 0, 0, 126, // Length
		3, // Msg Type (peer up)

		0,                        // Peer Type (global instance peer)
		0,                        // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 123, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 1, 1, // Peer Address (10.1.1.1)
		0, 0, 0, 200, // Peer AS = 200
		0, 0, 0, 200, // Peer BGP ID = 200
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 1, 2, // Local Address (10.1.1.2)
		0, 222, // Local Port
		0, 179, // Remote Port

		// Sent OPEN
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type (OPEN)
		4,      // BGP Version
		0, 100, // ASN
		0, 180, // Hold Time
		1, 0, 0, 100, // BGP ID
		0, // Ops Param Len

		// Received OPEN
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type (OPEN)
		4,      // BGP Version
		0, 200, // ASN
		0, 180, // Hold Time
		1, 0, 0, 200, // BGP ID
		0, // Ops Param Len
	}
	rtr.processMsg(peerUpA)

	if srv.GetRouter("NotExistent") != nil {
		t.Errorf("GetRouter() returned a non-existent router")
		return
	}

	aaaa := srv.GetRouter("10.0.255.1")
	if aaaa == nil {
		t.Errorf("Router AAAA not found")
		return
	}

	aaaaVRFs := aaaa.GetVRFs()
	if len(aaaaVRFs) != 1 {
		t.Errorf("Unexpected VRF count for router AAAA: %d", len(aaaaVRFs))
		return
	}

	peerUpB := []byte{
		3,            // Version
		0, 0, 0, 126, // Length
		3, // Msg Type (peer up)

		0,                        // Peer Type (global instance peer)
		0,                        // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 123, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 1, // Peer Address (10.1.2.1)
		0, 0, 0, 222, // Peer AS = 222
		0, 0, 0, 222, // Peer BGP ID = 222
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 2, // Local Address (10.1.2.2)
		0, 222, // Local Port
		0, 179, // Remote Port

		// Sent OPEN
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type (OPEN)
		4,      // BGP Version
		0, 100, // ASN
		0, 180, // Hold Time
		1, 0, 0, 100, // BGP ID
		0, // Ops Param Len

		// Received OPEN
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type (OPEN)
		4,      // BGP Version
		0, 222, // ASN
		0, 180, // Hold Time
		1, 0, 0, 222, // BGP ID
		0, // Ops Param Len
	}
	rtr.processMsg(peerUpB)

	aaaaVRFs = aaaa.GetVRFs()
	if len(aaaaVRFs) != 1 {
		t.Errorf("Unexpected VRF count for router AAAA: %d", len(aaaaVRFs))
		return
	}

	peerUpC := []byte{
		3,            // Version
		0, 0, 0, 126, // Length
		3, // Msg Type (peer up)

		0,                      // Peer Type (global instance peer)
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 1, // Peer Address (10.1.3.1)
		0, 0, 0, 233, // Peer AS = 233
		0, 0, 0, 233, // Peer BGP ID = 233
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 2, // Local Address (10.1.3.2)
		0, 222, // Local Port
		0, 179, // Remote Port

		// Sent OPEN
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type (OPEN)
		4,      // BGP Version
		0, 100, // ASN
		0, 180, // Hold Time
		1, 0, 0, 100, // BGP ID
		0, // Ops Param Len

		// Received OPEN
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type (OPEN)
		4,      // BGP Version
		0, 233, // ASN
		0, 180, // Hold Time
		1, 0, 0, 222, // BGP ID
		0, // Ops Param Len
	}
	rtr.processMsg(peerUpC)

	aaaaVRFs = aaaa.GetVRFs()
	if len(aaaaVRFs) != 2 {
		t.Errorf("Unexpected VRF count for router AAAA: %d", len(aaaaVRFs))
		return
	}

	peerDownC := []byte{
		3,           // Version
		0, 0, 0, 69, // Length
		2, // Msg Type (peer down)

		0,                      // Peer Type (global instance peer)
		0,                      // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 0, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 1, // Peer Address (10.1.3.1)
		0, 0, 0, 233, // Peer AS = 233
		0, 0, 0, 233, // Peer BGP ID = 233
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 2, // Local Address (10.1.3.2)
		0, 222, // Local Port
		0, 179, // Remote Port

		4, // Reason = unexpected termination of transport session
	}
	rtr.processMsg(peerDownC)

	if aaaa.neighborManager.getNeighbor(0, [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 1}) != nil {
		t.Errorf("Unexpected neighbor")
		return
	}

	if aaaa.neighborManager.getNeighbor(123, [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 1}) == nil {
		t.Errorf("Expected neighbor not found")
		return
	}

	v := aaaa.GetVRF(123)
	lr := v.IPv4UnicastRIB()
	if lr.Count() != 0 {
		t.Errorf("Unexpected route count")
		return
	}

	updateA1 := []byte{
		3,           // Version
		0, 0, 0, 93, // Length
		0, // Msg Type (route monitoring)

		0,                        // Peer Type (global instance peer)
		0,                        // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 123, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 1, 1, // Peer Address (10.1.1.1)
		0, 0, 0, 200, // Peer AS = 200
		0, 0, 0, 200, // Peer BGP ID = 200
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 45, // Length
		2, // Type (UPDATE)

		0, 0, // Withdraw length
		0, 20, // Total Path Attribute Length

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

		8, 10, // 10.0.0.0/8
	}
	rtr.processMsg(updateA1)

	if lr.Count() != 1 {
		t.Errorf("Unexpected route count")
		return
	}

	route := lr.Get(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8))
	if route == nil {
		t.Errorf("Expected route not found")
		return
	}

	peerDownA := []byte{
		3,           // Version
		0, 0, 0, 69, // Length
		2, // Msg Type (peer down)

		0,                        // Peer Type (global instance peer)
		0,                        // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 123, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 1, 1, // Peer Address (10.1.1.1)
		0, 0, 0, 200, // Peer AS = 200
		0, 0, 0, 200, // Peer BGP ID = 200
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 1, 2, // Local Address (10.1.1.2)
		0, 222, // Local Port
		0, 179, // Remote Port

		4, // Reason = unexpected termination of transport session
	}
	rtr.processMsg(peerDownA)

	if lr.Count() != 0 {
		t.Errorf("Unexpected route count")
		return
	}

	route = lr.Get(bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 0, 0), 8))
	if route != nil {
		t.Errorf("Unexpected route found")
		return
	}

	updateB1 := []byte{
		3,           // Version
		0, 0, 0, 93, // Length
		0, // Msg Type (route monitoring)

		0,                        // Peer Type (global instance peer)
		0,                        // Peer Flags
		0, 0, 0, 0, 0, 0, 0, 123, // Peer Distinguisher
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 1, 2, 1, // Peer Address (10.1.2.1)
		0, 0, 0, 222, // Peer AS = 222
		0, 0, 0, 222, // Peer BGP ID = 222
		0, 0, 0, 0, // Timestamp seconds
		0, 0, 0, 0, // Timestamp microseconds

		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 45, // Length
		2, // Type (UPDATE)

		0, 0, // Withdraw length
		0, 20, // Total Path Attribute Length

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

		8, 10, // 10.0.0.0/8
	}
	rtr.processMsg(updateB1)

	if lr.Count() != 1 {
		t.Errorf("Unexpected route count")
		return
	}

	termination := []byte{
		3,           // Version
		0, 0, 0, 11, // Length
		5, // Msg Type (termination)

		0, 1, // Type = Reason
		0, 1, // Length
		0, // Reason = Admin Down
	}
	rtr.processMsg(termination)

	if lr.Count() != 0 {
		t.Errorf("Unexpected route count")
		return
	}
}
