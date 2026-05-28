package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestParseIIPStatement(t *testing.T) {
	file := &File{Name: "test.fbp", Data: []byte("'hello' -> IN Greeter")}

	cases := []struct {
		name      string
		tokens    []Token
		wantFrags int
		wantErr   bool
		check     func(t *testing.T, frags []Fragment)
	}{
		{
			name: "quoted IIP",
			tokens: []Token{
				testToken(file, TokQuoted, 0, 1, 1, 7, "'hello'"),
				testToken(file, TokArrow, 8, 1, 9, 10, "->"),
				testToken(file, TokIdent, 11, 1, 12, 13, "IN"),
				testToken(file, TokIdent, 14, 1, 15, 21, "Greeter"),
			},
			wantFrags: 1,
			check: func(t *testing.T, frags []Fragment) {
				t.Helper()
				frag := frags[0]
				if frag.Kind != FragmentIIP {
					t.Fatalf("expected kind %q, got %q", FragmentIIP, frag.Kind)
				}
				if frag.IIP == nil {
					t.Fatal("IIP is nil")
				}
				if frag.IIP.Data != "hello" {
					t.Fatalf("expected data %q, got %v", "hello", frag.IIP.Data)
				}
				if frag.IIP.Tgt.Process != "Greeter" {
					t.Fatalf("expected process %q, got %q", "Greeter", frag.IIP.Tgt.Process)
				}
				if frag.IIP.Tgt.Port != "IN" {
					t.Fatalf("expected port %q, got %q", "IN", frag.IIP.Tgt.Port)
				}
			},
		},
		{
			name: "integer IIP",
			tokens: []Token{
				testToken(file, TokInt, 0, 1, 1, 2, "42"),
				testToken(file, TokArrow, 3, 1, 4, 5, "->"),
				testToken(file, TokIdent, 6, 1, 7, 8, "IN"),
				testToken(file, TokIdent, 9, 1, 10, 13, "Proc"),
			},
			wantFrags: 1,
			check: func(t *testing.T, frags []Fragment) {
				t.Helper()
				frag := frags[0]
				if frag.Kind != FragmentIIP {
					t.Fatalf("expected kind %q, got %q", FragmentIIP, frag.Kind)
				}
				if frag.IIP == nil {
					t.Fatal("IIP is nil")
				}
				if frag.IIP.Data != 42 {
					t.Fatalf("expected data 42, got %v", frag.IIP.Data)
				}
				if frag.IIP.Tgt.Process != "Proc" {
					t.Fatalf("expected process %q, got %q", "Proc", frag.IIP.Tgt.Process)
				}
				if frag.IIP.Tgt.Port != "IN" {
					t.Fatalf("expected port %q, got %q", "IN", frag.IIP.Tgt.Port)
				}
			},
		},
		{
			name: "IIP with component",
			tokens: []Token{
				testToken(file, TokQuoted, 0, 1, 1, 5, "'msg'"),
				testToken(file, TokArrow, 6, 1, 7, 8, "->"),
				testToken(file, TokIdent, 9, 1, 10, 11, "IN"),
				testToken(file, TokIdent, 12, 1, 13, 16, "Sink"),
				testToken(file, TokLParen, 16, 1, 17, 17, "("),
				testToken(file, TokIdent, 17, 1, 18, 20, "pkg"),
				testToken(file, TokSlash, 20, 1, 21, 21, "/"),
				testToken(file, TokIdent, 21, 1, 22, 25, "Sink"),
				testToken(file, TokRParen, 25, 1, 26, 26, ")"),
			},
			wantFrags: 2,
			check: func(t *testing.T, frags []Fragment) {
				t.Helper()
				if frags[0].Kind != FragmentProcess {
					t.Fatalf("fragment 0: expected kind %q, got %q", FragmentProcess, frags[0].Kind)
				}
				if frags[0].Process == nil {
					t.Fatal("fragment 0: Process is nil")
				}
				if frags[0].Process.Name != "Sink" {
					t.Fatalf("fragment 0: expected process name %q, got %q", "Sink", frags[0].Process.Name)
				}
				if frags[0].Process.Component != "pkg/Sink" {
					t.Fatalf("fragment 0: expected component %q, got %q", "pkg/Sink", frags[0].Process.Component)
				}
				if frags[1].Kind != FragmentIIP {
					t.Fatalf("fragment 1: expected kind %q, got %q", FragmentIIP, frags[1].Kind)
				}
				if frags[1].IIP == nil {
					t.Fatal("fragment 1: IIP is nil")
				}
				if frags[1].IIP.Data != "msg" {
					t.Fatalf("fragment 1: expected data %q, got %v", "msg", frags[1].IIP.Data)
				}
			},
		},
		{
			name: "IIP with array index",
			tokens: []Token{
				testToken(file, TokQuoted, 0, 1, 1, 3, "'x'"),
				testToken(file, TokArrow, 4, 1, 5, 6, "->"),
				testToken(file, TokIdent, 7, 1, 8, 9, "IN"),
				testToken(file, TokLBracket, 9, 1, 10, 10, "["),
				testToken(file, TokInt, 10, 1, 11, 11, "1"),
				testToken(file, TokRBracket, 11, 1, 12, 12, "]"),
				testToken(file, TokIdent, 13, 1, 14, 19, "Target"),
			},
			wantFrags: 1,
			check: func(t *testing.T, frags []Fragment) {
				t.Helper()
				frag := frags[0]
				if frag.Kind != FragmentIIP {
					t.Fatalf("expected kind %q, got %q", FragmentIIP, frag.Kind)
				}
				if frag.IIP == nil {
					t.Fatal("IIP is nil")
				}
				if frag.IIP.Tgt.Index == nil {
					t.Fatal("Tgt.Index is nil, expected pointer to 1")
				}
				if *frag.IIP.Tgt.Index != 1 {
					t.Fatalf("expected Tgt.Index == 1, got %d", *frag.IIP.Tgt.Index)
				}
			},
		},
		{
			name: "missing arrow",
			tokens: []Token{
				testToken(file, TokQuoted, 0, 1, 1, 6, "'data'"),
				testToken(file, TokIdent, 7, 1, 8, 9, "IN"),
				testToken(file, TokIdent, 10, 1, 11, 14, "Proc"),
			},
			wantErr: true,
		},
		{
			name: "missing process",
			tokens: []Token{
				testToken(file, TokQuoted, 0, 1, 1, 6, "'data'"),
				testToken(file, TokArrow, 7, 1, 8, 9, "->"),
				testToken(file, TokIdent, 10, 1, 11, 12, "IN"),
			},
			wantErr: true,
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			stmt := newStatement(tc.tokens)
			frags, err := parseIIPStatement(stmt)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(frags) != tc.wantFrags {
				t.Fatalf("expected %d fragments, got %d", tc.wantFrags, len(frags))
			}
			if tc.check != nil {
				tc.check(t, frags)
			}
		})
	}
}

func TestParseIIP(t *testing.T) {
	file := &File{Name: "test.fbp", Data: []byte("'hello' -> IN Greeter")}

	stmt := newStatement([]Token{
		testToken(file, TokQuoted, 0, 1, 1, 7, "'hello'"),
		testToken(file, TokArrow, 8, 1, 9, 10, "->"),
		testToken(file, TokIdent, 11, 1, 12, 13, "IN"),
		testToken(file, TokIdent, 14, 1, 15, 21, "Greeter"),
	})

	component := &ParseIIP{}
	in := make(chan Statement)
	out := make(chan Fragment, 2)
	component.In = in
	component.Out = out

	wait := goflow.Run(component)
	go func() {
		in <- stmt
		close(in)
	}()

	<-wait

	select {
	case frag, ok := <-out:
		if !ok {
			t.Fatal("output channel closed, expected a fragment")
		}
		if frag.Kind != FragmentIIP {
			t.Fatalf("expected kind %q, got %q", FragmentIIP, frag.Kind)
		}
		if frag.IIP == nil {
			t.Fatal("IIP is nil")
		}
		if frag.IIP.Data != "hello" {
			t.Fatalf("expected data %q, got %v", "hello", frag.IIP.Data)
		}
		if frag.IIP.Tgt.Process != "Greeter" {
			t.Fatalf("expected process %q, got %q", "Greeter", frag.IIP.Tgt.Process)
		}
		if frag.IIP.Tgt.Port != "IN" {
			t.Fatalf("expected port %q, got %q", "IN", frag.IIP.Tgt.Port)
		}
	default:
		t.Fatal("expected a fragment, got nothing")
	}
}
