package server

import (
	"sync"
	"testing"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/stretchr/testify/assert"
)

// TestFSM255UpdatesIPv4 emulates receiving 255 BGP updates and withdraws. Checks route counts.
func TestFSM255UpdatesIPv4(t *testing.T) {
	fsmA := newFSM(&peer{
		addr:     bnet.IPv4FromOctets(169, 254, 100, 100),
		routerID: bnet.IPv4FromOctets(1, 1, 1, 1).ToUint32(),
		ipv4: &peerAddressFamily{
			rib:          locRIB.New("inet.0"),
			importFilter: filter.NewAcceptAllFilter(),
			exportFilter: filter.NewAcceptAllFilter(),
		},
	})

	fsmA.holdTimer = time.NewTimer(time.Second * 90)
	fsmA.keepaliveTimer = time.NewTimer(time.Second * 30)
	fsmA.connectRetryTimer = time.NewTimer(time.Second * 120)
	fsmA.state = newEstablishedState(fsmA)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fsmA.con = fakeConn{}
		for {
			nextState, reason := fsmA.state.run()
			fsmA.state = nextState
			stateName := stateName(nextState)
			switch stateName {
			case "idle":
				wg.Done()
				return
			case "cease":
				t.Errorf("Unexpected cease state: %s", reason)
				wg.Done()
				return
			case "established":
				continue
			default:
				t.Errorf("Unexpected new state: %s", reason)
				wg.Done()
				return
			}
		}

	}()

	for i := uint8(0); i < 255; i++ {
		a := i % 10
		b := i % 8

		update := []byte{
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			0, 54,
			2,
			0, 0,
			0, 26,
			64, // Attribute flags
			1,  // Attribute Type code (ORIGIN)
			1,  // Length
			2,  // INCOMPLETE

			64,     // Attribute flags
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
			b + 25, 169, a, i, 0,
		}

		fsmA.msgRecvCh <- update

	}

	time.Sleep(time.Second)
	ribRouteCount := fsmA.ipv4Unicast.rib.RouteCount()
	if ribRouteCount != 255 {
		t.Errorf("Unexpected route count in LocRIB: %d", ribRouteCount)
	}

	for i := uint8(0); i < 255; i++ {
		a := i % 10
		b := i % 8

		update := []byte{
			255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			0, 28,
			2,
			0, 5,
			b + 25, 169, a, i, 0,
			0, 0,
		}
		fsmA.msgRecvCh <- update
		ribRouteCount = fsmA.ipv4Unicast.rib.RouteCount()
	}
	time.Sleep(time.Second * 1)

	ribRouteCount = fsmA.ipv4Unicast.rib.RouteCount()
	if ribRouteCount != 0 {
		t.Errorf("Unexpected route count in LocRIB: %d", ribRouteCount)
	}

	fsmA.eventCh <- ManualStop
	wg.Wait()
}

// TestFSM255UpdatesIPv6 emulates receiving 255 BGP updates and withdraws. Checks route counts.
func TestFSM255UpdatesIPv6(t *testing.T) {
	fsmA := newFSM(&peer{
		addr:     bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0xffff, 0, 0, 0, 1),
		routerID: bnet.IPv4FromOctets(1, 1, 1, 1).ToUint32(),
		ipv6: &peerAddressFamily{
			rib:          locRIB.New("inet6.0"),
			importFilter: filter.NewAcceptAllFilter(),
			exportFilter: filter.NewAcceptAllFilter(),
		},
	})

	fsmA.ipv6Unicast.multiProtocol = true
	fsmA.holdTimer = time.NewTimer(time.Second * 90)
	fsmA.keepaliveTimer = time.NewTimer(time.Second * 30)
	fsmA.connectRetryTimer = time.NewTimer(time.Second * 120)
	fsmA.state = newEstablishedState(fsmA)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fsmA.con = fakeConn{}
		for {
			nextState, reason := fsmA.state.run()
			fsmA.state = nextState
			stateName := stateName(nextState)
			switch stateName {
			case "idle":
				wg.Done()
				return
			case "cease":
				t.Errorf("Unexpected cease state: %s", reason)
				wg.Done()
				return
			case "established":
				continue
			default:
				t.Errorf("Unexpected new state: %s", reason)
				wg.Done()
				return
			}
		}

	}()

	for i := uint8(0); i < 255; i++ {
		update := []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0, 76,
			2,
			0, 0,
			0, 53,
			64, // Attribute flags
			1,  // Attribute Type code (ORIGIN)
			1,  // Length
			2,  // INCOMPLETE

			64,     // Attribute flags
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

			0x90,     // Attribute flags
			0x0e,     // MP_REACH_NLRI
			0x00, 30, // Length
			0x00, 0x02, // AFI
			0x01,                                                                                                 // SAFI
			0x10, 0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Nexthop
			0x00,
			64, 0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0x0, i,
		}

		fsmA.msgRecvCh <- update
	}

	time.Sleep(time.Second)
	ribRouteCount := fsmA.ipv6Unicast.rib.RouteCount()
	if ribRouteCount != 255 {
		t.Errorf("Unexpected route count in LocRIB: %d", ribRouteCount)
	}

	for i := uint8(0); i < 255; i++ {
		update := []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x00, 35, // Length
			0x02,       // UPDATE
			0x00, 0x00, // withdrawn routes
			0x00, 0x0c,
			0x90, 0x0f,
			0x00, 12, // Length
			0x00, 0x02, // AFI
			0x01, // SAFI
			64, 0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0x0, i,
		}
		fsmA.msgRecvCh <- update
		ribRouteCount = fsmA.ipv6Unicast.rib.RouteCount()
	}
	time.Sleep(time.Second * 1)

	ribRouteCount = fsmA.ipv6Unicast.rib.RouteCount()
	if ribRouteCount != 0 {
		t.Errorf("Unexpected route count in LocRIB: %d", ribRouteCount)
	}

	fsmA.eventCh <- ManualStop
	wg.Wait()
}

func TestOpenMessage(t *testing.T) {
	tests := []struct {
		name     string
		localASN uint32
		holdTime time.Duration
		routerID uint32
		expected packet.BGPOpen
	}{
		{
			name:     "16bit ASN",
			localASN: 12345,
			holdTime: time.Duration(30 * time.Second),
			routerID: 1,
			expected: packet.BGPOpen{
				ASN:           12345,
				BGPIdentifier: 1,
				HoldTime:      30,
				OptParams: []packet.OptParam{
					{
						Type: packet.CapabilitiesParamType,
						Value: packet.Capabilities{
							packet.Capability{
								Code: 65,
								Value: packet.ASN4Capability{
									ASN4: 12345,
								},
							},
						},
					},
				},
				Version: 4,
			},
		},
		{
			name:     "32bit ASN",
			localASN: 202739,
			holdTime: time.Duration(30 * time.Second),
			routerID: 1,
			expected: packet.BGPOpen{
				ASN:           23456,
				BGPIdentifier: 1,
				HoldTime:      30,
				OptParams: []packet.OptParam{
					{
						Type: packet.CapabilitiesParamType,
						Value: packet.Capabilities{
							packet.Capability{
								Code: 65,
								Value: packet.ASN4Capability{
									ASN4: 202739,
								},
							},
						},
					},
				},
				Version: 4,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := peer{
				localASN: test.localASN,
				holdTime: test.holdTime,
				routerID: test.routerID,
				optOpenParams: []packet.OptParam{
					{
						Type: packet.CapabilitiesParamType,
						Value: packet.Capabilities{
							packet.Capability{
								Code: 65,
								Value: packet.ASN4Capability{
									ASN4: test.localASN,
								},
							},
						},
					},
				},
			}

			fsm := newFSM(&p)
			msg := fsm.openMessage()

			assert.Equal(t, &test.expected, msg)
		})
	}
}
