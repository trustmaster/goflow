package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestStripTrivia(t *testing.T) {
	file := &File{Name: "test.fbp", Data: []byte("A # note\nB")}
	tokens := []Token{
		testToken(file, TokIdent, 0, 1, 1, 1, "A"),
		testToken(file, TokWhitespace, 1, 1, 2, 2, " "),
		testToken(file, TokComment, 2, 1, 3, 8, "# note"),
		testToken(file, TokEOL, 8, 1, 9, 9, "\n"),
		testToken(file, TokWhitespace, 9, 2, 1, 10, "\t"),
		testToken(file, TokIdent, 10, 2, 2, 11, "B"),
		testToken(file, TokEOF, 11, 2, 3, 11, file.Name),
	}
	want := []Token{tokens[0], tokens[3], tokens[5], tokens[6]}

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Fatal(err)
	}

	component, err := f.Create("dsl/StripTrivia")
	if err != nil {
		t.Fatal(err)
	}

	strip := component.(*StripTrivia)
	in := make(chan Token)
	out := make(chan Token, len(tokens))
	strip.In = in
	strip.Out = out

	wait := goflow.Run(strip)
	go func() {
		for i := range tokens {
			in <- tokens[i]
		}
		close(in)
	}()

	<-wait

	got := readTokens(t, out, len(want))
	assertTokenSlice(t, got, want)
}
