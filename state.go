package clock

type state byte

const (
	stateNotReady state = iota
	stateStarted
	stateStopped
)

func (s state) IsStarted() bool {
	return s == stateStarted
}

func (s state) IsStopped() bool {
	return s == stateStopped
}
