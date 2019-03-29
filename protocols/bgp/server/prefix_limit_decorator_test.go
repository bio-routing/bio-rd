package server

import (
	"testing"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	btesting "github.com/bio-routing/bio-rd/testing"
	"github.com/stretchr/testify/assert"
)

func TestAddPath(t *testing.T) {
	client := routingtable.NewRTMockClient()
	client.AddError = routingtable.NewPrefixLimitError(100)

	fsm := newFSM(&peer{})
	fsm.con = btesting.NewMockConn()

	decorator := &prefixLimitDecorator{
		client: client,
		fsm:    fsm,
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		e := <-fsm.eventCh
		assert.Equal(t, Cease, e)
		done <- struct{}{}
	}()

	decorator.AddPath(bnet.NewPfx(bnet.IPv4(0), 32), &route.Path{})

	select {
	case <-done:
		return
	case <-time.After(1 * time.Second):
		t.Fatal("cease event not raised")
	}
}
