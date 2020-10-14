package goflow

// Component is a unit that can start a process
type Component interface {
	Process()
}

// Done notifies that the process is finished
type Done struct{}

// Wait is a channel signaling of a completion
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
	ports    map[string]bool
	complete int
}

// NewInputGuard returns a guard for a given number of inputs
func NewInputGuard(ports ...string) *InputGuard {
	portMap := make(map[string]bool, len(ports))
	for _, p := range ports {
		portMap[p] = false
	}

	return &InputGuard{portMap, 0}
}

// Complete is called when a port is closed and returns true when all the ports have been closed
func (g *InputGuard) Complete(port string) bool {
	if !g.ports[port] {
		g.ports[port] = true
		g.complete++
	}

	return g.complete >= len(g.ports)
}
