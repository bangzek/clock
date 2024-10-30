package clock

import "time"

const (
	// Default added duration in Script.Now.
	DefaultScriptNow time.Duration = time.Millisecond

	// Default speed Script.Ratio for fake timer/ticker.
	DefaultScriptRatio time.Duration = 100
)

type Script struct {
	// How much duration clock.Now() will advance.
	Now time.Duration
	// The speed ratio between test's fake timer/ticker and real.
	Ratio time.Duration
}

func (s Script) canon() Script {
	if s.Now <= 0 {
		s.Now = DefaultScriptNow
	}
	if s.Ratio <= 0 {
		s.Ratio = DefaultScriptRatio
	}
	return s
}

func getScript(l [][]Script, no int, i *int, def Script) Script {
	if no <= len(l) && len(l[no-1]) > *i {
		*i++
		return l[no-1][*i-1].canon()
	} else {
		return def.canon()
	}
}
