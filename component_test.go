package flow

import (
	"testing"
)

type doubleOnce struct {
	In <-chan int
	Out chan<- int
}

func (c *doubleOnce) Process() {
	i := <-c.In
	c.Out <- 2*i
}

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

