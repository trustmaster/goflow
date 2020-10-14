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

func TestFactoryRegistration(t *testing.T) {
	f := NewFactory(FactoryConfig{
		RegistryCapacity: 10,
	})

	if err := RegisterTestComponents(f); err != nil {
		t.Error(err)
		return
	}

	err := f.Register("echo", func() (interface{}, error) {
		return new(echo), nil
	})
	if err == nil {
		t.Errorf("Expected an error")
		return
	}

	err = f.Annotate("notfound", Annotation{})
	if err == nil {
		t.Errorf("Expected an error")
		return
	}

	err = f.Unregister("echo")
	if err != nil {
		t.Error(err)
		return
	}

	err = f.Unregister("echo")
	if err == nil {
		t.Errorf("Expected an error")
		return
	}
}

func TestFactoryGraph(t *testing.T) {
	f := NewFactory()

	if err := RegisterTestComponents(f); err != nil {
		t.Error(err)
		return
	}

	if err := RegisterTestGraph(f); err != nil {
		t.Error(err)
		return
	}

	n := NewGraph()

	if err := n.AddNew("de", "doubleEcho", f); err != nil {
		t.Error(err)
		return
	}

	if err := n.AddNew("e", "echo", f); err != nil {
		t.Error(err)
		return
	}

	if err := n.AddNew("notfound", "notfound", f); err == nil {
		t.Errorf("Expected an error")
		return
	}

	if err := n.Connect("de", "Out", "e", "In"); err != nil {
		t.Error(err)
		return
	}

	n.MapInPort("In", "de", "In")
	n.MapOutPort("Out", "e", "Out")

	testGraphWithNumberSequence(n, t)
}
