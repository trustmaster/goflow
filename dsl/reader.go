package dsl

import (
	"fmt"
	"io/ioutil"
	"os"
)

// File represents a source file
type File struct {
	Name string
	Data []byte
}

// FileError is an error while reading from a file
type FileError struct {
	Name string
	Err  error
}

// Error returns an error message
func (e FileError) Error() string {
	return fmt.Sprintf("Error while opening the file '%s': %s", e.Name, e.Err.Error())
}

// Reader opens a file for reading
type Reader struct {
	Name <-chan string
	File chan<- *File
	Err  chan<- FileError
}

// Process handles the input and transforms it to output
func (c *Reader) Process() {
	check := func(err error, name string) bool {
		if err != nil {
			c.Err <- FileError{
				Name: name,
				Err:  err,
			}
			return false
		}
		return true
	}
	for name := range c.Name {
		r, err := os.Open(name)
		if !check(err, name) {
			continue
		}
		data, err := ioutil.ReadAll(r)
		if !check(err, name) {
			continue
		}
		c.File <- &File{
			Name: name,
			Data: data,
		}
	}
}
