package dsl

// Split copies an incoming token to each of the array outputs
type Split struct {
	In  <-chan Token
	Out [](chan<- Token)
}

// Process copies incoming tokens to the outputs
func (s *Split) Process() {
	outsCount := len(s.Out)

	for tok := range s.In {
		for i := 0; i < outsCount; i++ {
			tokCopy := tok
			s.Out[i] <- tokCopy
		}
	}

	// TODO Move closing array ports to GoFlow runtime?
	for i := 0; i < outsCount; i++ {
		close(s.Out[i])
	}
}
