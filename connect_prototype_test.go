// +build ignore

package goflow

import "testing"

func newArrayPorts() (*Graph, error) {
	n := NewGraph()

	components := map[string]interface{}{
		"e1": new(echo),
		"e2": new(echo),
		"e3": new(echo),
		"r":  new(router),
	}

	for name, c := range components {
		if err := n.Add(name, c); err != nil {
			return nil, err
		}
	}

	connections := []struct{ sn, sp, rn, rp string }{
		{"e1", "Out", "r", "In[0]"},
		{"e2", "Out", "r", "In[2]"},
		{"r", "Out[2]", "e3", "In"},
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
		{"r", "In[1]", 2},
		{"e2", "In", 3},
	}

	for _, p := range iips {
		if err := n.AddIIP(p.proc, p.port, p.v); err != nil {
			return nil, err
		}
	}

	outPorts := []struct{ pn, pp, name string }{
		{"r", "Out[0]", "O1"},
		{"r", "Out[1]", "O2"},
		{"e3", "Out", "O3"},
	}

	for _, p := range outPorts {
		if err := n.MapOutPort(p.pn, p.pp, p.name); err != nil {
			return nil, err
		}
	}

	return n, nil
}

func TestArrayPorts(t *testing.T) {
	n, err := newArrayPorts()
	if err != nil {
		t.Error(err)
		return
	}

	o1 := make(chan int)
	o2 := make(chan int)
	o3 := make(chan int)
	n.SetOutPort("O1", o1)
	n.SetOutPort("O2", o2)
	n.SetOutPort("O3", o3)

	wait := Run(n)

	v1 := <-o1
	v2 := <-o2
	v3 := <-o3

	expected := []int{3, 2, 1}
	actual := []int{v1, v2, v3}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %d, got %d", expected[i], v)
		}
	}

	<-wait
}
