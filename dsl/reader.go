package dsl

import (
	"fmt"
	"io"
	"os"
)

// File represents a source file
type File struct {
	Name   string
	Reader io.Reader
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
	Filename <-chan string
	File     chan<- File
	Err      chan<- FileError
}

// Process handles the input and transforms it to output
func (c *Reader) Process() {
	for name := range c.Filename {
		file, err := os.Open(name)
		if err == nil {
			c.File <- File{
				Name:   name,
				Reader: file,
			}
		} else {
			c.Err <- FileError{
				Name: name,
				Err:  err,
			}
		}
	}
}
