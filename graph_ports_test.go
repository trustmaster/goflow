package goflow

import "testing"

func TestOutportNotFound(t *testing.T) {
	sub, err := newDoubleEcho()
	if err != nil {
		t.Error(err)
		return
	}

	n := NewGraph()
	if err := n.Add("sub", sub); err != nil {
		t.Error(err)
		return
	}
	n.Add("e3", new(echo))

	if err := n.Connect("sub", "NoOut", "e3", "In"); err == nil {
		t.Errorf("Expected an error")
	}
}

func TestInPortNotFound(t *testing.T) {
	sub, err := newDoubleEcho()
	if err != nil {
		t.Error(err)
		return
	}

	n := NewGraph()
	if err := n.Add("sub", sub); err != nil {
		t.Error(err)
		return
	}
	n.Add("e3", new(echo))

	if err := n.Connect("e3", "Out", "sub", "NotIn"); err == nil {
		t.Errorf("Expected an error")
	}
}

func TestSetMissingProcPorts(t *testing.T) {
	n := NewGraph()

	if err := n.Add("e1", new(echo)); err != nil {
		t.Error(err)
		return
	}

	n.MapInPort("In", "nope", "In")
	n.MapOutPort("Out", "nope", "Out")

	if err := n.SetInPort("In", make(chan int)); err == nil {
		t.Errorf("Expected an error")
		return
	}

	if err := n.SetOutPort("Out", make(chan int)); err == nil {
		t.Errorf("Expected an error")
		return
	}
}
