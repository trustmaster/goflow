package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestMerge(t *testing.T) {
	data := "ScanFile(dsl/ScanFile) OUT -> IN Merge(dsl/Merge)"
	fname := "merge.fbp"
	file := &File{Name: fname, Data: []byte(data)}
	tokens := []Token{
		{
			File:  file,
			Pos:   0,
			Type:  tokIdent,
			Value: "ScanFile",
		},
		{
			File:  file,
			Pos:   8,
			Type:  tokLparen,
			Value: "(",
		},
		{
			File:  file,
			Pos:   9,
			Type:  tokIdent,
			Value: "dsl",
		},
	}

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	i, err := f.Create("dsl/Merge")
	if err != nil {
		t.Error(err)
		return
	}

	merge := i.(*Merge)

	in := make(chan Token, 3)
	out := make(chan Token)
	merge.In = in
	merge.Out = out

	wait := goflow.Run(merge)

	go func() {
		for _, t := range tokens {
			in <- t
		}

		close(in)
	}()

	for i := 0; i < len(tokens); i++ {
		tok := <-out
		if tok != tokens[i] {
			t.Errorf("Expected %s '%s', got %s '%s'", tokens[i].Type, tokens[i].Value, tok.Type, tok.Value)
		}
	}

	<-wait
}
