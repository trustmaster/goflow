package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

type lexerTestCase struct {
	name   string
	src    string
	tokens []Token
}

func TestLexer(t *testing.T) {
	cases := []lexerTestCase{
		{
			name: "connections, array ports, and eol",
			src:  "Read(dsl/Reader) OUT[0] -> IN Parse\n",
			tokens: []Token{
				{Type: TokIdent, Pos: 0, Value: "Read", Span: Span{File: "test.fbp", Offset: 0, Line: 1, Column: 1, End: 4}},
				{Type: TokLParen, Pos: 4, Value: "(", Span: Span{File: "test.fbp", Offset: 4, Line: 1, Column: 5, End: 5}},
				{Type: TokIdent, Pos: 5, Value: "dsl", Span: Span{File: "test.fbp", Offset: 5, Line: 1, Column: 6, End: 8}},
				{Type: TokSlash, Pos: 8, Value: "/", Span: Span{File: "test.fbp", Offset: 8, Line: 1, Column: 9, End: 9}},
				{Type: TokIdent, Pos: 9, Value: "Reader", Span: Span{File: "test.fbp", Offset: 9, Line: 1, Column: 10, End: 15}},
				{Type: TokRParen, Pos: 15, Value: ")", Span: Span{File: "test.fbp", Offset: 15, Line: 1, Column: 16, End: 16}},
				{Type: TokWhitespace, Pos: 16, Value: " ", Span: Span{File: "test.fbp", Offset: 16, Line: 1, Column: 17, End: 17}},
				{Type: TokIdent, Pos: 17, Value: "OUT", Span: Span{File: "test.fbp", Offset: 17, Line: 1, Column: 18, End: 20}},
				{Type: TokLBracket, Pos: 20, Value: "[", Span: Span{File: "test.fbp", Offset: 20, Line: 1, Column: 21, End: 21}},
				{Type: TokInt, Pos: 21, Value: "0", Span: Span{File: "test.fbp", Offset: 21, Line: 1, Column: 22, End: 22}},
				{Type: TokRBracket, Pos: 22, Value: "]", Span: Span{File: "test.fbp", Offset: 22, Line: 1, Column: 23, End: 23}},
				{Type: TokWhitespace, Pos: 23, Value: " ", Span: Span{File: "test.fbp", Offset: 23, Line: 1, Column: 24, End: 24}},
				{Type: TokArrow, Pos: 24, Value: "->", Span: Span{File: "test.fbp", Offset: 24, Line: 1, Column: 25, End: 26}},
				{Type: TokWhitespace, Pos: 26, Value: " ", Span: Span{File: "test.fbp", Offset: 26, Line: 1, Column: 27, End: 27}},
				{Type: TokIdent, Pos: 27, Value: "IN", Span: Span{File: "test.fbp", Offset: 27, Line: 1, Column: 28, End: 29}},
				{Type: TokWhitespace, Pos: 29, Value: " ", Span: Span{File: "test.fbp", Offset: 29, Line: 1, Column: 30, End: 30}},
				{Type: TokIdent, Pos: 30, Value: "Parse", Span: Span{File: "test.fbp", Offset: 30, Line: 1, Column: 31, End: 35}},
				{Type: TokEOL, Pos: 35, Value: "\n", Span: Span{File: "test.fbp", Offset: 35, Line: 1, Column: 36, End: 36}},
				{Type: TokEOF, Pos: 36, Value: "test.fbp", Span: Span{File: "test.fbp", Offset: 36, Line: 2, Column: 1, End: 36}},
			},
		},
		{
			name: "quoted strings, comment, and crlf",
			src:  "'a\\'b' # note\r\nOUTPORT=Parser.OUT:TREE",
			tokens: []Token{
				{Type: TokQuoted, Pos: 0, Value: "'a'b'", Span: Span{File: "test.fbp", Offset: 0, Line: 1, Column: 1, End: 6}},
				{Type: TokWhitespace, Pos: 6, Value: " ", Span: Span{File: "test.fbp", Offset: 6, Line: 1, Column: 7, End: 7}},
				{Type: TokComment, Pos: 7, Value: "# note", Span: Span{File: "test.fbp", Offset: 7, Line: 1, Column: 8, End: 13}},
				{Type: TokEOL, Pos: 13, Value: "\r\n", Span: Span{File: "test.fbp", Offset: 13, Line: 1, Column: 14, End: 15}},
				{Type: TokIdent, Pos: 15, Value: "OUTPORT", Span: Span{File: "test.fbp", Offset: 15, Line: 2, Column: 1, End: 22}},
				{Type: TokEqual, Pos: 22, Value: "=", Span: Span{File: "test.fbp", Offset: 22, Line: 2, Column: 8, End: 23}},
				{Type: TokIdent, Pos: 23, Value: "Parser", Span: Span{File: "test.fbp", Offset: 23, Line: 2, Column: 9, End: 29}},
				{Type: TokDot, Pos: 29, Value: ".", Span: Span{File: "test.fbp", Offset: 29, Line: 2, Column: 15, End: 30}},
				{Type: TokIdent, Pos: 30, Value: "OUT", Span: Span{File: "test.fbp", Offset: 30, Line: 2, Column: 16, End: 33}},
				{Type: TokColon, Pos: 33, Value: ":", Span: Span{File: "test.fbp", Offset: 33, Line: 2, Column: 19, End: 34}},
				{Type: TokIdent, Pos: 34, Value: "TREE", Span: Span{File: "test.fbp", Offset: 34, Line: 2, Column: 20, End: 38}},
				{Type: TokEOF, Pos: 38, Value: "test.fbp", Span: Span{File: "test.fbp", Offset: 38, Line: 2, Column: 24, End: 38}},
			},
		},
	}

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Fatal(err)
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			component, err := f.Create("dsl/Lexer")
			if err != nil {
				t.Fatal(err)
			}

			lexer := component.(*goflow.Graph)
			in := make(chan *File)
			out := make(chan Token)
			if err := lexer.SetInPort("In", in); err != nil {
				t.Fatal(err)
			}
			if err := lexer.SetOutPort("Out", out); err != nil {
				t.Fatal(err)
			}

			wait := goflow.Run(lexer)
			go func() {
				in <- &File{Name: "test.fbp", Data: []byte(tc.src)}
				close(in)
			}()

			var tokens []Token
			for tok := range out {
				tokens = append(tokens, tok)
			}

			<-wait

			if len(tokens) != len(tc.tokens) {
				t.Fatalf("expected %d tokens, got %d", len(tc.tokens), len(tokens))
			}

			for idx := range tc.tokens {
				if !lexerTokEql(tokens[idx], tc.tokens[idx]) {
					t.Fatalf("token %d mismatch: expected %#v, got %#v", idx, tc.tokens[idx], tokens[idx])
				}
			}
		})
	}
}

func lexerTokEql(actual, expected Token) bool {
	return actual.Type == expected.Type &&
		actual.Pos == expected.Pos &&
		actual.Value == expected.Value &&
		actual.Span == expected.Span
}
