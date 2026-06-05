package parse

import (
	"testing"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/types"
)

func TestParseExportStatement(t *testing.T) {
	file := &types.File{Name: "test.fbp", Data: []byte("INPORT = Reader.FILE:FILE")}

	cases := []struct {
		name    string
		tokens  []types.Token
		want    types.ExportDef
		wantErr bool
	}{
		{
			name: "valid INPORT",
			tokens: []types.Token{
				testToken(file, types.TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, types.TokEqual, 7, 1, 8, 8, "="),
				testToken(file, types.TokIdent, 9, 1, 10, 15, "Reader"),
				testToken(file, types.TokDot, 15, 1, 16, 16, "."),
				testToken(file, types.TokIdent, 16, 1, 17, 20, "FILE"),
				testToken(file, types.TokColon, 21, 1, 22, 22, ":"),
				testToken(file, types.TokIdent, 23, 1, 24, 27, "FILE"),
			},
			want: types.ExportDef{Kind: types.ExportInPort, Public: "FILE", Proc: "Reader", Port: "FILE"},
		},
		{
			name: "valid OUTPORT",
			tokens: []types.Token{
				testToken(file, types.TokIdent, 0, 1, 1, 7, "OUTPORT"),
				testToken(file, types.TokEqual, 8, 1, 9, 9, "="),
				testToken(file, types.TokIdent, 10, 1, 11, 16, "Parser"),
				testToken(file, types.TokDot, 16, 1, 17, 17, "."),
				testToken(file, types.TokIdent, 17, 1, 18, 20, "OUT"),
				testToken(file, types.TokColon, 21, 1, 22, 22, ":"),
				testToken(file, types.TokIdent, 23, 1, 24, 27, "TREE"),
			},
			want: types.ExportDef{Kind: types.ExportOutPort, Public: "TREE", Proc: "Parser", Port: "OUT"},
		},
		{
			name: "wrong keyword",
			tokens: []types.Token{
				testToken(file, types.TokIdent, 0, 1, 1, 4, "OOPS"),
			},
			wantErr: true,
		},
		{
			name: "missing equals",
			tokens: []types.Token{
				testToken(file, types.TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, types.TokIdent, 7, 1, 8, 13, "Reader"),
			},
			wantErr: true,
		},
		{
			name: "missing colon",
			tokens: []types.Token{
				testToken(file, types.TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, types.TokEqual, 7, 1, 8, 8, "="),
				testToken(file, types.TokIdent, 9, 1, 10, 13, "Proc"),
				testToken(file, types.TokDot, 13, 1, 14, 14, "."),
				testToken(file, types.TokIdent, 14, 1, 15, 18, "Port"),
			},
			wantErr: true,
		},
		{
			name: "missing public name",
			tokens: []types.Token{
				testToken(file, types.TokIdent, 0, 1, 1, 6, "INPORT"),
				testToken(file, types.TokEqual, 7, 1, 8, 8, "="),
				testToken(file, types.TokIdent, 9, 1, 10, 13, "Proc"),
				testToken(file, types.TokDot, 13, 1, 14, 14, "."),
				testToken(file, types.TokIdent, 14, 1, 15, 18, "Port"),
				testToken(file, types.TokColon, 19, 1, 20, 20, ":"),
			},
			wantErr: true,
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			stmt := types.NewStatement(tc.tokens)
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
	file := &types.File{Name: "test.fbp", Data: []byte("INPORT = Reader.FILE:FILE\nOUTPORT = Parser.OUT:TREE")}

	stmts := []types.Statement{
		types.NewStatement([]types.Token{
			testToken(file, types.TokIdent, 0, 1, 1, 6, "INPORT"),
			testToken(file, types.TokEqual, 7, 1, 8, 8, "="),
			testToken(file, types.TokIdent, 9, 1, 10, 15, "Reader"),
			testToken(file, types.TokDot, 15, 1, 16, 16, "."),
			testToken(file, types.TokIdent, 16, 1, 17, 20, "FILE"),
			testToken(file, types.TokColon, 21, 1, 22, 22, ":"),
			testToken(file, types.TokIdent, 23, 1, 24, 27, "FILE"),
		}),
		types.NewStatement([]types.Token{
			testToken(file, types.TokIdent, 26, 2, 1, 33, "OUTPORT"),
			testToken(file, types.TokEqual, 34, 2, 9, 35, "="),
			testToken(file, types.TokIdent, 36, 2, 11, 42, "Parser"),
			testToken(file, types.TokDot, 42, 2, 17, 43, "."),
			testToken(file, types.TokIdent, 43, 2, 18, 46, "OUT"),
			testToken(file, types.TokColon, 47, 2, 22, 48, ":"),
			testToken(file, types.TokIdent, 49, 2, 24, 53, "TREE"),
		}),
	}

	want := []types.ExportDef{
		{Kind: types.ExportInPort, Public: "FILE", Proc: "Reader", Port: "FILE"},
		{Kind: types.ExportOutPort, Public: "TREE", Proc: "Parser", Port: "OUT"},
	}

	component := &ParseExport{}
	in := make(chan types.Statement)
	out := make(chan types.Fragment, len(stmts))
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
			if frag.Kind != types.FragmentExport {
				t.Fatalf("fragment %d: expected kind %q, got %q", i, types.FragmentExport, frag.Kind)
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
