package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewIdleState(t *testing.T) {
	tests := []struct {
		name     string
		fsm      *FSM2
		expected *idleState
	}{
		{
			name: "Test #1",
			fsm:  &FSM2{},
			expected: &idleState{
				fsm: &FSM2{},
			},
		},
	}

	for _, test := range tests {
		res := newIdleState(test.fsm)
		assert.Equalf(t, test.expected, res, "Test: %s", test.name)
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name     string
		state    *idleState
		expected *idleState
	}{
		{
			name: "Test #1",
			state: &idleState{
				fsm: &FSM2{
					connectRetryCounter: 5,
					connectRetryTimer:   time.NewTimer(time.Second * 20),
				},
				newStateReason: "Foo Bar",
			},
			expected: &idleState{
				fsm: &FSM2{
					connectRetryCounter: 0,
					connectRetryTimer:   time.NewTimer(time.Second * 20),
				},
				newStateReason: "Foo Bar",
			},
		},
	}

	for _, test := range tests {
		if !test.expected.fsm.connectRetryTimer.Stop() {
			<-test.expected.fsm.connectRetryTimer.C
		}
		test.state.start()
		assert.Equalf(t, test.expected, test.state, "Test: %s", test.name)
	}
}
