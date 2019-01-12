package flow

import (
	"testing"
)

func newRepeatGraph() (*Graph, error) {
	n := NewGraph()

	if err := n.Add("r", new(repeater)); err != nil {
		return nil, err
	}

	if err := n.MapInPort("Word", "r", "Word"); err != nil {
		return nil, err
	}
	if err := n.MapOutPort("Words", "r", "Words"); err != nil {
		return nil, err
	}

	return n, nil
}

func TestBasicIIP(t *testing.T) {
	qty := 5

	n, err := newRepeatGraph()
	if err != nil {
		t.Error(err)
		return
	}
	if err := n.AddIIP("r", "Times", qty); err != nil {
		t.Error(err)
		return
	}

	input := "hello"
	output := []string{"hello", "hello", "hello", "hello", "hello"}

	in := make(chan string)
	out := make(chan string)

	if err := n.SetInPort("Word", in); err != nil {
		t.Error(err)
		return
	}
	if err := n.SetOutPort("Words", out); err != nil {
		t.Error(err)
		return
	}

	wait := Run(n)

	go func() {
		in <- input
		close(in)
	}()

	i := 0
	for actual := range out {
		expected := output[i]
		if actual != expected {
			t.Errorf("%s != %s", actual, expected)
		}
		i++
	}
	if i != qty {
		t.Errorf("Returned %d words instead of %d", i, qty)
	}

	<-wait
}

func newRepeatGraph2Ins() (*Graph, error) {
	n := NewGraph()

	if err := n.Add("r", new(repeater)); err != nil {
		return nil, err
	}

	if err := n.MapInPort("Word", "r", "Word"); err != nil {
		return nil, err
	}
	if err := n.MapInPort("Times", "r", "Times"); err != nil {
		return nil, err
	}
	if err := n.MapOutPort("Words", "r", "Words"); err != nil {
		return nil, err
	}

	return n, nil
}

func TestGraphInportIIP(t *testing.T) {
	qty := 5

	n, err := newRepeatGraph2Ins()
	if err != nil {
		t.Error(err)
		return
	}

	input := "hello"
	output := []string{"hello", "hello", "hello", "hello", "hello"}

	in := make(chan string)
	times := make(chan int)
	out := make(chan string)

	if err := n.SetInPort("Word", in); err != nil {
		t.Error(err)
		return
	}
	if err := n.SetInPort("Times", times); err != nil {
		t.Error(err)
		return
	}
	if err := n.SetOutPort("Words", out); err != nil {
		t.Error(err)
		return
	}
	if err := n.AddIIP("r", "Times", qty); err != nil {
		t.Error(err)
		return
	}

	wait := Run(n)

	go func() {
		in <- input
		close(in)
	}()

	i := 0
	for actual := range out {
		if i == 0 {
			// The graph inport needs to be closed once the IIP is sent
			close(times)
		}
		expected := output[i]
		if actual != expected {
			t.Errorf("%s != %s", actual, expected)
		}
		i++
	}
	if i != qty {
		t.Errorf("Returned %d words instead of %d", i, qty)
	}

	<-wait
}

func TestInternalConnectionIIP(t *testing.T) {
	input := 1
	iip := 2
	output := []int{1, 2}
	qty := 2

	n, err := newDoubleEcho()
	if err != nil {
		t.Error(err)
		return
	}

	if err := n.AddIIP("e2", "In", iip); err != nil {
		t.Error(err)
		return
	}

	in := make(chan int)
	out := make(chan int)

	if err := n.SetInPort("In", in); err != nil {
		t.Error(err)
		return
	}
	if err := n.SetOutPort("Out", out); err != nil {
		t.Error(err)
		return
	}

	wait := Run(n)

	go func() {
		in <- input
		close(in)
	}()

	i := 0
	for actual := range out {
		// The order of output is not guaranteed in this case
		if actual != output[0] && actual != output[1] {
			t.Errorf("Unexpected value %d", actual)
		}
		i++
	}
	if i != qty {
		t.Errorf("Returned %d words instead of %d", i, qty)
	}

	<-wait
}

func TestAddRemoveIIP(t *testing.T) {
	n := NewGraph()

	if err := n.Add("e", new(echo)); err != nil {
		t.Error(err)
		return
	}

	if err := n.AddIIP("e", "In", 5); err != nil {
		t.Error(err)
		return
	}

	// Adding an IIP to a non-existing process/port should fail
	if err := n.AddIIP("d", "No", 404); err == nil {
		t.FailNow()
		return
	}

	if err := n.RemoveIIP("e", "In"); err != nil {
		t.Error(err)
		return
	}

	// Second attempt to remove same IIP should fail
	if err := n.RemoveIIP("e", "In"); err == nil {
		t.FailNow()
		return
	}
}
