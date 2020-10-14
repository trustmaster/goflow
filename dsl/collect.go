package dsl

// Collect gathers tokens processed in parallel, returns a matched one,
// and moves pointer to the next potential token in the Next output.
// If end of data is reached, or an error occured, only Out is returned.
// Collect expects all connected inputs to provide data, otherwise
// it will deadlock.
type Collect struct {
	In   [](<-chan Token)
	Next chan<- Token
	Out  chan<- Token
}

// Process loops through all connected inputs, reads a token from each of them,
// and returns a first matched token to Out. It moves the Pos pointer for the token
// to the next readable position and sends it to Next. If none of the scanners matched,
// it sends an illegal token to Out.
// Order of inputs defines preference: if both In[1] and In[5] matched, In[1] will be
// returned.
func (c *Collect) Process() {
	insCount := len(c.In)
	closedCount := 0
	for closedCount < insCount {
		var res, last Token // first matched token
		matched := false
		for i := 0; i < insCount; i++ {
			i := i
			ch := c.In[i]
			t, ok := <-ch
			if ok {
				if t.Type != tokIllegal && !matched {
					res = t
					matched = true
				} else {
					last = t
				}
			} else {
				closedCount++
			}
		}
		if closedCount == insCount {
			// It was a round of close signals
			return
		}
		if matched {
			c.Out <- res
			t := res // makes a copy
			t.Pos += len(t.Value)
			if t.Pos < len(t.File.Data) {
				c.Next <- t
			} else {
				t.Type = tokEOF
				t.Value = t.File.Name
				c.Out <- t
				// FIXME how to process multiple files in the same network and provide graceful shutdown?
				return
			}
		} else {
			// Nothing matched, it's an error
			c.Out <- last
		}
	}
}
