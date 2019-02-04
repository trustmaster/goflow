package dsl

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/trustmaster/goflow"
)

func TestReadFile(t *testing.T) {
	in := make(chan string, 2)
	out := make(chan io.Reader)
	e := make(chan error)

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	i, err := f.Create("dsl/ReadFile")
	if err != nil {
		t.Error(err)
		return
	}
	c := i.(*ReadFile)
	c.File = in
	c.Reader = out
	c.Err = e

	wait := goflow.Run(c)

	in <- "dsl.fbp"
	in <- "404notfound.fbp"

	go func() {
		expectations := []string{"data", "error"}
		for len(expectations) > 0 {
			expected := expectations[0]
			expectations = expectations[1:]
			select {
			case r := <-out:
				if expected == "error" {
					t.Errorf("Unexpected Reader")
					break
				}
				_, err := ioutil.ReadAll(r)
				if err != nil {
					t.Error(err)
				}
			case err := <-e:
				if expected == "data" {
					t.Errorf("Unexpected error: %s", err.Error())
					break
				}
			}
		}
		close(in)
	}()

	<-wait
}
