package dsl

import (
	"fmt"
	"testing"

	"github.com/trustmaster/goflow"
)

// intPtr returns a pointer to n, used to express array port index expectations.
func intPtr(n int) *int { return &n }

// assertEndpoint checks that two Endpoint values match.
func assertEndpoint(t *testing.T, label string, got, want Endpoint) {
	t.Helper()

	if got.Process != want.Process {
		t.Errorf("%s Process: want %q, got %q", label, want.Process, got.Process)
	}

	if got.Port != want.Port {
		t.Errorf("%s Port: want %q, got %q", label, want.Port, got.Port)
	}

	switch {
	case want.Index == nil && got.Index != nil:
		t.Errorf("%s Index: want nil, got %d", label, *got.Index)
	case want.Index != nil && got.Index == nil:
		t.Errorf("%s Index: want %d, got nil", label, *want.Index)
	case want.Index != nil && got.Index != nil && *got.Index != *want.Index:
		t.Errorf("%s Index: want %d, got %d", label, *want.Index, *got.Index)
	}
}

// assertFragments checks that a Fragment slice matches the expected slice.
func assertFragments(t *testing.T, got, want []Fragment) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("fragment count: want %d, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i].Kind != want[i].Kind {
			t.Errorf("fragment[%d] kind: want %q, got %q", i, want[i].Kind, got[i].Kind)
			continue
		}

		//nolint:exhaustive // only Process and Connection are relevant for connection tests
		switch want[i].Kind {
		case FragmentProcess:
			if got[i].Process == nil {
				t.Errorf("fragment[%d]: Process is nil", i)
				continue
			}

			if *got[i].Process != *want[i].Process {
				t.Errorf("fragment[%d] Process: want %+v, got %+v", i, *want[i].Process, *got[i].Process)
			}

		case FragmentConnection:
			if got[i].Connection == nil {
				t.Errorf("fragment[%d]: Connection is nil", i)
				continue
			}

			assertEndpoint(t, fmt.Sprintf("fragment[%d].Src", i), got[i].Connection.Src, want[i].Connection.Src)
			assertEndpoint(t, fmt.Sprintf("fragment[%d].Tgt", i), got[i].Connection.Tgt, want[i].Connection.Tgt)

		default:
			t.Errorf("fragment[%d] unexpected kind %q", i, want[i].Kind)
		}
	}
}

func TestParseConnectionStatement(t *testing.T) {
	file := &File{Name: "conn_test.fbp", Data: []byte("parse connection statement test data for spans")}

	// tok is a shorthand: each token is assigned position i with span [i, i+1).
	tok := func(tt TokenType, i int, v string) Token {
		return testToken(file, tt, i, 1, i+1, i+1, v)
	}

	cases := []struct {
		name    string
		tokens  []Token
		want    []Fragment
		wantErr bool
	}{
		{
			// Read(dsl/Reader) OUT -> IN Parse(dsl/Parser)
			name: "full with components",
			tokens: []Token{
				tok(TokIdent, 0, "Read"),
				tok(TokLParen, 1, "("),
				tok(TokIdent, 2, "dsl"),
				tok(TokSlash, 3, "/"),
				tok(TokIdent, 4, "Reader"),
				tok(TokRParen, 5, ")"),
				tok(TokIdent, 6, "OUT"),
				tok(TokArrow, 7, "->"),
				tok(TokIdent, 8, "IN"),
				tok(TokIdent, 9, "Parse"),
				tok(TokLParen, 10, "("),
				tok(TokIdent, 11, "dsl"),
				tok(TokSlash, 12, "/"),
				tok(TokIdent, 13, "Parser"),
				tok(TokRParen, 14, ")"),
			},
			want: []Fragment{
				{Kind: FragmentProcess, Process: &ProcessDef{Name: "Read", Component: "dsl/Reader"}},
				{Kind: FragmentProcess, Process: &ProcessDef{Name: "Parse", Component: "dsl/Parser"}},
				{Kind: FragmentConnection, Connection: &ConnectionDef{
					Src: Endpoint{Process: "Read", Port: "OUT"},
					Tgt: Endpoint{Process: "Parse", Port: "IN"},
				}},
			},
		},
		{
			// Reader OUT -> IN Parser  (no inline components — no process fragments)
			name: "no components",
			tokens: []Token{
				tok(TokIdent, 0, "Reader"),
				tok(TokIdent, 1, "OUT"),
				tok(TokArrow, 2, "->"),
				tok(TokIdent, 3, "IN"),
				tok(TokIdent, 4, "Parser"),
			},
			want: []Fragment{
				{Kind: FragmentConnection, Connection: &ConnectionDef{
					Src: Endpoint{Process: "Reader", Port: "OUT"},
					Tgt: Endpoint{Process: "Parser", Port: "IN"},
				}},
			},
		},
		{
			// Split OUT[0] -> IN Foo
			name: "array port source",
			tokens: []Token{
				tok(TokIdent, 0, "Split"),
				tok(TokIdent, 1, "OUT"),
				tok(TokLBracket, 2, "["),
				tok(TokInt, 3, "0"),
				tok(TokRBracket, 4, "]"),
				tok(TokArrow, 5, "->"),
				tok(TokIdent, 6, "IN"),
				tok(TokIdent, 7, "Foo"),
			},
			want: []Fragment{
				{Kind: FragmentConnection, Connection: &ConnectionDef{
					Src: Endpoint{Process: "Split", Port: "OUT", Index: intPtr(0)},
					Tgt: Endpoint{Process: "Foo", Port: "IN"},
				}},
			},
		},
		{
			// Source OUT -> IN[2] Dest
			name: "array port target",
			tokens: []Token{
				tok(TokIdent, 0, "Source"),
				tok(TokIdent, 1, "OUT"),
				tok(TokArrow, 2, "->"),
				tok(TokIdent, 3, "IN"),
				tok(TokLBracket, 4, "["),
				tok(TokInt, 5, "2"),
				tok(TokRBracket, 6, "]"),
				tok(TokIdent, 7, "Dest"),
			},
			want: []Fragment{
				{Kind: FragmentConnection, Connection: &ConnectionDef{
					Src: Endpoint{Process: "Source", Port: "OUT"},
					Tgt: Endpoint{Process: "Dest", Port: "IN", Index: intPtr(2)},
				}},
			},
		},
		{
			// Src(a/b/c) PORT -> IN Tgt  — three-segment component path
			name: "multi-part component",
			tokens: []Token{
				tok(TokIdent, 0, "Src"),
				tok(TokLParen, 1, "("),
				tok(TokIdent, 2, "a"),
				tok(TokSlash, 3, "/"),
				tok(TokIdent, 4, "b"),
				tok(TokSlash, 5, "/"),
				tok(TokIdent, 6, "c"),
				tok(TokRParen, 7, ")"),
				tok(TokIdent, 8, "PORT"),
				tok(TokArrow, 9, "->"),
				tok(TokIdent, 10, "IN"),
				tok(TokIdent, 11, "Tgt"),
			},
			want: []Fragment{
				{Kind: FragmentProcess, Process: &ProcessDef{Name: "Src", Component: "a/b/c"}},
				{Kind: FragmentConnection, Connection: &ConnectionDef{
					Src: Endpoint{Process: "Src", Port: "PORT"},
					Tgt: Endpoint{Process: "Tgt", Port: "IN"},
				}},
			},
		},
		{
			// Read(dsl/Reader) ->  — arrow where source port expected
			name:    "missing source port",
			wantErr: true,
			tokens: []Token{
				tok(TokIdent, 0, "Read"),
				tok(TokLParen, 1, "("),
				tok(TokIdent, 2, "dsl"),
				tok(TokSlash, 3, "/"),
				tok(TokIdent, 4, "Reader"),
				tok(TokRParen, 5, ")"),
				tok(TokArrow, 6, "->"),
			},
		},
		{
			// Read OUT IN Parse  — IN where arrow expected
			name:    "missing arrow",
			wantErr: true,
			tokens: []Token{
				tok(TokIdent, 0, "Read"),
				tok(TokIdent, 1, "OUT"),
				tok(TokIdent, 2, "IN"),
				tok(TokIdent, 3, "Parse"),
			},
		},
		{
			// Read OUT -> IN  — EOF after target port, no target process
			name:    "missing target process",
			wantErr: true,
			tokens: []Token{
				tok(TokIdent, 0, "Read"),
				tok(TokIdent, 1, "OUT"),
				tok(TokArrow, 2, "->"),
				tok(TokIdent, 3, "IN"),
			},
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			stmt := newStatement(tc.tokens)
			got, err := parseConnectionStatement(stmt)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertFragments(t, got, tc.want)
		})
	}
}

// TestParseConnection is a smoke test for the ParseConnection component.
// It wires the component manually and verifies that one statement produces
// the correct fragment stream.
func TestParseConnection(t *testing.T) {
	file := &File{Name: "smoke.fbp", Data: []byte("Read(dsl/Reader) OUT -> IN Parse(dsl/Parser)")}

	tok := func(tt TokenType, i int, v string) Token {
		return testToken(file, tt, i, 1, i+1, i+1, v)
	}

	tokens := []Token{
		tok(TokIdent, 0, "Read"),
		tok(TokLParen, 1, "("),
		tok(TokIdent, 2, "dsl"),
		tok(TokSlash, 3, "/"),
		tok(TokIdent, 4, "Reader"),
		tok(TokRParen, 5, ")"),
		tok(TokIdent, 6, "OUT"),
		tok(TokArrow, 7, "->"),
		tok(TokIdent, 8, "IN"),
		tok(TokIdent, 9, "Parse"),
		tok(TokLParen, 10, "("),
		tok(TokIdent, 11, "dsl"),
		tok(TokSlash, 12, "/"),
		tok(TokIdent, 13, "Parser"),
		tok(TokRParen, 14, ")"),
	}

	stmt := newStatement(tokens)

	pc := &ParseConnection{}
	in := make(chan Statement, 1)
	out := make(chan Fragment, 5)
	pc.In = in
	pc.Out = out

	wait := goflow.Run(pc)
	in <- stmt
	close(in)
	<-wait

	want := []Fragment{
		{Kind: FragmentProcess, Process: &ProcessDef{Name: "Read", Component: "dsl/Reader"}},
		{Kind: FragmentProcess, Process: &ProcessDef{Name: "Parse", Component: "dsl/Parser"}},
		{Kind: FragmentConnection, Connection: &ConnectionDef{
			Src: Endpoint{Process: "Read", Port: "OUT"},
			Tgt: Endpoint{Process: "Parse", Port: "IN"},
		}},
	}

	var got []Fragment
drain:
	for {
		select {
		case f := <-out:
			got = append(got, f)
		default:
			break drain
		}
	}

	assertFragments(t, got, want)
}
