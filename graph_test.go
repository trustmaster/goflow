package flow

import (
	"testing"
)

type echo struct {
	In  <-chan int
	Out chan<- int
}

func (c *echo) Process() {
	for i := range c.In {
		c.Out <- i
	}
	close(c.Out)
}

func newDoubleEcho() (*Graph, error) {
	n := NewGraph()
	// Components
	e1 := new(echo)
	e2 := new(echo)
	// Structure
	if err := n.Add("e1", e1); err != nil {
		return nil, err
	}
	if err := n.Add("e2", e2); err != nil {
		return nil, err
	}
	if err := n.Connect("e1", "Out", "e2", "In"); err != nil {
		return nil, err
	}
	// Ports
	if err := n.MapInPort("In", "e1", "In"); err != nil {
		return nil, err
	}
	if err := n.MapOutPort("Out", "e2", "Out"); err != nil {
		return nil, err
	}
	return n, nil
}

func TestSimpleGraph(t *testing.T) {
	data := []int{7, 97, 16, 356, 81}

	n, err := newDoubleEcho()
	if err != nil {
		t.Error(err)
		return
	}

	in := make(chan int)
	out := make(chan int)
	n.SetInPort("In", in)
	n.SetOutPort("Out", out)

	wait := Run(n)

	go func() {
		for _, n := range data {
			in <- n
		}
		close(in)
	}()

	i := 0
	for actual := range out {
		expected := data[i]
		if actual != expected {
			t.Errorf("%d != %d", actual, expected)
		}
		i++
	}

	<-wait
}
