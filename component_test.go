package goflow

import (
	"testing"
)

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

// Test a simple long running component with one input
func TestSimpleLongRunningComponent(t *testing.T) {
	data := map[int]int{
		12:  24,
		7:   14,
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

		actual := <-out

		if actual != expected {
			t.Errorf("%d != %d", actual, expected)
		}
	}

	// We have to close input for the process to finish
	close(in)
	<-wait
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

	for i := 0; i < len(sums); i++ {
		actual := <-out
		expected := sums[i]

		if actual != expected {
			t.Errorf("%d != %d", actual, expected)
		}
	}

	<-wait
}
