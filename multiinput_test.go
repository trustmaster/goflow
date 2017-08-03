package flow

import (
	"fmt"
	"testing"
	"time"
)

// Test taken from https://gist.github.com/lovromazgon/ff8432017ab312c5fb18a1b7f377c773

// component 1
type C1 struct {
	Component
	In  <-chan int
	Out chan<- int
}

func (c *C1) OnIn(i int) {
	c.Out <- i
}

func (c *C1) Finish() {
	fmt.Println("Finish c1")
}

// component 2
type C2 struct {
	Component
	In1 <-chan int
	In2 <-chan int
}

func (c *C2) OnIn1(i int) {
	fmt.Printf("in1: %d\n", i)
}

func (c *C2) OnIn2(i int) {
	fmt.Printf("in2: %d\n", i)
}

func (c *C2) Finish() {
	fmt.Println("Finish c2")
}

// test
func TestMultiInput(t *testing.T) {
	DefaultComponentMode = ComponentModeSync
	n := new(Graph)    // creates the object in heap
	n.InitGraphState() // allocates memory for the graph

	// Add processes to the network
	n.Add(new(C1), "c1-1")
	n.Add(new(C1), "c1-2")
	n.Add(new(C2), "c2")

	n.Connect("c1-1", "Out", "c2", "In1")
	n.Connect("c1-2", "Out", "c2", "In2")

	n.MapInPort("In1", "c1-1", "In")
	n.MapInPort("In2", "c1-2", "In")

	in1 := make(chan int)
	in2 := make(chan int)
	n.SetInPort("In1", in1)
	n.SetInPort("In2", in2)

	RunNet(n)

	in1 <- 1
	in2 <- 2
	in1 <- 3
	in2 <- 4
	in1 <- 5

	time.Sleep(time.Second)
	close(in1)
	time.Sleep(time.Second)
	close(in2)

	select {
	case <-n.Wait():
		t.Log("Success")
	case <-time.NewTimer(time.Second * 3).C:
		t.Log("Waited 3 seconds for close!")
		t.Fail()
	}
}
