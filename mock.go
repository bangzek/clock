package clock

import (
	"strconv"
	"sync"
	"time"
)

// A Mock represents simple mock of clock.
//
// The clock operation can be directed by various *Scripts fields and Default
// field. All fields are optional, the clock can run fine without any script
// and will use DefaultScriptNow and DefaultScriptRatio on zero Default field.
// All clock operation result can be tested using Calls() and Times() method.
//
// This clock runs on fake Timer/Ticker speed divided by ratio so unit testing
// don't have to wait a long time. The default speed ratio can be adjusted on
// Default.Ratio field, or it can be scripted in TimerScripts/TickerScripts.
type Mock struct {
	// How much duration clock.Now() will advance.
	NowScripts []time.Duration
	// The scripts for clock.Timer.
	TimerScripts [][]Script
	// The scripts for clock.Ticker.
	TickerScripts [][]Script
	// The default setting for scripts.
	Default Script

	calls   []string
	nows    []time.Time
	timers  []*mockTimer
	tickers []*mockTicker
	sLock   sync.Mutex
	tLock   sync.Mutex
	cLock   sync.Mutex
	time    time.Time
	iNow    int
	state   state
}

// Start mocking clock and setting time to t.
func (m *Mock) Start(t time.Time) {
	m.sLock.Lock()
	m.state = stateStarted
	m.sLock.Unlock()

	m.tLock.Lock()
	m.time = t
	m.tLock.Unlock()
}

// Stop mocking clock.
func (m *Mock) Stop() {
	m.sLock.Lock()
	if len(m.timers) > 0 {
		for _, t := range m.timers {
			t.stopFake()
		}
		m.timers = m.timers[:0]
	}
	if len(m.tickers) > 0 {
		for _, t := range m.tickers {
			t.stopFake()
		}
		m.tickers = m.tickers[:0]
	}

	m.state = stateStopped
	m.sLock.Unlock()
}

// Returns list of method call.
func (m *Mock) Calls() []string {
	if !m.hasStopped() {
		panic("clock.Mock must be Stop() first")
	}

	m.cLock.Lock()
	defer m.cLock.Unlock()

	return m.calls
}

// Returns list of time.
// This is the result of [clock.Now], Ticker/Timer New or Reset and
// their channel value.
func (m *Mock) Times() []time.Time {
	if !m.hasStopped() {
		panic("clock.Mock must be Stop() first")
	}

	m.tLock.Lock()
	defer m.tLock.Unlock()

	return m.nows
}

// Now returns the current mocked time.
// Please note this always advance the time.
func (m *Mock) Now() time.Time {
	if !m.hasStarted() {
		panic("clock.Mock must be Start() first")
	}

	m.addCall("now")
	return m.incTime(m.incNow())
}

func (m *Mock) incNow() time.Duration {
	if m.iNow++; m.iNow <= len(m.NowScripts) && m.NowScripts[m.iNow-1] > 0 {
		return m.NowScripts[m.iNow-1]
	} else {
		return m.Default.canon().Now
	}
}

func (m *Mock) hasStopped() bool {
	m.sLock.Lock()
	defer m.sLock.Unlock()

	return m.state.IsStopped()
}

func (m *Mock) hasStarted() bool {
	m.sLock.Lock()
	defer m.sLock.Unlock()

	return m.state.IsStarted()
}

func (m *Mock) incTime(d time.Duration) time.Time {
	m.tLock.Lock()
	defer m.tLock.Unlock()

	m.time = m.time.Add(d)
	m.nows = append(m.nows, m.time)
	return m.time
}

func (m *Mock) incTimeTo(t time.Time) {
	m.tLock.Lock()
	if t.After(m.time) {
		m.time = t
		m.nows = append(m.nows, m.time)
	}
	m.tLock.Unlock()
}

func (m *Mock) addCall(call string) {
	m.cLock.Lock()
	m.calls = append(m.calls, call)
	m.cLock.Unlock()
}

// NewTimer returns a new [time.Timer] compatible Timer.
func (m *Mock) NewTimer(d time.Duration) *Timer {
	if !m.hasStarted() {
		panic("clock.Mock must be Start() first")
	}

	m.addCall("timer " + d.String())
	t := new(mockTimer)
	m.timers = append(m.timers, t)
	t.init(m, len(m.timers))
	s := getScript(m.TimerScripts, t.no, &t.i, m.Default)
	t.update(s)
	t.fake = time.NewTimer(d / s.Ratio)
	t.setNow()
	ch := make(chan time.Time, 1)
	go t.run(ch, t.fake.C)

	return &Timer{
		Timerable: t,
		C:         ch,
	}
}

// NewTicker returns a new [time.Ticker] compatible Ticker.
func (m *Mock) NewTicker(d time.Duration) *Ticker {
	if !m.hasStarted() {
		panic("clock.Mock must be Start() first")
	}

	m.addCall("ticker " + d.String())
	t := new(mockTicker)
	m.tickers = append(m.tickers, t)
	t.init(m, len(m.tickers))
	s := getScript(m.TickerScripts, t.no, &t.i, m.Default)
	t.update(s)
	t.fake = time.NewTicker(d / s.Ratio)
	t.setNow()
	ch := make(chan time.Time, 1)
	go t.run(ch, t.fake.C)

	return &Ticker{
		Tickerable: t,
		C:          ch,
	}
}

// ===========================================================================

type common struct {
	stop  chan struct{}
	time  time.Time
	rtime time.Time
	ratio time.Duration
	lock  sync.Mutex
	mock  *Mock
	no    int
	i     int
}

func (c *common) init(mock *Mock, no int) {
	c.mock = mock
	c.no = no
	c.stop = make(chan struct{})
}

func (c *common) update(s Script) {
	c.ratio = s.Ratio
	c.time = c.mock.incTime(s.Now)
}

func (c *common) setNow() {
	c.lock.Lock()
	c.rtime = time.Now()
	c.lock.Unlock()
}

func (c *common) addTime(t time.Time) time.Time {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.time = c.time.Add(t.Sub(c.rtime) * c.ratio)
	c.rtime = t
	return c.time
}

func (c *common) run(dst chan<- time.Time, src <-chan time.Time) {
	for {
		select {
		case <-c.stop:
			return
		case t := <-src:
			nt := c.addTime(t)
			c.mock.incTimeTo(nt)
			if len(dst) == 0 {
				dst <- nt
			}
		}
	}
}

// ===========================================================================

type mockTimer struct {
	common
	fake *time.Timer
}

func (t *mockTimer) Stop() bool {
	t.mock.addCall("timer-" + strconv.Itoa(t.no) + ".stop")
	return t.fake.Stop()
}

func (t *mockTimer) Reset(d time.Duration) bool {
	s := getScript(t.mock.TimerScripts, t.no, &t.i, t.mock.Default)
	t.update(s)
	ret := t.fake.Reset(d / s.Ratio)
	t.setNow()
	t.mock.addCall("timer-" + strconv.Itoa(t.no) + ".reset " + d.String())
	return ret
}

func (t *mockTimer) stopFake() {
	t.lock.Lock()
	t.fake.Stop()
	close(t.stop)
	t.lock.Unlock()
}

// ===========================================================================

type mockTicker struct {
	common
	fake *time.Ticker
}

func (t *mockTicker) Stop() {
	t.fake.Stop()
	t.mock.addCall("ticker-" + strconv.Itoa(t.no) + ".stop")
}

func (t *mockTicker) Reset(d time.Duration) {
	s := getScript(t.mock.TickerScripts, t.no, &t.i, t.mock.Default)
	t.update(s)
	t.fake.Reset(d / s.Ratio)
	t.setNow()
	t.mock.addCall("ticker-" + strconv.Itoa(t.no) + ".reset " + d.String())
}

func (t *mockTicker) stopFake() {
	t.lock.Lock()
	t.fake.Stop()
	close(t.stop)
	t.lock.Unlock()
}
