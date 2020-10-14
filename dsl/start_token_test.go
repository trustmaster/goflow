package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestStartToken(t *testing.T) {
	data := "ScanFile(dsl/ScanFile) OUT -> IN Split(dsl/Split)"
	fname := "tokenizer.fbp"
	file := &File{Name: fname, Data: []byte(data)}

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Error(err)
		return
	}

	i, err := f.Create("dsl/StartToken")
	if err != nil {
		t.Error(err)
		return
	}

	start := i.(*StartToken)

	in := make(chan *File)
	init := make(chan Token)
	next := make(chan Token)
	start.File = in
	start.Init = init
	start.Next = next

	wait := goflow.Run(start)

	go func() {
		in <- file
		close(in)
	}()

	tok := <-init
	if tok.Type != tokNewFile {
		t.Errorf("unexpected token type %s", tok.Type)
	}

	if tok.File != file {
		t.Errorf("unexpected file")
	}

	t2 := <-next
	if t2 != tok {
		t.Errorf("Init and Next tokens are not equal")
	}

	<-wait
}
