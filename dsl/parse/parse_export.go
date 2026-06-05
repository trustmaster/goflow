package parse

import (
	"fmt"

	"github.com/trustmaster/goflow/dsl/types"
)

// ParseExport is a FBP component that parses INPORT/OUTPORT export statements.
type ParseExport struct {
	In  <-chan types.Statement
	Out chan<- types.Fragment
}

// Process parses each incoming Statement as an export declaration.
func (p *ParseExport) Process() {
	for stmt := range p.In {
		def, err := parseExportStatement(stmt)
		if err != nil {
			p.Out <- types.Fragment{Kind: types.FragmentError, Err: err}
			continue
		}

		p.Out <- types.Fragment{Kind: types.FragmentExport, Export: &def}
	}
}

// parseExportStatement parses a single export statement.
//
// Expected token sequence:
//
//	INPORT  = ProcName . PortName : PublicName
//	OUTPORT = ProcName . PortName : PublicName
func parseExportStatement(stmt types.Statement) (types.ExportDef, *types.ParseError) {
	cur := newTokenCursor(stmt.Tokens)

	keyword, err := cur.expectIdent()
	if err != nil {
		return types.ExportDef{}, &types.ParseError{Span: stmt.Span, Err: fmt.Errorf("export statement must start with INPORT or OUTPORT")}
	}

	var kind types.ExportKind

	switch keyword.Value {
	case "INPORT":
		kind = types.ExportInPort
	case "OUTPORT":
		kind = types.ExportOutPort
	default:
		return types.ExportDef{}, &types.ParseError{Span: keyword.Span, Err: fmt.Errorf("expected INPORT or OUTPORT, got %q", keyword.Value)}
	}

	if _, err := cur.expect(types.TokEqual); err != nil {
		return types.ExportDef{}, err
	}

	proc, err := cur.expectIdent()
	if err != nil {
		return types.ExportDef{}, err
	}

	if _, err := cur.expect(types.TokDot); err != nil {
		return types.ExportDef{}, err
	}

	port, err := cur.expectIdent()
	if err != nil {
		return types.ExportDef{}, err
	}

	if _, err := cur.expect(types.TokColon); err != nil {
		return types.ExportDef{}, err
	}

	public, err := cur.expectIdent()
	if err != nil {
		return types.ExportDef{}, err
	}

	return types.ExportDef{
		Kind:   kind,
		Public: public.Value,
		Proc:   proc.Value,
		Port:   port.Value,
	}, nil
}
