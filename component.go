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
		defer func() {
			wait <- Done{}
		}()
		c.Process()
	}()
	return wait
}
