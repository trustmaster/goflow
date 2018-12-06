package flow

import (
	"testing"
)

type repeater struct {
	Word  <-chan string
	Times <-chan int

	Words chan<- string
}

func (c *repeater) Process() {
	defer close(c.Words)
	guard := NewInputGuard(2)

	times := 0
	word := ""

	for {
		select {
		case t, ok := <-c.Times:
			if ok {
				times = t
				c.repeat(word, times)
			} else if guard.Complete() {
				return
			}
		case w, ok := <-c.Word:
			if ok {
				word = w
				c.repeat(word, times)
			} else if guard.Complete() {
				return
			}
		}
	}
}

func (c *repeater) repeat(word string, times int) {
	if word == "" || times <= 0 {
		return
	}
	for i := 0; i < times; i++ {
		c.Words <- word
	}
}

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
