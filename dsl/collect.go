package dsl

import (
	"sync"
)

// Collect gathers tokens processed in parallel, returns a matched one,
// and moves pointer to the next potential token
type Collect struct {
	In   [](<-chan Token)
	Next chan<- Token
	Out  chan<- Token
}

// Process loops through all connected inputs, reads a token from each of them,
// and returns a first matched token to Out. It moves the Pos pointer for the token
// to the next readable position and sends it to Next. If none of the scanners matched,
// it sends an illegal token to Out.
func (c *Collect) Process() {
	insCount := len(c.In)
	closedCount := 0
	for closedCount < insCount {
		var res, last Token // first matched token
		matched := false
		wg := new(sync.WaitGroup)
		m := new(sync.Mutex)
		for i := 0; i < insCount; i++ {
			i := i
			wg.Add(1)
			go func() {
				ch := c.In[i]
				t, ok := <-ch
				m.Lock()
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
				m.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
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
			}
		} else {
			// Nothing matched, it's an error
			c.Out <- last
		}
	}
}
