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
			"connect: getProcPort: process 'e2' does not have a valid port 'NotIn'",
		},
		{
			"Invalid sender proc",
			n.Connect("noproc", "Out", "e2", "In"),
			"connect: getProcPort: process 'noproc' not found",
		},
		{
			"Invalid sender port",
			n.Connect("e1", "NotOut", "e2", "In"),
			"connect: getProcPort: process 'e1' does not have a valid port 'NotOut'",
		},
		{
			"Sending to output",
			n.Connect("e1", "Out", "e2", "Out"),
			"connect 'e2.Out': channel does not support direction <-chan",
		},
		{
			"Sending from input",
			n.Connect("e1", "In", "e2", "In"),
			"connect 'e1.In': channel does not support direction chan<-",
		},
		{
			"Connecting to non-chan",
			n.Connect("e1", "Out", "inv", "NotChan"),
			"connect 'inv.NotChan': not a channel",
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

	n.MapInPort("In", "e1", "In")
	n.MapOutPort("Out", "e2", "Out")

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

func newMapPorts() (*Graph, error) {
	n := NewGraph()

	components := map[string]interface{}{
		"e1":  new(echo),
		"e11": new(echo),
		"e22": new(echo),
		"r":   new(router),
	}

	for name, c := range components {
		if err := n.Add(name, c); err != nil {
			return nil, err
		}
	}

	connections := []struct{ sn, sp, rn, rp string }{
		{"e1", "Out", "r", "In[e1]"},
		{"r", "Out[e2]", "e22", "In"},
		{"r", "Out[e1]", "e11", "In"},
	}

	for _, c := range connections {
		if err := n.Connect(c.sn, c.sp, c.rn, c.rp); err != nil {
			return nil, err
		}
	}

	iips := []struct {
		proc, port string
		v          int
	}{
		{"e1", "In", 1},
		{"r", "In[e3]", 3},
	}

	for _, p := range iips {
		if err := n.AddIIP(p.proc, p.port, p.v); err != nil {
			return nil, err
		}
	}

	n.MapInPort("I2", "r", "In[e2]")

	outPorts := []struct{ pn, pp, name string }{
		{"e11", "Out", "O1"},
		{"e22", "Out", "O2"},
		{"r", "Out[e3]", "O3"},
	}

	for _, p := range outPorts {
		n.MapOutPort(p.name, p.pn, p.pp)
	}

	return n, nil
}

func TestMapPorts(t *testing.T) {
	n, err := newMapPorts()
	if err != nil {
		t.Error(err)
		return
	}

	i2 := make(chan int, 1)
	o1 := make(chan int)
	o2 := make(chan int)
	o3 := make(chan int)
	if err := n.SetInPort("I2", i2); err != nil {
		t.Error(err)
		return
	}
	n.SetOutPort("O1", o1)
	n.SetOutPort("O2", o2)
	if err := n.SetOutPort("O3", o3); err != nil {
		t.Error(err)
		return
	}

	wait := Run(n)

	i2 <- 2
	close(i2)
	v1 := <-o1
	v2 := <-o2
	v3 := <-o3

	expected := []int{1, 2, 3}
	actual := []int{v1, v2, v3}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %d, got %d", expected[i], v)
		}
	}

	<-wait
}

func newArrayPorts() (*Graph, error) {
	n := NewGraph()

	components := map[string]interface{}{
		"e0":  new(echo),
		"e00": new(echo),
		"e11": new(echo),
		"r":   new(irouter),
	}

	for name, c := range components {
		if err := n.Add(name, c); err != nil {
			return nil, err
		}
	}

	connections := []struct{ sn, sp, rn, rp string }{
		{"e0", "Out", "r", "In[0]"},
		{"r", "Out[1]", "e11", "In"},
		{"r", "Out[0]", "e00", "In"},
	}

	for _, c := range connections {
		if err := n.Connect(c.sn, c.sp, c.rn, c.rp); err != nil {
			return nil, err
		}
	}

	iips := []struct {
		proc, port string
		v          int
	}{
		{"e0", "In", 1},
		{"r", "In[2]", 3},
	}

	for _, p := range iips {
		if err := n.AddIIP(p.proc, p.port, p.v); err != nil {
			return nil, err
		}
	}

	n.MapInPort("I1", "r", "In[1]")

	outPorts := []struct{ pn, pp, name string }{
		{"e00", "Out", "O0"},
		{"e11", "Out", "O1"},
		{"r", "Out[2]", "O2"},
	}

	for _, p := range outPorts {
		n.MapOutPort(p.name, p.pn, p.pp)
	}

	return n, nil
}

func TestArrayPorts(t *testing.T) {
	n, err := newArrayPorts()
	if err != nil {
		t.Error(err)
		return
	}

	i1 := make(chan int, 1)
	o0 := make(chan int)
	o1 := make(chan int)
	o2 := make(chan int)
	if err := n.SetInPort("I1", i1); err != nil {
		t.Error(err)
		return
	}
	n.SetOutPort("O0", o0)
	n.SetOutPort("O1", o1)
	if err := n.SetOutPort("O2", o2); err != nil {
		t.Error(err)
		return
	}

	wait := Run(n)

	i1 <- 2
	close(i1)
	v0 := <-o0
	v1 := <-o1
	v2 := <-o2

	expected := []int{1, 2, 3}
	actual := []int{v0, v1, v2}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %d, got %d", expected[i], v)
		}
	}

	<-wait
}
