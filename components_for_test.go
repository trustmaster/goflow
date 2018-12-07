package flow

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
	defer close(c.Sum)

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
	close(c.Out)
}

// repeater repeats an input string a given number of times
type repeater struct {
	Word  <-chan string
	Times <-chan int

	Words chan<- string
}

func (c *repeater) Process() {
	defer close(c.Words)
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
