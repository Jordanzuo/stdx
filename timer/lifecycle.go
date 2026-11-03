package timer

import "fmt"

// lifecycleState is the Init/Start/Stop state machine.
// The terminal state is stopped; Init/Start/Stop run once per process.
type lifecycleState int

const (
	idle lifecycleState = iota
	inited
	running
	stopped
)

func (this lifecycleState) String() string {
	switch this {
	case idle:
		return "idle"
	case inited:
		return "inited"
	case running:
		return "running"
	case stopped:
		return "stopped"
	default:
		return fmt.Sprintf("lifecycleState(%d)", int(this))
	}
}
