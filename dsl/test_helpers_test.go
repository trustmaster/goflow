package dsl_test

import (
	"sync"

	"github.com/trustmaster/goflow"
)

// testReceiver passes int values from In to Out.
type testReceiver struct {
	In  <-chan int
	Out chan<- int
}

func (c *testReceiver) Process() {
	for v := range c.In {
		c.Out <- v
	}
}

// testArrayReceiver reads from an array input port and forwards values to Out.
type testArrayReceiver struct {
	In  [](<-chan int)
	Out chan<- int
}

func (c *testArrayReceiver) Process() {
	var wg sync.WaitGroup

	for _, ch := range c.In {
		wg.Add(1)

		go func(ch <-chan int) {
			defer wg.Done()

			for v := range ch {
				c.Out <- v
			}
		}(ch)
	}

	wg.Wait()
}

// testSender sends a fixed value on Out and exits.
type testSender struct {
	Out chan<- int
}

func (c *testSender) Process() {
	c.Out <- 42
}

// registerTestComponents registers minimal components needed for tests.
func registerTestComponents(f *goflow.Factory) error {
	comps := []struct {
		name string
		ctor func() (interface{}, error)
	}{
		{"test/receiver", func() (interface{}, error) { return new(testReceiver), nil }},
		{"test/arrayReceiver", func() (interface{}, error) { return new(testArrayReceiver), nil }},
		{"test/sender", func() (interface{}, error) { return new(testSender), nil }},
	}

	for _, c := range comps {
		if err := f.Register(c.name, c.ctor); err != nil {
			return err
		}
	}

	return nil
}
