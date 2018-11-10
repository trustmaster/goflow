package flow

import (
	"testing"
)

// This component interface is common for many test cases
type intInAndOut struct {
	In <-chan int
	Out chan<- int
}

type doubleOnce intInAndOut

func (c *doubleOnce) Process() {
	i := <-c.In
	c.Out <- 2*i
}

// Test a simple component that runs only once
func TestSimpleComponent(t *testing.T) {
	in := make(chan int)
	out := make(chan int)
	c := &doubleOnce{
		in,
		out,
	}

	wait := Run(c)

	in <- 12
	res := <-out

	if res != 24 {
		t.Errorf("%d != %d", res, 24)
	}

	<-wait
}

type doubler intInAndOut

func (c *doubler) Process() {
	for i := range c.In {
		c.Out <- 2*i
	}
}

// Test a simple long running component with one input
func TestSimpleLongRunningComponent(t *testing.T) {
	data := map[int]int{
		12: 24,
		7: 14,
		400: 800,
	}
	in := make(chan int)
	out := make(chan int)
	c := &doubler{
		in,
		out,
	}

	wait := Run(c)

	for src, expected := range data {
		in <- src
		actual := <- out

		if actual != expected {
			t.Errorf("%d != %d", actual, expected)
		}
	}

	// We have to close input for the process to finish
	close(in)
	<-wait
}

// A component with two inputs and one output
type adder struct {
	Op1 <-chan int
	Op2 <-chan int
	Sum chan<- int
}

func (c *adder) Process() {
	op1Buf := make([]int, 0, 10)
	op2Buf := make([]int, 0, 10)
	addOp := func(op int, buf, otherBuf *[]int) {
		if len(*otherBuf) > 0 {
			otherOp := (*otherBuf)[0]
			*otherBuf = (*otherBuf)[1:]
			c.Sum <- (op + otherOp)
		} else {
			*buf = append(*buf, op)
		}
	}

	const inputCount = 2
	closed := make([]struct{}, 0, inputCount)
	completed := func() bool {
		closed = append(closed, struct{}{})
		return len(closed) >= inputCount
	}

	closeOuts := func() {
		close(c.Sum)
	}

	defer closeOuts()

	for {
		select {
		case op1, ok := <-c.Op1:
			if ok {
				addOp(op1, &op1Buf, &op2Buf)
			} else if completed() {
				return
			}
			
		case op2, ok := <-c.Op2:
			if ok {
				addOp(op2, &op2Buf, &op1Buf)
			} else if completed() {
				return
			}
		}
	}
}

func TestComponentWithTwoInputs(t *testing.T) {
	op1 := []int{3, 5, 92, 28}
	op2 := []int{38, 94, 4, 9}
	sums := []int{41, 99, 96, 37}

	in1 := make(chan int)
	in2 := make(chan int)
	out := make(chan int)
	c := &adder{in1, in2, out}

	wait := Run(c)

	go func() {
		for _, n := range op1 {
			in1 <- n
		}
		close(in1)
	}()

	go func() {
		for _, n := range op2 {
			in2 <- n
		}
		close(in2)
	}()
	
	i := 0
	for actual := range out {
		expected := sums[i]
		if actual != expected {
			t.Errorf("%d != %d", actual, expected)
		}
		i++
	}
	
	<-wait
}