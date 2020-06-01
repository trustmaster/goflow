package dsl

// Merge simply sends its input to output, but its inport connection
// can be used as FIFO to merge multiple connections
type Merge struct {
	In  <-chan Token
	Out chan<- Token
}

// Process sends incoming packets to output
func (m *Merge) Process() {
	for t := range m.In {
		m.Out <- t
	}
}
