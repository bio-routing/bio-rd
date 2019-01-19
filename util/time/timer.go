package time

import gotime "time"

// Timer is a timer interface that allows mocking timers
type Timer interface {
	C() <-chan gotime.Time
	Reset(d gotime.Duration) bool
	Stop() bool
}

// BIOTimer is a wrapper for time.Timer
type BIOTimer struct {
	t  *gotime.Timer
	ch <-chan gotime.Time
}

// NewBIOTimer create a new BIO timer
func NewBIOTimer(d gotime.Duration) *BIOTimer {
	bt := &BIOTimer{
		t: gotime.NewTimer(d),
	}

	bt.ch = bt.t.C
	return bt
}

// C gets the channel of the timer
func (m *BIOTimer) C() <-chan gotime.Time {
	return m.ch
}

// Reset resets the timer
func (m *BIOTimer) Reset(d gotime.Duration) bool {
	return m.t.Reset(d)
}

// Stop stops the timer
func (m *BIOTimer) Stop() bool {
	return m.t.Stop()
}

// MockTimer os a mocked timer
type MockTimer struct {
	ch chan gotime.Time
}

// NewMockTimer creates a new mock timer
func NewMockTimer() *MockTimer {
	return &MockTimer{
		ch: make(chan gotime.Time),
	}
}

// C gets the channel of the timer
func (m *MockTimer) C() <-chan gotime.Time {
	return m.ch
}

// Reset resets the timer
func (m *MockTimer) Reset(d gotime.Duration) bool {
	return false
}

// Stop stops the timers
func (m *MockTimer) Stop() bool {
	return false
}

// Timeout lets the mock timer time out
func (m *MockTimer) Timeout() {
	m.ch <- gotime.Now()
}
