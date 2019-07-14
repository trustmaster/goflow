package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestTokenizer(t *testing.T) {
	in := make(chan *File)
	out := make(chan Token)
	e := make(chan LexError)

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	i, err := f.Create("dsl/Tokenizer")
	if err != nil {
		t.Error(err)
		return
	}
	c := i.(*Tokenizer)
	c.File = in
	c.Token = out
	c.Err = e
}
