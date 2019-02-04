package dsl

import (
	"io"
	"os"
)

// ReadFile opens a file for reading
type ReadFile struct {
	File   <-chan string
	Reader chan<- io.Reader
	Err    chan<- error
}

// Process handles the input and transforms it to output
func (c *ReadFile) Process() {
	for name := range c.File {
		file, err := os.Open(name)
		if err == nil {
			c.Reader <- file
		} else {
			c.Err <- err
		}
	}
}
