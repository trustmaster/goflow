package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestSplit(t *testing.T) {
	data := "ScanFile(dsl/ScanFile) OUT -> IN Split(dsl/Split)"
	fname := "split.fbp"
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

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	i, err := f.Create("dsl/Split")
	if err != nil {
		t.Error(err)
		return
	}

	split := i.(*Split)

	runSplitTestCase(t, split, numOuts, tokens)
}

func runSplitTestCase(t *testing.T, split *Split, numOuts int, tokens []Token) {
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
		for i := range tokens {
			in <- tokens[i]
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
