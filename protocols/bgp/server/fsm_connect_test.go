package server

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectStateManualStop(t *testing.T) {
	fsm := &FSM2{
		eventCh:             make(chan int),
		connectRetryCounter: 100,
		connectRetryTimer:   time.NewTimer(time.Second * 120),
	}
	fsm.startConnectRetryTimer()
	fsm.state = newConnectState(fsm)

	var wg sync.WaitGroup
	var nextState state
	var reason string
	wg.Add(1)
	go func() {
		nextState, reason = fsm.state.run()
		wg.Done()
	}()

	fsm.eventCh <- ManualStop
	wg.Wait()

	assert.IsType(t, &idleState{}, nextState, "Unexpected state returned")
	assert.Equalf(t, 0, fsm.connectRetryCounter, "Unexpected resetConnectRetryCounter: %d", fsm.connectRetryCounter)
}

func TestConnectStateConnectRetryTimer(t *testing.T) {
	fsm := &FSM2{
		eventCh:           make(chan int),
		connectRetryTimer: time.NewTimer(time.Second * 120),
	}
	fsm.startConnectRetryTimer()
	fsm.state = newConnectState(fsm)

	var wg sync.WaitGroup
	var nextState state
	var reason string
	wg.Add(1)
	go func() {
		fsm.connectRetryTimer = time.NewTimer(time.Duration(0))
		nextState, reason = fsm.state.run()
		wg.Done()
	}()

	wg.Wait()

	assert.IsType(t, &connectState{}, nextState, "Unexpected state returned")
}

func TestConnectStateConEstablished(t *testing.T) {
	fsm := &FSM2{
		eventCh:           make(chan int),
		connectRetryTimer: time.NewTimer(time.Second * 120),
	}
	fsm.startConnectRetryTimer()
	fsm.state = newConnectState(fsm)

	var wg sync.WaitGroup
	var nextState state
	var reason string
	wg.Add(1)
	go func() {
		fsm.connectRetryTimer = time.NewTimer(time.Duration(0))
		nextState, reason = fsm.state.run()
		wg.Done()
	}()

	wg.Wait()

	assert.IsType(t, &connectState{}, nextState, "Unexpected state returned")
}
