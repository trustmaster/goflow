package goflow

import (
	"sync"
)

// doubler doubles its input.
type doubler struct {
	In  <-chan int
	Out chan<- int
}

func (c *doubler) Process() {
	for i := range c.In {
		c.Out <- 2 * i
	}
}

// doubleOnce is a non-resident version of doubler.
type doubleOnce struct {
	In  <-chan int
	Out chan<- int
}

func (c *doubleOnce) Process() {
	i := <-c.In
	c.Out <- 2 * i
}

// A component with two inputs and one output.
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

	for c.Op1 != nil || c.Op2 != nil {
		select {
		case op1, ok := <-c.Op1:
			if !ok {
				c.Op1 = nil
				break
			}

			addOp(op1, &op1Buf, &op2Buf)
		case op2, ok := <-c.Op2:
			if !ok {
				c.Op2 = nil
				break
			}

			addOp(op2, &op2Buf, &op1Buf)
		}
	}
}

// echo passes input to the output.
type echo struct {
	In  <-chan int
	Out chan<- int
}

func (c *echo) Process() {
	for i := range c.In {
		c.Out <- i
	}
}

// repeater repeats an input string a given number of times.
type repeater struct {
	Word  <-chan string
	Times <-chan int

	Words chan<- string
}

func (c *repeater) Process() {
	times := 0
	word := ""

	for c.Times != nil || c.Word != nil {
		select {
		case t, ok := <-c.Times:
			if !ok {
				c.Times = nil
				break
			}

			times = t
			c.repeat(word, times)
		case w, ok := <-c.Word:
			if !ok {
				c.Word = nil
				break
			}

			word = w
			c.repeat(word, times)
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

// router routes input map port to output.
type router struct {
	In  map[string]<-chan int
	Out map[string]chan<- int
}

// Process routes incoming packets to the output by sending them to the same
// outport key as the inport key they arrived at.
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

// irouter routes input array port to output.
type irouter struct {
	In  [](<-chan int)
	Out [](chan<- int)
}

// Process routes incoming packets to the output by sending them to the same
// outport key as the inport key they arrived at.
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
	f.Register("echo", func() (Component, error) { return new(echo), nil })
	f.Annotate("echo", Annotation{
		Description: "Passes an int from in to out without changing it",
		Icon:        "arrow-right",
	})

	f.Register("doubler", func() (Component, error) { return new(doubler), nil })
	f.Annotate("doubler", Annotation{
		Description: "Doubles its input",
		Icon:        "times-circle",
	})

	f.Register("repeater", func() (Component, error) { return new(repeater), nil })
	f.Annotate("repeater", Annotation{
		Description: "Repeats Word given numer of Times",
		Icon:        "times-circle",
	})

	f.Register("adder", func() (Component, error) { return new(adder), nil })
	f.Annotate("adder", Annotation{
		Description: "Sums integers coming to its inports",
		Icon:        "plus-circle",
	})

	return nil
}
