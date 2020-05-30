package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestSplit(t *testing.T) {
	data := "ScanFile(dsl/ScanFile) OUT -> IN Split(dsl/Split)"
	fname := "tokenizer.fbp"
	file := &File{Name: fname, Data: []byte(data)}
	numOuts := 3
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

	split := new(Split)

	in := make(chan Token)
	outs := make([](chan Token), numOuts)
	split.In = in
	split.Out = make([](chan<- Token), numOuts)

	for i := 0; i < numOuts; i++ {
		outs[i] = make(chan Token)
		split.Out[i] = outs[i]
	}

	wait := goflow.Run(split)

	go func() {
		for _, t := range tokens {
			in <- t
		}
		close(in)
	}()

	for i := 0; i < len(tokens); i++ {
		for j := 0; j < numOuts; j++ {
			tok := <-(outs[j])
			if tok != tokens[i] {
				t.Errorf("Expected %s '%s', got %s '%s'", tokens[i].Type, tokens[i].Value, tok.Type, tok.Value)
			}
		}
	}

	<-wait
}
