package device

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAdapter struct {
	started   bool
	startFail bool
}

func (m *mockAdapter) start() error {
	m.started = true
	if m.startFail {
		return fmt.Errorf("Fail")
	}

	return nil
}

func (m *mockAdapter) loadAdapter() error {
	return nil
}

func TestStart(t *testing.T) {
	tests := []struct {
		name     string
		adapter  *mockAdapter
		wantFail bool
		expected *mockAdapter
	}{
		{
			name: "Test with failure",
			adapter: &mockAdapter{
				startFail: true,
			},
			wantFail: true,
		},
		{
			name:     "Test with success",
			adapter:  &mockAdapter{},
			wantFail: false,
			expected: &mockAdapter{
				started: true,
			},
		},
	}

	for _, test := range tests {
		s := newWithAdapter(test.adapter)
		err := s.Start()
		if err != nil {
			if test.wantFail {
				continue
			}

			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equalf(t, test.expected, test.adapter, "Test %q", test.name)
	}
}

func TestStop(t *testing.T) {
	a := &mockAdapter{}
	s := newWithAdapter(a)
	s.Stop()

	// This will cause a timeout if channel was not closed
	<-s.done
}

type mockClient struct {
	deviceUpdateCalled uint
}

func (m *mockClient) DeviceUpdate(d *Device) {
	m.deviceUpdateCalled++
}

func TestNotify(t *testing.T) {
	mc := &mockClient{}
	a := &mockAdapter{}
	s := newWithAdapter(a)

	s.addDevice(&Device{
		Name:  "eth0",
		Index: 100,
	})
	s.addDevice(&Device{
		Name:  "eth1",
		Index: 101,
	})
	s.addDevice(&Device{
		Name:  "eth2",
		Index: 102,
	})

	s.Subscribe(mc, "eth1")
	assert.Equal(t, uint(1), mc.deviceUpdateCalled)
	s.notify(100)
	assert.Equal(t, uint(1), mc.deviceUpdateCalled)

	s.notify(101)
	assert.Equal(t, uint(2), mc.deviceUpdateCalled)

	s.delDevice(101)
	s.notify(101)
	assert.Equal(t, uint(2), mc.deviceUpdateCalled)
}
