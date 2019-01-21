package goflow

import (
	"testing"
)

func TestFactoryCreate(t *testing.T) {
	f := NewFactory()
	err := RegisterTestComponents(f)
	if err != nil {
		t.Error(err)
		return
	}

	instance, err := f.Create("echo")
	if err != nil {
		t.Error(err)
		return
	}
	c, ok := instance.(Component)
	if !ok {
		t.Errorf("%+v is not a Component", c)
		return
	}

	_, err = f.Create("notfound")
	if err == nil {
		t.Errorf("Expected an error")
	}
}

func TestFactoryGraph(t *testing.T) {
	f := NewFactory()
	err := RegisterTestComponents(f)
	if err != nil {
		t.Error(err)
		return
	}
	err = RegisterTestGraph(f)
	if err != nil {
		t.Error(err)
		return
	}

	n := NewGraph()

	if err = n.AddNew("de", "doubleEcho", f); err != nil {
		t.Error(err)
		return
	}
	if err = n.AddNew("e", "echo", f); err != nil {
		t.Error(err)
		return
	}

	if err = n.Connect("de", "Out", "e", "In"); err != nil {
		t.Error(err)
		return
	}

	if err = n.MapInPort("In", "de", "In"); err != nil {
		t.Error(err)
		return
	}
	if err = n.MapOutPort("Out", "e", "Out"); err != nil {
		t.Error(err)
		return
	}

	testGraphWithNumberSequence(n, t)
}
