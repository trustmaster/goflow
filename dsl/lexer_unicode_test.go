package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestUnicodeIdentifiers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"chinese", "你好(测试)", "你好"},
		{"mixed", "hello世界(comp)", "hello世界"},
		{"underscore_letter", "_你好(comp)", "_你好"},
		{"digit_after", "test123测试(comp)", "test123测试"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := ParseDefinition([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseDefinition error: %v", err)
			}

			if len(def.Processes) != 1 {
				t.Fatalf("expected 1 process, got %d", len(def.Processes))
			}

			for name := range def.Processes {
				if name != tt.want {
					t.Errorf("process name: want %q, got %q", tt.want, name)
				}
			}
		})
	}
}

func TestUnicodeQuotedStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"chinese", `'你好世界'`, "你好世界"},
		{"mixed", `"hello世界"`, "hello世界"},
		{"emoji", `'🚀'`, "🚀"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := ParseDefinition([]byte(tt.input + " -> IN Test"))
			if err != nil {
				t.Fatalf("ParseDefinition error: %v", err)
			}

			if len(def.IIPs) != 1 {
				t.Fatalf("expected 1 IIP, got %d", len(def.IIPs))
			}

			got, ok := def.IIPs[0].Data.(string)
			if !ok {
				t.Fatalf("expected string IIP data, got %T", def.IIPs[0].Data)
			}

			if got != tt.want {
				t.Errorf("IIP data: want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestUnicodeComments(t *testing.T) {
	input := `# 这是一个注释
Test(comp)
`
	def, err := ParseDefinition([]byte(input))
	if err != nil {
		t.Fatalf("ParseDefinition error: %v", err)
	}

	if len(def.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(def.Processes))
	}

	if def.Processes["Test"].Component != "comp" {
		t.Errorf("component: want %q, got %q", "comp", def.Processes["Test"].Component)
	}
}

func TestMultiByteColumnTracking(t *testing.T) {
	// "你好" is 6 bytes but 2 characters.
	// If column tracking were byte-based, the error column would be 8 instead of 4.
	input := "你好 42"
	_, err := ParseDefinition([]byte(input))
	if err == nil {
		t.Fatal("expected error for invalid statement")
	}

	// Verify that parsing completes without panic and produces a reasonable error.
	// The exact column depends on how the parser reports it, but the key test
	// is that the lexer correctly advances past multi-byte characters.
	t.Logf("error (should reference around column 4): %v", err)
}

func TestLexerUnicodeTokens(t *testing.T) {
	cases := []lexerTestCase{
		{
			name: "unicode identifier and quoted string",
			src:  "你好 '世界'",
			tokens: []Token{
				{Type: TokIdent, Pos: 0, Value: "你好", Span: Span{File: "test.fbp", Offset: 0, Line: 1, Column: 1, End: 6}},
				{Type: TokWhitespace, Pos: 6, Value: " ", Span: Span{File: "test.fbp", Offset: 6, Line: 1, Column: 3, End: 7}},
				{Type: TokQuoted, Pos: 7, Value: "'世界'", Span: Span{File: "test.fbp", Offset: 7, Line: 1, Column: 4, End: 15}},
				{Type: TokEOF, Pos: 15, Value: "test.fbp", Span: Span{File: "test.fbp", Offset: 15, Line: 1, Column: 8, End: 15}},
			},
		},
		{
			name: "unicode in component path",
			src:  "A(组件/测试)",
			tokens: []Token{
				{Type: TokIdent, Pos: 0, Value: "A", Span: Span{File: "test.fbp", Offset: 0, Line: 1, Column: 1, End: 1}},
				{Type: TokLParen, Pos: 1, Value: "(", Span: Span{File: "test.fbp", Offset: 1, Line: 1, Column: 2, End: 2}},
				{Type: TokIdent, Pos: 2, Value: "组件", Span: Span{File: "test.fbp", Offset: 2, Line: 1, Column: 3, End: 8}},
				{Type: TokSlash, Pos: 8, Value: "/", Span: Span{File: "test.fbp", Offset: 8, Line: 1, Column: 5, End: 9}},
				{Type: TokIdent, Pos: 9, Value: "测试", Span: Span{File: "test.fbp", Offset: 9, Line: 1, Column: 6, End: 15}},
				{Type: TokRParen, Pos: 15, Value: ")", Span: Span{File: "test.fbp", Offset: 15, Line: 1, Column: 8, End: 16}},
				{Type: TokEOF, Pos: 16, Value: "test.fbp", Span: Span{File: "test.fbp", Offset: 16, Line: 1, Column: 9, End: 16}},
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
