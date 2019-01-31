package dsl

import "io/ioutil"

// Reader reads a file to string
type Reader struct {
	File <-chan string
	Data chan<- string
	Err  chan<- error
}

// Process handles the input and transforms it to output
func (c *Reader) Process() {
	for name := range c.File {
		data, err := ioutil.ReadFile(name)
		if err == nil {
			c.Data <- string(data)
		} else {
			c.Err <- err
		}
	}
}
