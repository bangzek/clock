package clock

import "time"

const (
	// Default added duration in Now
	DefaultScriptNow time.Duration = time.Millisecond

	// Default speed ratio for fake timer/ticker
	DefaultScriptRatio time.Duration = 100
)

type Script struct {
	Now   time.Duration
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
