package clock

import (
	"strconv"
	"sync"
	"time"
)

// A Mock represents simple mock of clock.
//
// The clock operation can be directed by various *Scripts fields. All scripts
// are optional, the clock can run fine without any script. All clock
// operation will be recorded on Calls field.
//
// This clock runs on fake timer/ticker speed divided by ratio so unit testing
// don't have to wait a long time. The default speed ratio can be adjusted on
// Default.Ratio field, or it can be scripted in
// TimerScripts/TickerScripts field.
type Mock struct {
	NowScripts    []time.Duration
	TimerScripts  [][]Script
	TickerScripts [][]Script
	Calls         []string

	Default Script

	lock    sync.Mutex
	time    time.Time
	iNow    int
	iTimer  int
	iTicker int
}

// NewMock creates a new Mock with the time initialized to t.
func NewMock(t time.Time) *Mock {
	return &Mock{
		time: t,
	}
}

// Now returns the current mocked time.
// Please note this always advance the time.
func (m *Mock) Now() time.Time {
	m.incTime(m.incNow())
	m.Calls = append(m.Calls, m.time.Format("now "+time.RFC3339Nano))
	return m.time
}

func (m *Mock) incNow() time.Duration {
	if m.iNow++; m.iNow <= len(m.NowScripts) && m.NowScripts[m.iNow-1] > 0 {
		return m.NowScripts[m.iNow-1]
	} else {
		return m.Default.canon().Now
	}
}

func (m *Mock) incTime(d time.Duration) time.Time {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.time = m.time.Add(d)
	return m.time
}

func (m *Mock) incTimeTo(t time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if t.After(m.time) {
		m.time = t
	}
}

// NewTimer returns a new time.Timer compatible timer.
func (m *Mock) NewTimer(d time.Duration) *Timer {
	m.Calls = append(m.Calls, "timer "+d.String())

	m.iTimer++
	t := &mockTimer{
		no: m.iTimer,
	}
	t.mock = m
	s := getScript(m.TimerScripts, t.no, &t.i, m.Default)
	t.set(s.Ratio, m.incTime(s.Now))
	t.fake = time.NewTimer(d / s.Ratio)
	ch := make(chan time.Time, 1)
	go t.run(ch, t.fake.C)

	return &Timer{
		Timerable: t,
		C:         ch,
	}
}

// NewTicker returns a new time.Ticker compatible ticker.
func (m *Mock) NewTicker(d time.Duration) *Ticker {
	m.Calls = append(m.Calls, "ticker "+d.String())

	m.iTicker++
	t := &mockTicker{
		no: m.iTicker,
	}
	t.mock = m
	s := getScript(m.TickerScripts, t.no, &t.i, m.Default)
	t.set(s.Ratio, m.incTime(s.Now))
	t.fake = time.NewTicker(d / s.Ratio)
	ch := make(chan time.Time, 1)
	go t.run(ch, t.fake.C)

	return &Ticker{
		Tickerable: t,
		C:          ch,
	}
}

// ===========================================================================

type rat struct {
	mock  *Mock
	time  time.Time
	rtime time.Time
	ratio time.Duration
}

func (r *rat) set(rat time.Duration, t time.Time) {
	r.ratio = rat
	r.time = t
	r.rtime = time.Now()
}

func (r *rat) run(dst chan<- time.Time, src <-chan time.Time) {
	for t := range src {
		r.time = r.time.Add(t.Sub(r.rtime) * r.ratio)
		r.mock.incTimeTo(r.time)
		if len(dst) == 0 {
			dst <- r.time
		}
		r.rtime = t
	}
}

// ===========================================================================

type mockTimer struct {
	rat
	fake *time.Timer
	no   int
	i    int
}

func (t *mockTimer) Stop() bool {
	t.mock.Calls = append(t.mock.Calls, "timer-"+strconv.Itoa(t.no)+".stop")
	return t.fake.Stop()
}

func (t *mockTimer) Reset(d time.Duration) bool {
	t.mock.Calls = append(t.mock.Calls,
		"timer-"+strconv.Itoa(t.no)+".reset "+d.String())

	s := getScript(t.mock.TimerScripts, t.no, &t.i, t.mock.Default)
	t.set(s.Ratio, t.mock.incTime(s.Now))
	return t.fake.Reset(d / s.Ratio)
}

// ===========================================================================

type mockTicker struct {
	rat
	fake *time.Ticker
	no   int
	i    int
}

func (t *mockTicker) Stop() {
	t.mock.Calls = append(t.mock.Calls, "ticker-"+strconv.Itoa(t.no)+".stop")
	t.fake.Stop()
}

func (t *mockTicker) Reset(d time.Duration) {
	t.mock.Calls = append(t.mock.Calls,
		"ticker-"+strconv.Itoa(t.no)+".reset "+d.String())

	s := getScript(t.mock.TickerScripts, t.no, &t.i, t.mock.Default)
	t.set(s.Ratio, t.mock.incTime(s.Now))
	t.fake.Reset(d / s.Ratio)
}
