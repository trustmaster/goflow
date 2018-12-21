package flow

import (
	"testing"
)

type withInvalidPorts struct {
	NotChan int
	Chan    <-chan int
}

func (c *withInvalidPorts) Process() {
	// Dummy
}

func TestConnectInvalidParams(t *testing.T) {
	n := NewGraph()

	n.Add("e1", new(echo))
	n.Add("e2", new(echo))
	n.Add("inv", new(withInvalidPorts))

	cases := []struct {
		scenario string
		err      error
		msg      string
	}{
		{
			"Invalid receiver proc",
			n.Connect("e1", "Out", "noproc", "In"),
			"Connect error: process 'noproc' not found",
		},
		{
			"Invalid receiver port",
			n.Connect("e1", "Out", "e2", "NotIn"),
			"Connect error: process 'e2' does not have port 'NotIn'",
		},
		{
			"Invalid sender proc",
			n.Connect("noproc", "Out", "e2", "In"),
			"Connect error: process 'noproc' not found",
		},
		{
			"Invalid sender port",
			n.Connect("e1", "NotOut", "e2", "In"),
			"Connect error: process 'e1' does not have port 'NotOut'",
		},
		{
			"Sending to output",
			n.Connect("e1", "Out", "e2", "Out"),
			"Connect error: 'e2.Out' is not of the correct chan type",
		},
		{
			"Sending from input",
			n.Connect("e1", "In", "e2", "In"),
			"Connect error: 'e1.In' is not of the correct chan type",
		},
		{
			"Connecting to non-chan",
			n.Connect("e1", "Out", "inv", "NotChan"),
			"Connect error: 'inv.NotChan' is not of the correct chan type",
		},
	}

	for _, item := range cases {
		c := item
		t.Run(c.scenario, func(t *testing.T) {
			t.Parallel()
			if c.err == nil {
				t.Fail()
			} else if c.msg != c.err.Error() {
				t.Error(c.err)
			}
		})
	}
}

func TestSubgraphSender(t *testing.T) {
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

	if err := n.Connect("sub", "Out", "e3", "In"); err != nil {
		t.Error(err)
		return
	}

	n.MapInPort("In", "sub", "In")
	n.MapOutPort("Out", "e3", "Out")

	testGraphWithNumberSequence(n, t)
}

func TestSubgraphReceiver(t *testing.T) {
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

	if err := n.Connect("e3", "Out", "sub", "In"); err != nil {
		t.Error(err)
		return
	}

	n.MapInPort("In", "e3", "In")
	n.MapOutPort("Out", "sub", "Out")

	testGraphWithNumberSequence(n, t)
}
