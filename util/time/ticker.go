package time

import (
	gotime "time"
)

// Ticker is a ticker interface that allows mocking tickers
type Ticker interface {
	C() <-chan gotime.Time
	Stop()
}

// BIOTicker is a wrapper for time.Ticker
type BIOTicker struct {
	t  *gotime.Ticker
	ch <-chan gotime.Time
}

// NewBIOTicker creates a new BIO ticker
func NewBIOTicker(interval gotime.Duration) *BIOTicker {
	bt := &BIOTicker{
		t: gotime.NewTicker(interval),
	}

	bt.ch = bt.t.C
	return bt
}

// C returns the channel
func (bt *BIOTicker) C() <-chan gotime.Time {
	return bt.ch
}

// Stop stops the ticker
func (bt *BIOTicker) Stop() {
	bt.t.Stop()
}

// MockTicker os a mocked ticker
type MockTicker struct {
	ch chan gotime.Time
}

// NewMockTicker creates a new mock ticker
func NewMockTicker() *MockTicker {
	return &MockTicker{
		ch: make(chan gotime.Time),
	}
}

// C gets the channel of the ticker
func (m *MockTicker) C() <-chan gotime.Time {
	return m.ch
}

// Stop is here to fulfill an interface
func (m *MockTicker) Stop() {

}

// Tick lets the mock ticker tick
func (m *MockTicker) Tick() {
	m.ch <- gotime.Now()
}
