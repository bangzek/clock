package clock

import (
	"time"
)

// Clock is interface for time operation, so time sensitive application
// can be easily unit tested by mocking time.
type Clock interface {
	// See [time.Now] for details.
	Now() time.Time
	// See [time.NewTicker] for details.
	NewTicker(d time.Duration) *Ticker
	// See [time.NewTimer] for details.
	NewTimer(d time.Duration) *Timer
}

// Timerable is interface for [time.Timer].
type Timerable interface {
	// See [time.Timer.Reset] for details.
	Reset(d time.Duration) bool
	// See [time.Timer.Stop] for details.
	Stop() bool
}

// Tickerable is interface for [time.Ticker].
type Tickerable interface {
	// See [time.Ticker.Reset] for details.
	Reset(d time.Duration)
	// See [time.Ticker.Stop] for details.
	Stop()
}

// Timer is [time.Timer] drop-in replacement.
type Timer struct {
	// The real timer implementation
	Timerable
	// The channel on which the timer are delivered.
	C <-chan time.Time
}

// Ticker is [time.Ticker] drop-in replacement.
type Ticker struct {
	// The real ticker implementation
	Tickerable
	// The channel on which the ticks are delivered.
	C <-chan time.Time
}

// ===================================================================

type clock struct{}

// New returns a new real-time Clock.
func New() Clock {
	return &clock{}
}

func (c *clock) Now() time.Time {
	return time.Now()
}

func (c *clock) NewTimer(d time.Duration) *Timer {
	t := time.NewTimer(d)
	return &Timer{
		Timerable: t,
		C:         t.C,
	}
}

func (c *clock) NewTicker(d time.Duration) *Ticker {
	t := time.NewTicker(d)
	return &Ticker{
		Tickerable: t,
		C:          t.C,
	}
}
