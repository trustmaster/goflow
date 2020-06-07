package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestTokenizer(t *testing.T) {
	in := make(chan *File)
	out := make(chan Token)

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
	n := i.(*goflow.Graph)
	if err := n.SetInPort("In", in); err != nil {
		t.Error(err)
		return
	}
	if err := n.SetOutPort("Out", out); err != nil {
		t.Error(err)
		return
	}
}
