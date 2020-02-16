package goflow

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
			"connect: getProcPort: process 'noproc' not found",
		},
		{
			"Invalid receiver port",
			n.Connect("e1", "Out", "e2", "NotIn"),
			"connect: getProcPort: process 'e2' does not have port 'NotIn'",
		},
		{
			"Invalid sender proc",
			n.Connect("noproc", "Out", "e2", "In"),
			"connect: getProcPort: process 'noproc' not found",
		},
		{
			"Invalid sender port",
			n.Connect("e1", "NotOut", "e2", "In"),
			"connect: getProcPort: process 'e1' does not have port 'NotOut'",
		},
		{
			"Sending to output",
			n.Connect("e1", "Out", "e2", "Out"),
			"connect: validation of 'e2.Out' failed: channel does not support direction <-chan",
		},
		{
			"Sending from input",
			n.Connect("e1", "In", "e2", "In"),
			"connect: validation of 'e1.In' failed: channel does not support direction chan<-",
		},
		{
			"Connecting to non-chan",
			n.Connect("e1", "Out", "inv", "NotChan"),
			"connect: validation of 'inv.NotChan' failed: not a channel",
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

func newFanOutFanIn() (*Graph, error) {
	n := NewGraph()

	components := map[string]interface{}{
		"e1": new(echo),
		"d1": new(doubler),
		"d2": new(doubler),
		"d3": new(doubler),
		"e2": new(echo),
	}

	for name, c := range components {
		if err := n.Add(name, c); err != nil {
			return nil, err
		}
	}

	connections := []struct{ sn, sp, rn, rp string }{
		{"e1", "Out", "d1", "In"},
		{"e1", "Out", "d2", "In"},
		{"e1", "Out", "d3", "In"},
		{"d1", "Out", "e2", "In"},
		{"d2", "Out", "e2", "In"},
		{"d3", "Out", "e2", "In"},
	}

	for _, c := range connections {
		if err := n.Connect(c.sn, c.sp, c.rn, c.rp); err != nil {
			return nil, err
		}
	}

	if err := n.MapInPort("In", "e1", "In"); err != nil {
		return nil, err
	}

	if err := n.MapOutPort("Out", "e2", "Out"); err != nil {
		return nil, err
	}

	return n, nil
}

func TestFanOutFanIn(t *testing.T) {
	inData := []int{1, 2, 3, 4, 5, 6, 7, 8}
	outData := []int{2, 4, 6, 8, 10, 12, 14, 16}

	n, err := newFanOutFanIn()
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
		for _, n := range inData {
			in <- n
		}
		close(in)
	}()

	i := 0
	for actual := range out {
		found := false
		for j := 0; j < len(outData); j++ {
			if outData[j] == actual {
				found = true
				outData = append(outData[:j], outData[j+1:]...)
			}
		}
		if !found {
			t.Errorf("%d not found in expected data", actual)
		}
		i++
	}

	if i != len(inData) {
		t.Errorf("Output count missmatch: %d != %d", i, len(inData))
	}

	<-wait
}
