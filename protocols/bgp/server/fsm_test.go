package server

import (
	"sync"
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/routingtable/filter"

	"net"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

// TestFSM100Updates emulates receiving 100 BGP updates and withdraws. Checks route counts.
func TestFSM100Updates(t *testing.T) {
	fsmA := newFSM2(&peer{
		addr:         net.ParseIP("169.254.100.100"),
		rib:          locRIB.New(),
		importFilter: filter.NewAcceptAllFilter(),
		exportFilter: filter.NewAcceptAllFilter(),
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
			2,      // Path Segement Length
			59, 65, // AS15169
			12, 248, // AS3320
			1,      // Type = AS_SET
			2,      // Path Segement Length
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
	ribRouteCount := fsmA.rib.RouteCount()
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
		ribRouteCount = fsmA.rib.RouteCount()
	}
	time.Sleep(time.Second * 1)

	ribRouteCount = fsmA.rib.RouteCount()
	if ribRouteCount != 0 {
		t.Errorf("Unexpected route count in LocRIB: %d", ribRouteCount)
	}

	fsmA.eventCh <- ManualStop
	wg.Wait()
}
