package flow

import (
	"testing"
)

func newRepeat5Times() (*Graph, error) {
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

	if err := n.AddIIP("r", "Times", 5); err != nil {
		return nil, err
	}

	return n, nil
}

func TestBasicIIP(t *testing.T) {
	input := "hello"
	output := []string{"hello", "hello", "hello", "hello", "hello"}

	n, err := newRepeat5Times()
	if err != nil {
		t.Error(err)
		return
	}

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
	if i != 5 {
		t.Errorf("Returned %d words instead of 5", i)
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
