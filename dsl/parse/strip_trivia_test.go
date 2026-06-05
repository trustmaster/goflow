package parse_test

import (
	"testing"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/parse"
	"github.com/trustmaster/goflow/dsl/types"
)

func TestStripTrivia(t *testing.T) {
	file := &types.File{Name: "test.fbp", Data: []byte("A # note\nB")}
	tokens := []types.Token{
		testToken(file, types.TokIdent, 0, 1, 1, 1, "A"),
		testToken(file, types.TokWhitespace, 1, 1, 2, 2, " "),
		testToken(file, types.TokComment, 2, 1, 3, 8, "# note"),
		testToken(file, types.TokEOL, 8, 1, 9, 9, "\n"),
		testToken(file, types.TokWhitespace, 9, 2, 1, 10, "\t"),
		testToken(file, types.TokIdent, 10, 2, 2, 11, "B"),
		testToken(file, types.TokEOF, 11, 2, 3, 11, file.Name),
	}
	want := []types.Token{tokens[0], tokens[3], tokens[5], tokens[6]}

	f := goflow.NewFactory()
	if err := parse.RegisterParseComponents(f); err != nil {
		t.Fatal(err)
	}

	component, err := f.Create("dsl/StripTrivia")
	if err != nil {
		t.Fatal(err)
	}

	strip := component.(*parse.StripTrivia)
	in := make(chan types.Token)
	out := make(chan types.Token, len(tokens))
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
