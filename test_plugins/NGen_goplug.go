package main

//randomNumberGenerator something
type generator struct {
	Init chan [10]int
	Out  chan int //output
}

// Process listen
func (c *generator) Process() {
	values := <-c.Init
	for _, v := range values {
		//fmt.Println("Generating", v)
		c.Out <- v
	}
}

//NGen something
func NGen() (interface{}, error) {
	return new(generator), nil
}
