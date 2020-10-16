package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

// StartToken(dsl/StartToken) INIT -> Merge
func TestTokenizer(t *testing.T) {
	fbpData := "StartToken(dsl/StartToken) INIT -> Merge"
	expected := []Token{
		{Type: tokNewFile, Pos: 0, Value: "test.fbp"},
		{Type: tokIdent, Pos: 0, Value: "StartToken"},
		{Type: tokLparen, Pos: 10, Value: "("},
		{Type: tokIdent, Pos: 11, Value: "dsl"},
		{Type: tokSlash, Pos: 14, Value: "/"},
		{Type: tokIdent, Pos: 15, Value: "StartToken"},
		{Type: tokRparen, Pos: 25, Value: ")"},
		{Type: tokWhitespace, Pos: 26, Value: " "},
		{Type: tokIdent, Pos: 27, Value: "INIT"},
		{Type: tokWhitespace, Pos: 31, Value: " "},
		{Type: tokArrow, Pos: 32, Value: "->"},
		{Type: tokWhitespace, Pos: 34, Value: " "},
		{Type: tokIdent, Pos: 35, Value: "Merge"},
		{Type: tokEOF, Pos: 40, Value: "test.fbp"},
	}

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

	runTokenizerTestCase(t, n, fbpData, expected)
}

func runTokenizerTestCase(t *testing.T, n *goflow.Graph, fbpData string, expected []Token) {
	in := make(chan *File)
	out := make(chan Token)

	if err := n.SetInPort("In", in); err != nil {
		t.Error(err)
		return
	}

	if err := n.SetOutPort("Out", out); err != nil {
		t.Error(err)
		return
	}

	file := &File{
		Name: "test.fbp",
		Data: []byte(fbpData),
	}

	wait := goflow.Run(n)

	go func() {
		in <- file
		close(in)
	}()

	j := 0

	for tok := range out {
		if !tokEql(tok, expected[j]) {
			t.Errorf("Expected '%s': '%s' at %d, got '%s': '%s' at %d", expected[j].Type, expected[j].Value, expected[j].Pos, tok.Type, tok.Value, tok.Pos)
		}
		j++
	}

	if j != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), j)
	}

	<-wait
}

func tokEql(t1, t2 Token) bool {
	return t1.Type == t2.Type &&
		t1.Pos == t2.Pos &&
		t1.Value == t2.Value
}
