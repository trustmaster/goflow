package goflow

import (
	"fmt"
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
	n.MapInPort("In", "e1", "In")
	n.MapOutPort("Out", "e2", "Out")

	return n, nil
}

func TestSimpleGraph(t *testing.T) {
	n, err := newDoubleEcho()
	if err != nil {
		t.Error(err)
		return
	}

	testGraphWithNumberSequence(n, t)
}

func testGraphWithNumberSequence(n *Graph, t *testing.T) {
	data := []int{7, 97, 16, 356, 81}

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

// TestFanInCloseRace verifies that when multiple senders share an output
// channel (fan-in), no data race or double-close panic occurs when their
// closeProcOuts goroutines run concurrently. Regression test for the
// channel close race guarded by Graph.closeChan.
func TestFanInCloseRace(t *testing.T) {
	const numSenders = 30

	n := NewGraph()

	// Create senders
	senders := make([]*singleShot, numSenders)
	for i := range numSenders {
		senders[i] = new(singleShot)
		if err := n.Add(fmt.Sprintf("s%d", i), senders[i]); err != nil {
			t.Fatal(err)
		}
	}

	// Create a single receiver (echo reads until its input closes)
	r := new(echo)
	if err := n.Add("r", r); err != nil {
		t.Fatal(err)
	}

	// Fan-in: connect all sender outputs to the same receiver input
	for i := range numSenders {
		if err := n.Connect(fmt.Sprintf("s%d", i), "Out", "r", "In"); err != nil {
			t.Fatal(err)
		}
	}

	// Map inports
	for i := range numSenders {
		n.MapInPort(fmt.Sprintf("In%d", i), fmt.Sprintf("s%d", i), "In")
	}
	n.MapOutPort("Out", "r", "Out")

	// Set up external channels
	inChans := make([]chan int, numSenders)
	for i := range numSenders {
		inChans[i] = make(chan int)
		if err := n.SetInPort(fmt.Sprintf("In%d", i), inChans[i]); err != nil {
			t.Fatal(err)
		}
	}

	out := make(chan int, numSenders)
	if err := n.SetOutPort("Out", out); err != nil {
		t.Fatal(err)
	}

	wait := Run(n)

	// Send one value to each sender and close each input
	for i := range numSenders {
		inChans[i] <- i
		close(inChans[i])
	}

	// Read all results
	count := 0
	for range out {
		count++
	}

	if count != numSenders {
		t.Errorf("expected %d values, got %d", numSenders, count)
	}

	<-wait
}

// TestIIPAndCloseOutRace verifies that when an IIP sends on a channel that
// is also an output port of another process, no double-close race occurs.
// The IIP goroutine and closeProcOuts both target the same underlying channel.
func TestIIPAndCloseOutRace(t *testing.T) {
	n := NewGraph()

	// A sender component
	s := new(singleShot)
	if err := n.Add("s", s); err != nil {
		t.Fatal(err)
	}

	// Receiver
	r := new(echo)
	if err := n.Add("r", r); err != nil {
		t.Fatal(err)
	}

	// Connect sender to receiver
	if err := n.Connect("s", "Out", "r", "In"); err != nil {
		t.Fatal(err)
	}

	// Add IIP to the same receiver input port (same underlying channel)
	if err := n.AddIIP("r", "In", 99); err != nil {
		t.Fatal(err)
	}

	// Map ports
	n.MapInPort("In", "s", "In")
	n.MapOutPort("Out", "r", "Out")

	in := make(chan int)
	out := make(chan int, 2)

	if err := n.SetInPort("In", in); err != nil {
		t.Fatal(err)
	}

	if err := n.SetOutPort("Out", out); err != nil {
		t.Fatal(err)
	}

	wait := Run(n)

	// Send to the sender
	in <- 42
	close(in)

	count := 0
	for range out {
		count++
	}

	// Expect: 42 from sender + 99 from IIP
	if count != 2 {
		t.Errorf("expected 2 values, got %d", count)
	}

	<-wait
}

func RegisterTestGraph(f *Factory) error {
	f.Register("doubleEcho", func() (any, error) {
		return newDoubleEcho()
	})

	f.Annotate("doubleEcho", Annotation{
		Description: "Contains a chain of two echo components",
	})

	return nil
}
