package flow

// Component is a unit that can start a process
type Component interface {
	Process()
}

// Done notifies that the process is finished
type Done struct{}

// Wait is a channel signalling of a completion
type Wait chan struct{}

// Run the component process
func Run(c Component) Wait {
	wait := make(Wait)
	go func() {
		c.Process()
		wait <- Done{}
	}()
	return wait
}

// InputGuard counts number of closed inputs
type InputGuard struct {
	ports uint
	complete uint
}

// NewInputGuard returns a guard for a given number of inputs
func NewInputGuard(ports uint) *InputGuard {
	return &InputGuard{ports, 0}
}

// Complete is called when a port is closed and returns true when all the ports have been closed
func (g *InputGuard) Complete() bool {
	g.complete++
	return g.complete >= g.ports
}
