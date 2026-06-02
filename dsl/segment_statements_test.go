package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestSegmentStatements(t *testing.T) {
	file := &File{Name: "test.fbp", Data: []byte("A -> B\n\n'x' -> IN Sink")}
	cases := []struct {
		name  string
		input []Token
		want  []Statement
	}{
		{
			name: "splits on eol and ignores empty statements",
			input: []Token{
				testToken(file, TokIdent, 0, 1, 1, 1, "A"),
				testToken(file, TokArrow, 2, 1, 3, 4, "->"),
				testToken(file, TokIdent, 5, 1, 6, 6, "B"),
				testToken(file, TokEOL, 6, 1, 7, 7, "\n"),
				testToken(file, TokEOL, 7, 2, 1, 8, "\n"),
				testToken(file, TokQuoted, 8, 3, 1, 11, "'x'"),
				testToken(file, TokArrow, 12, 3, 5, 14, "->"),
				testToken(file, TokIdent, 15, 3, 8, 17, "IN"),
				testToken(file, TokIdent, 18, 3, 11, 22, "Sink"),
				testToken(file, TokEOF, 22, 3, 15, 22, file.Name),
			},
			want: []Statement{
				{
					Tokens: []Token{
						testToken(file, TokIdent, 0, 1, 1, 1, "A"),
						testToken(file, TokArrow, 2, 1, 3, 4, "->"),
						testToken(file, TokIdent, 5, 1, 6, 6, "B"),
					},
					Span: Span{File: file.Name, Offset: 0, Line: 1, Column: 1, End: 6},
				},
				{
					Tokens: []Token{
						testToken(file, TokQuoted, 8, 3, 1, 11, "'x'"),
						testToken(file, TokArrow, 12, 3, 5, 14, "->"),
						testToken(file, TokIdent, 15, 3, 8, 17, "IN"),
						testToken(file, TokIdent, 18, 3, 11, 22, "Sink"),
					},
					Span: Span{File: file.Name, Offset: 8, Line: 3, Column: 1, End: 22},
				},
			},
		},
		{
			name: "flushes trailing tokens when input closes without eof",
			input: []Token{
				testToken(file, TokIdent, 0, 1, 1, 1, "A"),
				testToken(file, TokArrow, 2, 1, 3, 4, "->"),
				testToken(file, TokIdent, 5, 1, 6, 6, "B"),
			},
			want: []Statement{
				{
					Tokens: []Token{
						testToken(file, TokIdent, 0, 1, 1, 1, "A"),
						testToken(file, TokArrow, 2, 1, 3, 4, "->"),
						testToken(file, TokIdent, 5, 1, 6, 6, "B"),
					},
					Span: Span{File: file.Name, Offset: 0, Line: 1, Column: 1, End: 6},
				},
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
			component, err := f.Create("dsl/SegmentStatements")
			if err != nil {
				t.Fatal(err)
			}

			segment := component.(*SegmentStatements)
			in := make(chan Token)
			out := make(chan Statement, len(tc.want)+1)
			segment.In = in
			segment.Out = out

			wait := goflow.Run(segment)
			go func() {
				for j := range tc.input {
					in <- tc.input[j]
				}
				close(in)
			}()

			<-wait

			got := readStatements(t, out, len(tc.want))
			assertStatements(t, got, tc.want)
		})
	}
}

func TestLexerToStatementsPipeline(t *testing.T) {
	src := "# comment-only line\nRead(dsl/Reader) OUT -> IN Parse\n\n'hello' -> IN Print # trailing comment\n"
	file := &File{Name: "test.fbp", Data: []byte(src)}

	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		t.Fatal(err)
	}

	component, err := f.Create("dsl/Lexer")
	if err != nil {
		t.Fatal(err)
	}

	lexer := component.(*goflow.Graph)
	lexerIn := make(chan *File)
	lexerOut := make(chan Token)
	if err := lexer.SetInPort("In", lexerIn); err != nil {
		t.Fatal(err)
	}
	if err := lexer.SetOutPort("Out", lexerOut); err != nil {
		t.Fatal(err)
	}

	stripped := make(chan Token, 32)
	statementsOut := make(chan Statement, 4)
	strip := &StripTrivia{In: lexerOut, Out: stripped}
	segment := &SegmentStatements{In: stripped, Out: statementsOut}

	lexerWait := goflow.Run(lexer)
	stripWait := goflow.Run(strip)
	segmentWait := goflow.Run(segment)

	go func() {
		lexerIn <- file
		close(lexerIn)
	}()

	<-lexerWait
	<-stripWait
	<-segmentWait

	want := []Statement{
		{
			Tokens: []Token{
				testToken(file, TokIdent, 20, 2, 1, 24, "Read"),
				testToken(file, TokLParen, 24, 2, 5, 25, "("),
				testToken(file, TokIdent, 25, 2, 6, 28, "dsl"),
				testToken(file, TokSlash, 28, 2, 9, 29, "/"),
				testToken(file, TokIdent, 29, 2, 10, 35, "Reader"),
				testToken(file, TokRParen, 35, 2, 16, 36, ")"),
				testToken(file, TokIdent, 37, 2, 18, 40, "OUT"),
				testToken(file, TokArrow, 41, 2, 22, 43, "->"),
				testToken(file, TokIdent, 44, 2, 25, 46, "IN"),
				testToken(file, TokIdent, 47, 2, 28, 52, "Parse"),
			},
			Span: Span{File: file.Name, Offset: 20, Line: 2, Column: 1, End: 52},
		},
		{
			Tokens: []Token{
				testToken(file, TokQuoted, 54, 4, 1, 61, "'hello'"),
				testToken(file, TokArrow, 62, 4, 9, 64, "->"),
				testToken(file, TokIdent, 65, 4, 12, 67, "IN"),
				testToken(file, TokIdent, 68, 4, 15, 73, "Print"),
			},
			Span: Span{File: file.Name, Offset: 54, Line: 4, Column: 1, End: 73},
		},
	}

	got := readStatements(t, statementsOut, len(want))
	assertStatements(t, got, want)
}

func testToken(file *File, tokenType TokenType, pos, line, column, end int, value string) Token {
	return Token{
		Type:  tokenType,
		File:  file,
		Pos:   pos,
		Span:  Span{File: file.Name, Offset: pos, Line: line, Column: column, End: end},
		Value: value,
	}
}

func readTokens(t *testing.T, ch <-chan Token, want int) []Token {
	t.Helper()

	tokens := make([]Token, 0, want)
	for i := 0; i < want; i++ {
		select {
		case tok, ok := <-ch:
			if !ok {
				t.Fatalf("expected %d tokens, got %d", want, len(tokens))
			}
			tokens = append(tokens, tok)
		default:
			t.Fatalf("expected %d tokens, got %d", want, len(tokens))
		}
	}

	select {
	case tok, ok := <-ch:
		if ok {
			t.Fatalf("got unexpected extra token %#v", tok)
		}
	default:
	}

	return tokens
}

func readStatements(t *testing.T, ch <-chan Statement, want int) []Statement {
	t.Helper()

	statements := make([]Statement, 0, want)
	for i := 0; i < want; i++ {
		select {
		case statement, ok := <-ch:
			if !ok {
				t.Fatalf("expected %d statements, got %d", want, len(statements))
			}
			statements = append(statements, statement)
		default:
			t.Fatalf("expected %d statements, got %d", want, len(statements))
		}
	}

	select {
	case statement, ok := <-ch:
		if ok {
			t.Fatalf("got unexpected extra statement %#v", statement)
		}
	default:
	}

	return statements
}

func assertTokenSlice(t *testing.T, got, want []Token) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d", len(want), len(got))
	}

	for i := range want {
		if !lexerTokEql(got[i], want[i]) {
			t.Fatalf("token %d mismatch: expected %#v, got %#v", i, want[i], got[i])
		}
	}
}

func assertStatements(t *testing.T, got, want []Statement) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d statements, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i].Span != want[i].Span {
			t.Fatalf("statement %d span mismatch: expected %#v, got %#v", i, want[i].Span, got[i].Span)
		}

		assertTokenSlice(t, got[i].Tokens, want[i].Tokens)
	}
}
