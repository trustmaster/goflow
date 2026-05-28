package dsl

import (
	"testing"

	"github.com/trustmaster/goflow"
)

func TestParseExportStatement(t *testing.T) {
	file := &File{Name: "test.fbp", Data: []byte("INPORT = Reader.FILE:FILE")}

	cases := []struct {
		name    string
		tokens  []Token
		want    ExportDef
		wantErr bool
	}{
		{
			name: "valid INPORT",
			tokens: []Token{
				testToken(file, TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, TokEqual, 7, 1, 8, 8, "="),
				testToken(file, TokIdent, 9, 1, 10, 15, "Reader"),
				testToken(file, TokDot, 15, 1, 16, 16, "."),
				testToken(file, TokIdent, 16, 1, 17, 20, "FILE"),
				testToken(file, TokColon, 21, 1, 22, 22, ":"),
				testToken(file, TokIdent, 23, 1, 24, 27, "FILE"),
			},
			want: ExportDef{Kind: ExportInPort, Public: "FILE", Proc: "Reader", Port: "FILE"},
		},
		{
			name: "valid OUTPORT",
			tokens: []Token{
				testToken(file, TokIdent, 0, 1, 1, 7, "OUTPORT"),
				testToken(file, TokEqual, 8, 1, 9, 9, "="),
				testToken(file, TokIdent, 10, 1, 11, 16, "Parser"),
				testToken(file, TokDot, 16, 1, 17, 17, "."),
				testToken(file, TokIdent, 17, 1, 18, 20, "OUT"),
				testToken(file, TokColon, 21, 1, 22, 22, ":"),
				testToken(file, TokIdent, 23, 1, 24, 27, "TREE"),
			},
			want: ExportDef{Kind: ExportOutPort, Public: "TREE", Proc: "Parser", Port: "OUT"},
		},
		{
			name: "wrong keyword",
			tokens: []Token{
				testToken(file, TokIdent, 0, 1, 1, 4, "OOPS"),
			},
			wantErr: true,
		},
		{
			name: "missing equals",
			tokens: []Token{
				testToken(file, TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, TokIdent, 7, 1, 8, 13, "Reader"),
			},
			wantErr: true,
		},
		{
			name: "missing colon",
			tokens: []Token{
				testToken(file, TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, TokEqual, 7, 1, 8, 8, "="),
				testToken(file, TokIdent, 9, 1, 10, 13, "Proc"),
				testToken(file, TokDot, 13, 1, 14, 14, "."),
				testToken(file, TokIdent, 14, 1, 15, 18, "Port"),
			},
			wantErr: true,
		},
		{
			name: "missing public name",
			tokens: []Token{
				testToken(file, TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, TokEqual, 7, 1, 8, 8, "="),
				testToken(file, TokIdent, 9, 1, 10, 13, "Proc"),
				testToken(file, TokDot, 13, 1, 14, 14, "."),
				testToken(file, TokIdent, 14, 1, 15, 18, "Port"),
				testToken(file, TokColon, 19, 1, 20, 20, ":"),
			},
			wantErr: true,
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			stmt := newStatement(tc.tokens)
			got, err := parseExportStatement(stmt)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %#v, got %#v", tc.want, got)
			}
		})
	}
}

func TestParseExport(t *testing.T) {
	file := &File{Name: "test.fbp", Data: []byte("INPORT = Reader.FILE:FILE\nOUTPORT = Parser.OUT:TREE")}

	stmts := []Statement{
		newStatement([]Token{
			testToken(file, TokIdent, 0, 1, 1, 6, "INPORT"),
			testToken(file, TokEqual, 7, 1, 8, 8, "="),
			testToken(file, TokIdent, 9, 1, 10, 15, "Reader"),
			testToken(file, TokDot, 15, 1, 16, 16, "."),
			testToken(file, TokIdent, 16, 1, 17, 20, "FILE"),
			testToken(file, TokColon, 21, 1, 22, 22, ":"),
			testToken(file, TokIdent, 23, 1, 24, 27, "FILE"),
		}),
		newStatement([]Token{
			testToken(file, TokIdent, 26, 2, 1, 33, "OUTPORT"),
			testToken(file, TokEqual, 34, 2, 9, 35, "="),
			testToken(file, TokIdent, 36, 2, 11, 42, "Parser"),
			testToken(file, TokDot, 42, 2, 17, 43, "."),
			testToken(file, TokIdent, 43, 2, 18, 46, "OUT"),
			testToken(file, TokColon, 47, 2, 22, 48, ":"),
			testToken(file, TokIdent, 49, 2, 24, 53, "TREE"),
		}),
	}

	want := []ExportDef{
		{Kind: ExportInPort, Public: "FILE", Proc: "Reader", Port: "FILE"},
		{Kind: ExportOutPort, Public: "TREE", Proc: "Parser", Port: "OUT"},
	}

	component := &ParseExport{}
	in := make(chan Statement)
	out := make(chan Fragment, len(stmts))
	component.In = in
	component.Out = out

	wait := goflow.Run(component)
	go func() {
		for _, s := range stmts {
			in <- s
		}
		close(in)
	}()

	<-wait

	for i, wantDef := range want {
		select {
		case frag, ok := <-out:
			if !ok {
				t.Fatalf("output channel closed early, expected %d fragments", len(want))
			}
			if frag.Kind != FragmentExport {
				t.Fatalf("fragment %d: expected kind %q, got %q", i, FragmentExport, frag.Kind)
			}
			if frag.Export == nil {
				t.Fatalf("fragment %d: Export is nil", i)
			}
			if *frag.Export != wantDef {
				t.Fatalf("fragment %d: expected %#v, got %#v", i, wantDef, *frag.Export)
			}
		default:
			t.Fatalf("expected fragment %d, got nothing", i)
		}
	}
}
