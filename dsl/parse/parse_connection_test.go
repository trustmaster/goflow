package parse

import (
	"fmt"
	"testing"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/types"
)

// intPtr returns a pointer to n, used to express array port index expectations.
func intPtr(n int) *int { return &n }

// assertEndpoint checks that two Endpoint values match.
func assertEndpoint(t *testing.T, label string, got, want types.Endpoint) {
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
func assertFragments(t *testing.T, got, want []types.Fragment) {
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
		case types.FragmentProcess:
			if got[i].Process == nil {
				t.Errorf("fragment[%d]: Process is nil", i)
				continue
			}

			if *got[i].Process != *want[i].Process {
				t.Errorf("fragment[%d] Process: want %+v, got %+v", i, *want[i].Process, *got[i].Process)
			}

		case types.FragmentConnection:
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
	file := &types.File{Name: "conn_test.fbp", Data: []byte("parse connection statement test data for spans")}

	// tok is a shorthand: each token is assigned position i with span [i, i+1).
	tok := func(tt types.TokenType, i int, v string) types.Token {
		return testToken(file, tt, i, 1, i+1, i+1, v)
	}

	cases := []struct {
		name    string
		tokens  []types.Token
		want    []types.Fragment
		wantErr bool
	}{
		{
			// Read(dsl/Reader) OUT -> IN Parse(dsl/Parser)
			name: "full with components",
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Read"),
				tok(types.TokLParen, 1, "("),
				tok(types.TokIdent, 2, "dsl"),
				tok(types.TokSlash, 3, "/"),
				tok(types.TokIdent, 4, "Reader"),
				tok(types.TokRParen, 5, ")"),
				tok(types.TokIdent, 6, "OUT"),
				tok(types.TokArrow, 7, "->"),
				tok(types.TokIdent, 8, "IN"),
				tok(types.TokIdent, 9, "Parse"),
				tok(types.TokLParen, 10, "("),
				tok(types.TokIdent, 11, "dsl"),
				tok(types.TokSlash, 12, "/"),
				tok(types.TokIdent, 13, "Parser"),
				tok(types.TokRParen, 14, ")"),
			},
			want: []types.Fragment{
				{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Read", Component: "dsl/Reader"}},
				{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Parse", Component: "dsl/Parser"}},
				{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
					Src: types.Endpoint{Process: "Read", Port: "OUT"},
					Tgt: types.Endpoint{Process: "Parse", Port: "IN"},
				}},
			},
		},
		{
			// Reader OUT -> IN Parser  (no inline components — no process fragments)
			name: "no components",
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Reader"),
				tok(types.TokIdent, 1, "OUT"),
				tok(types.TokArrow, 2, "->"),
				tok(types.TokIdent, 3, "IN"),
				tok(types.TokIdent, 4, "Parser"),
			},
			want: []types.Fragment{
				{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
					Src: types.Endpoint{Process: "Reader", Port: "OUT"},
					Tgt: types.Endpoint{Process: "Parser", Port: "IN"},
				}},
			},
		},
		{
			// Split OUT[0] -> IN Foo
			name: "array port source",
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Split"),
				tok(types.TokIdent, 1, "OUT"),
				tok(types.TokLBracket, 2, "["),
				tok(types.TokInt, 3, "0"),
				tok(types.TokRBracket, 4, "]"),
				tok(types.TokArrow, 5, "->"),
				tok(types.TokIdent, 6, "IN"),
				tok(types.TokIdent, 7, "Foo"),
			},
			want: []types.Fragment{
				{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
					Src: types.Endpoint{Process: "Split", Port: "OUT", Index: intPtr(0)},
					Tgt: types.Endpoint{Process: "Foo", Port: "IN"},
				}},
			},
		},
		{
			// Source OUT -> IN[2] Dest
			name: "array port target",
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Source"),
				tok(types.TokIdent, 1, "OUT"),
				tok(types.TokArrow, 2, "->"),
				tok(types.TokIdent, 3, "IN"),
				tok(types.TokLBracket, 4, "["),
				tok(types.TokInt, 5, "2"),
				tok(types.TokRBracket, 6, "]"),
				tok(types.TokIdent, 7, "Dest"),
			},
			want: []types.Fragment{
				{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
					Src: types.Endpoint{Process: "Source", Port: "OUT"},
					Tgt: types.Endpoint{Process: "Dest", Port: "IN", Index: intPtr(2)},
				}},
			},
		},
		{
			// Src(a/b/c) PORT -> IN Tgt  — three-segment component path
			name: "multi-part component",
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Src"),
				tok(types.TokLParen, 1, "("),
				tok(types.TokIdent, 2, "a"),
				tok(types.TokSlash, 3, "/"),
				tok(types.TokIdent, 4, "b"),
				tok(types.TokSlash, 5, "/"),
				tok(types.TokIdent, 6, "c"),
				tok(types.TokRParen, 7, ")"),
				tok(types.TokIdent, 8, "PORT"),
				tok(types.TokArrow, 9, "->"),
				tok(types.TokIdent, 10, "IN"),
				tok(types.TokIdent, 11, "Tgt"),
			},
			want: []types.Fragment{
				{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Src", Component: "a/b/c"}},
				{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
					Src: types.Endpoint{Process: "Src", Port: "PORT"},
					Tgt: types.Endpoint{Process: "Tgt", Port: "IN"},
				}},
			},
		},
		{
			// Read(dsl/Reader) ->  — arrow where source port expected
			name:    "missing source port",
			wantErr: true,
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Read"),
				tok(types.TokLParen, 1, "("),
				tok(types.TokIdent, 2, "dsl"),
				tok(types.TokSlash, 3, "/"),
				tok(types.TokIdent, 4, "Reader"),
				tok(types.TokRParen, 5, ")"),
				tok(types.TokArrow, 6, "->"),
			},
		},
		{
			// Read OUT IN Parse  — IN where arrow expected
			name:    "missing arrow",
			wantErr: true,
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Read"),
				tok(types.TokIdent, 1, "OUT"),
				tok(types.TokIdent, 2, "IN"),
				tok(types.TokIdent, 3, "Parse"),
			},
		},
		{
			// Read OUT -> IN  — EOF after target port, no target process
			name:    "missing target process",
			wantErr: true,
			tokens: []types.Token{
				tok(types.TokIdent, 0, "Read"),
				tok(types.TokIdent, 1, "OUT"),
				tok(types.TokArrow, 2, "->"),
				tok(types.TokIdent, 3, "IN"),
			},
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			stmt := types.NewStatement(tc.tokens)
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
	file := &types.File{Name: "smoke.fbp", Data: []byte("Read(dsl/Reader) OUT -> IN Parse(dsl/Parser)")}

	tok := func(tt types.TokenType, i int, v string) types.Token {
		return testToken(file, tt, i, 1, i+1, i+1, v)
	}

	tokens := []types.Token{
		tok(types.TokIdent, 0, "Read"),
		tok(types.TokLParen, 1, "("),
		tok(types.TokIdent, 2, "dsl"),
		tok(types.TokSlash, 3, "/"),
		tok(types.TokIdent, 4, "Reader"),
		tok(types.TokRParen, 5, ")"),
		tok(types.TokIdent, 6, "OUT"),
		tok(types.TokArrow, 7, "->"),
		tok(types.TokIdent, 8, "IN"),
		tok(types.TokIdent, 9, "Parse"),
		tok(types.TokLParen, 10, "("),
		tok(types.TokIdent, 11, "dsl"),
		tok(types.TokSlash, 12, "/"),
		tok(types.TokIdent, 13, "Parser"),
		tok(types.TokRParen, 14, ")"),
	}

	stmt := types.NewStatement(tokens)

	pc := &ParseConnection{}
	in := make(chan types.Statement, 1)
	out := make(chan types.Fragment, 5)
	pc.In = in
	pc.Out = out

	wait := goflow.Run(pc)
	in <- stmt
	close(in)
	<-wait

	want := []types.Fragment{
		{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Read", Component: "dsl/Reader"}},
		{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Parse", Component: "dsl/Parser"}},
		{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
			Src: types.Endpoint{Process: "Read", Port: "OUT"},
			Tgt: types.Endpoint{Process: "Parse", Port: "IN"},
		}},
	}

	var got []types.Fragment
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
