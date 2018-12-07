package flow

import (
	"testing"
)

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

func TestAddInvalidProcess(t *testing.T) {
	s := struct{ Name string }{"This is not a Component"}
	n := NewGraph()
	err := n.Add("wrong", s)
	if err == nil {
		t.Errorf("Expected an error")
	}
}

func TestRemove(t *testing.T) {
	n := NewGraph()
	e1 := new(echo)
	if err := n.Add("e1", e1); err != nil {
		t.Error(err)
		return
	}
	if err := n.Remove("e1"); err != nil {
		t.Error(err)
		return
	}
	if err := n.Remove("e2"); err == nil {
		t.Errorf("Expected an error")
		return
	}
}
