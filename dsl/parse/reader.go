package parse

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/trustmaster/goflow/dsl/types"
)

// FileError is an error while reading from a file.
type FileError struct {
	Name string
	Err  error
}

// Error returns an error message.
func (e FileError) Error() string {
	return fmt.Sprintf("Error while opening the file '%s': %s", e.Name, e.Err.Error())
}

// Reader opens a file for reading.
type Reader struct {
	Name <-chan string
	File chan<- *types.File
	Err  chan<- FileError
}

// Process handles the input and transforms it to output.
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
		data, err := os.ReadFile(filepath.Clean(name))
		if !check(err, name) {
			continue
		}

		c.File <- &types.File{
			Name: name,
			Data: data,
		}
	}
}
