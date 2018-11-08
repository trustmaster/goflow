package flow

import (
	"testing"
)

// This component interface is common for many test cases
type intInAndOut struct {
	In <-chan int
	Out chan<- int
}

type doubleOnce intInAndOut

func (c *doubleOnce) Process() {
	i := <-c.In
	c.Out <- 2*i
}

// Test a simple component that runs only once
func TestSimpleComponent(t *testing.T) {
	in := make(chan int)
	out := make(chan int)
	c := &doubleOnce{
		in,
		out,
	}

	wait := Run(c)

	in <- 12
	res := <-out

	if res != 24 {
		t.Errorf("%d != %d", res, 24)
	}

	<-wait
}

type doubler intInAndOut

func (c *doubler) Process() {
	for i := range c.In {
		c.Out <- 2*i
	}
}

func TestSimpleLongRunningComponent(t *testing.T) {
	data := map[int]int{
		12: 24,
		7: 14,
		400: 800,
	}
	in := make(chan int)
	out := make(chan int)
	c := &doubler{
		in,
		out,
	}

	wait := Run(c)

	for src, expected := range data {
		in <- src
		actual := <- out

		if actual != expected {
			t.Errorf("%d != %d", actual, expected)
		}
	}

	// We have to close input for the process to finish
	close(in)
	<-wait
}