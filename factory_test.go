package flow

import (
	"testing"
)

// Tests run-time process creating with flow.Factory
func TestFactory(t *testing.T) {
	procs := make(map[string]interface{})
	procs["e1"] = Factory("echoer")
	in, out := make(chan int), make(chan int)
	e1 := procs["e1"].(*echoer)
	e1.In = in
	e1.Out = out
	RunProc(procs["e1"])
	for i := 0; i < 10; i++ {
		in <- i
		i2 := <-out
		if i != i2 {
			t.Errorf("%d != %d", i2, i)
		}
	}
	// Shutdown proc
	close(in)
}

type dummyNet struct {
	Graph
}

// Tests connection between 2 processes created at run-time
func TestFactoryConnection(t *testing.T) {
	net := new(dummyNet)
	net.InitGraphState()

	net.AddNew("echoer", "e1")
	net.AddNew("echoer", "e2")

	net.Connect("e1", "Out", "e2", "In")

	net.MapInPort("In", "e1", "In")
	net.MapOutPort("Out", "e2", "Out")

	in, out := make(chan int), make(chan int)

	net.SetInPort("In", in)
	net.SetOutPort("Out", out)

	RunNet(net)

	in <- 123
	i := <-out

	close(in)

	if i != 123 {
		t.Errorf("Error: %d != 123", i)
	}
}

// Creates a graph that will be loaded at run-time
func newDummyNet() interface{} {
	net := new(dummyNet)
	net.InitGraphState()

	net.AddNew("echoer", "e1")
	net.AddNew("echoer", "e2")

	net.Connect("e1", "Out", "e2", "In")

	net.MapInPort("In", "e1", "In")
	net.MapOutPort("Out", "e2", "Out")

	return net
}

func init() {
	Register("DummyNet", newDummyNet)
}

// Tests adding subgraph components at run-time
func TestFactorySubgraph(t *testing.T) {
	net := new(dummyNet)
	net.InitGraphState()

	net.AddNew("DummyNet", "d1")

	net.MapInPort("In", "d1", "In")
	net.MapOutPort("Out", "d1", "Out")

	in, out := make(chan int), make(chan int)

	net.SetInPort("In", in)
	net.SetOutPort("Out", out)

	RunNet(net)

	in <- 123
	i := <-out

	close(in)

	if i != 123 {
		t.Errorf("Error: %d != 123", i)
	}
}
