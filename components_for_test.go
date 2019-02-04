package goflow

import (
	"fmt"
	"sync"
)

// doubler doubles its input
type doubler struct {
	In  <-chan int
	Out chan<- int
}

func (c *doubler) Process() {
	for i := range c.In {
		c.Out <- 2 * i
	}
}

// doubleOnce is a non-resident version of doubler
type doubleOnce struct {
	In  <-chan int
	Out chan<- int
}

func (c *doubleOnce) Process() {
	i := <-c.In
	c.Out <- 2 * i
}

// A component with two inputs and one output
type adder struct {
	Op1 <-chan int
	Op2 <-chan int
	Sum chan<- int
}

func (c *adder) Process() {
	guard := NewInputGuard("op1", "op2")

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

	for {
		select {
		case op1, ok := <-c.Op1:
			if ok {
				addOp(op1, &op1Buf, &op2Buf)
			} else if guard.Complete("op1") {
				return
			}

		case op2, ok := <-c.Op2:
			if ok {
				addOp(op2, &op2Buf, &op1Buf)
			} else if guard.Complete("op2") {
				return
			}
		}
	}
}

// echo passes input to the output
type echo struct {
	In  <-chan int
	Out chan<- int
}

func (c *echo) Process() {
	for i := range c.In {
		c.Out <- i
	}
}

// repeater repeats an input string a given number of times
type repeater struct {
	Word  <-chan string
	Times <-chan int

	Words chan<- string
}

func (c *repeater) Process() {
	guard := NewInputGuard("word", "times")

	times := 0
	word := ""

	for {
		select {
		case t, ok := <-c.Times:
			if ok {
				times = t
				c.repeat(word, times)
			} else if guard.Complete("times") {
				return
			}
		case w, ok := <-c.Word:
			if ok {
				word = w
				c.repeat(word, times)
			} else if guard.Complete("word") {
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

// router routes input map port to output
type router struct {
	In  map[string]<-chan int
	Out map[string]chan<- int
}

// Process routes incoming packets to the output by sending them to the same
// outport key as the inport key they arrived at
func (c *router) Process() {
	wg := new(sync.WaitGroup)
	for k, ch := range c.In {
		k := k
		ch := ch
		wg.Add(1)
		go func() {
			for n := range ch {
				c.Out[k] <- n
			}
			close(c.Out[k])
			wg.Done()
		}()
	}
	wg.Wait()
}

// irouter routes input array port to output
type irouter struct {
	In  [](<-chan int)
	Out [](chan<- int)
}

// Process routes incoming packets to the output by sending them to the same
// outport key as the inport key they arrived at
func (c *irouter) Process() {
	wg := new(sync.WaitGroup)
	for k, ch := range c.In {
		k := k
		ch := ch
		wg.Add(1)
		go func() {
			for n := range ch {
				c.Out[k] <- n
			}
			close(c.Out[k])
			wg.Done()
		}()
	}
	wg.Wait()
}

func RegisterTestComponents(f *Factory) error {
	f.Register("echo", func() (interface{}, error) {
		return new(echo), nil
	})
	f.Annotate("echo", Annotation{
		Description: "Passes an int from in to out without changing it",
		Icon:        "arrow-right",
	})
	f.Register("doubler", func() (interface{}, error) {
		return new(doubler), nil
	})
	f.Annotate("doubler", Annotation{
		Description: "Doubles its input",
		Icon:        "times-circle",
	})
	f.Register("repeater", func() (interface{}, error) {
		return new(repeater), nil
	})
	f.Annotate("repeater", Annotation{
		Description: "Repeats Word given numer of Times",
		Icon:        "times-circle",
	})
	f.Register("adder", func() (interface{}, error) {
		return new(adder), nil
	})
	f.Annotate("adder", Annotation{
		Description: "Sums integers coming to its inports",
		Icon:        "plus-circle",
	})
	return nil
}

// pipeline allows chaining simple calls in tests
type pipeline struct {
	err error
}

// ok asserts that a function does not return an error
func (p *pipeline) ok(f func() error) *pipeline {
	if p.err != nil {
		return p
	}

	p.err = f()
	return p
}

// fails asserts that a function returns an error
func (p *pipeline) fails(f func() error) *pipeline {
	if p.err != nil {
		return p
	}

	err := f()
	if err == nil {
		p.err = fmt.Errorf("Expected an error")
	}
	return p
}
