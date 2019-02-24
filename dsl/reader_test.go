package dsl

import (
	"io/ioutil"
	"testing"

	"github.com/trustmaster/goflow"
)

func TestReader(t *testing.T) {
	in := make(chan string, 2)
	out := make(chan File)
	e := make(chan FileError)

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	i, err := f.Create("dsl/Reader")
	if err != nil {
		t.Error(err)
		return
	}
	c := i.(*Reader)
	c.Filename = in
	c.File = out
	c.Err = e

	wait := goflow.Run(c)

	filenames := []string{"dsl.fbp", "404notfound.fbp"}
	expectations := []string{"data", "error"}

	for _, name := range filenames {
		in <- name
	}

	go func() {
		for len(expectations) > 0 {
			expected := expectations[0]
			expectations = expectations[1:]
			name := filenames[0]
			filenames = filenames[1:]
			select {
			case f := <-out:
				if f.Name != name {
					t.Errorf("Expected file '%s', got '%s'", name, f.Name)
					break
				}
				if expected == "error" {
					t.Errorf("Unexpected Reader")
					break
				}
				_, err := ioutil.ReadAll(f.Reader)
				if err != nil {
					t.Error(err)
				}
			case fe := <-e:
				if fe.Name != name {
					t.Errorf("Expected file '%s', got '%s'", name, fe.Name)
					break
				}
				if expected == "data" {
					t.Errorf("Unexpected error: %s", fe.Err.Error())
					break
				}
			}
		}
		close(in)
	}()

	<-wait
}
