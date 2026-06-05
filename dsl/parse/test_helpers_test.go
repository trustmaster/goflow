package parse_test

import (
	"testing"

	"github.com/trustmaster/goflow/dsl/types"
)

func testToken(file *types.File, tokenType types.TokenType, pos, line, column, end int, value string) types.Token {
	return types.Token{
		Type:  tokenType,
		File:  file,
		Pos:   pos,
		Span:  types.Span{File: file.Name, Offset: pos, Line: line, Column: column, End: end},
		Value: value,
	}
}

func readTokens(t *testing.T, ch <-chan types.Token, want int) []types.Token {
	t.Helper()

	tokens := make([]types.Token, 0, want)
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

func readStatements(t *testing.T, ch <-chan types.Statement, want int) []types.Statement {
	t.Helper()

	statements := make([]types.Statement, 0, want)
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

func assertTokenSlice(t *testing.T, got, want []types.Token) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d", len(want), len(got))
	}

	for i := range want {
		if !tokenEqual(got[i], want[i]) {
			t.Fatalf("token %d mismatch: expected %#v, got %#v", i, want[i], got[i])
		}
	}
}

func assertStatements(t *testing.T, got, want []types.Statement) {
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

// tokenEqual compares two tokens for equality (ignoring some internal details).
func tokenEqual(a, b types.Token) bool {
	return a.Type == b.Type &&
		a.Value == b.Value &&
		a.Span.File == b.Span.File &&
		a.Span.Offset == b.Span.Offset &&
		a.Span.Line == b.Span.Line &&
		a.Span.Column == b.Span.Column &&
		a.Span.End == b.Span.End
}
